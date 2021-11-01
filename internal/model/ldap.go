package model

import (
	"crypto/tls"

	"time"

	ldappool "github.com/RandolphCYG/ldapPool"
	"github.com/go-ldap/ldap/v3"
	"gorm.io/gorm"

	"gitee.com/RandolphCYG/akita/internal/middleware/log"
)

var (
	LdapCfgs   LdapCfg
	LdapFields LdapField
	LdapPool   ldappool.Pool
)

// LdapCfg LDAP服务器连接配置
type LdapCfg struct {
	gorm.Model
	// 连接地址
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index;not null;comment:连接地址 逻辑外键"`
	// SSL加密方式
	SslEncryption bool `json:"ssl_encryption" gorm:"type:tinyint;length:1;comment:SSL加密方式"`
	// 超时设置
	Timeout time.Duration `json:"timeout" gorm:"type:int;comment:超时设置"`
	// 根目录
	BaseDn string `json:"base_dn" gorm:"type:varchar(255);not null;comment:根目录"`
	// 用户名
	AdminAccount string `json:"admin_account" gorm:"type:varchar(255);not null;comment:用户名"`
	// 密码
	Password string `json:"password" gorm:"type:varchar(255);not null;comment:密码"`
}

// CompanyType 公司类型
type CompanyType struct {
	IsOuter bool   `json:"is_outer"` // 是否外部公司
	Prefix  string `json:"prefix"`   // 用户名前缀 外部公司才有
}

// LdapField LDAP服务器字段配置
type LdapField struct {
	gorm.Model

	// 连接地址 逻辑外键
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index;not null;comment:连接地址 逻辑外键"`

	// 用户基础字段
	// 拉取结点 默认不填表示从根目录开始搜索
	BasicPullNode string `json:"basic_pull_node" gorm:"type:varchar(255);comment:拉取结点 默认不填表示从根目录开始搜索"`
	// 用户对象类
	UserClass string `json:"user_class" gorm:"type:varchar(50);not null;comment:用户对象类"`
	// 用户对象过滤
	UserFilter string `json:"user_filter" gorm:"type:varchar(255);not null;comment:用户对象过滤"`
	// 组织架构类
	OrganizationClass string `json:"organization_class" gorm:"type:varchar(50);not null;comment:组织架构类"`
	// 用户名
	Username string `json:"username" gorm:"type:varchar(50);not null;comment:用户名"`
	// 中文名
	DisplayName string `json:"display_name" gorm:"type:varchar(50);not null;comment:中文名"`
	// 邮箱
	Email string `json:"email" gorm:"type:varchar(50);not null;comment:邮箱"`
	// 手机号
	Mobile string `json:"mobile" gorm:"type:varchar(50);not null;comment:手机号"`

	// 用户拓展字段
	// 组织过滤规则
	SearchFilterOu string `json:"search_filter_ou" gorm:"type:varchar(255);comment:组织过滤规则"`
	// 禁用用户DN
	BaseDnDisabled string `json:"base_dn_disabled" gorm:"type:varchar(255);comment:禁用用户DN"`
	// 公司类型
	BaseDnOuter        string                 `json:"base_dn_outer" gorm:"type:varchar(255);comment:外部用户DN"`
	BaseDnToBeAssigned string                 `json:"base_dn_to_be_assigned" gorm:"type:varchar(255);comment:公司内部待分配DN"`
	CompanyType        string                 `json:"company_type" gorm:"type:varchar(3000);comment:公司类型"`
	CompanyTypes       map[string]CompanyType `json:"company_types" gorm:"-"` // 非数据库字段 用来处理复杂数据结构

	// 用户组字段
	// 用户组对象类
	UserGroupClass string `json:"user_group_class" gorm:"type:varchar(255);not null;comment:用户组对象类"`
	// 用户组对象过滤
	UserGroupFilter string `json:"user_group_filter" gorm:"type:varchar(255);not null;comment:用户组对象过滤"`
	// 用户组名
	UserGroupName string `json:"user_group_name" gorm:"type:varchar(255);not null;comment:用户组名"`
	// 用户组描述
	UserGroupDescription string `json:"user_group_description" gorm:"type:varchar(255);not null;comment:用户组描述"`
}

// Init 初始化连接池
func Init(c *LdapCfg) (err error) {
	LdapCfgs = LdapCfg{
		ConnUrl:       c.ConnUrl,
		SslEncryption: c.SslEncryption,
		Timeout:       c.Timeout,
		BaseDn:        c.BaseDn,
		AdminAccount:  c.AdminAccount,
		Password:      c.Password,
	}
	// 初始化ldap连接池
	LdapPool, err = ldappool.NewChannelPool(50, 1000, "originalLdapPool",
		func(s string) (ldap.Client, error) {
			conn, err := ldap.DialURL(LdapCfgs.ConnUrl)
			if err != nil {
				log.Log.Error("Fail to dial ldap url, err: ", err)
			}

			// 重新连接TLS
			if err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
				log.Log.Error("Fail to start tls, err: ", err)
			}

			// 与只读用户绑定
			if err = conn.Bind(LdapCfgs.AdminAccount, LdapCfgs.Password); err != nil {
				log.Log.Error("admin user auth failed, err: ", err)
			}
			return conn, nil
		}, []uint16{ldap.LDAPResultTimeLimitExceeded, ldap.ErrorNetwork})
	if err != nil {
		log.Log.Error(err)
		return
	}
	return
}

// GetAllLdapConn 查询所有ldap连接
func GetAllLdapConn() (LdapCfg, error) {
	var conn LdapCfg
	result := DB.Find(&conn)
	return conn, result.Error
}

// GetLdapConn 查询一个ldap连接
func GetLdapConn(id uint) (LdapCfg, error) {
	var conn LdapCfg
	result := DB.First(&conn, id)
	return conn, result.Error
}

// GetLdapConnByConnUrl  根据conn_url查询一个ldap连接
func GetLdapConnByConnUrl(url string) (LdapCfg, error) {
	var conn LdapCfg
	result := DB.Where("conn_url = ?", url).First(&conn)
	return conn, result.Error
}

// NewLdapConn 返回一个新的空 LdapConn
func NewLdapConn() LdapCfg {
	return LdapCfg{}
}

// NewLdapField 返回一个新的空 LdapField
func NewLdapField() LdapField {
	return LdapField{}
}

// GetLdapFieldByConnUrl 根据连接的URL查询ldap连接的字段明细
func GetLdapFieldByConnUrl(url string) (LdapField, error) {
	var field LdapField
	result := DB.Where("conn_url = ?", url).First(&field)
	return field, result.Error
}

/*
* LDAP用户部门变更历史
*
 */

// LdapUserDepartRecord LDAP用户部门变动记录
type LdapUserDepartRecord struct {
	gorm.Model
	Name      string `json:"name" gorm:"type:varchar(255);not null;comment:真实姓名"`
	Eid       string `json:"eid" gorm:"type:varchar(255);not null;comment:工号"`
	OldDepart string `json:"old_depart" gorm:"type:varchar(255);not null;comment:旧部门"`
	NewDepart string `json:"new_depart" gorm:"type:varchar(255);not null;comment:新部门"`
	Level     string `json:"level" gorm:"type:varchar(255);not null;comment:级别"`
}

// CreateLdapUserDepartRecord 用户架构变化记录
func CreateLdapUserDepartRecord(name string, eid string, oldDepart string, newDepart string, level string) {
	DB.Model(&LdapUserDepartRecord{}).Create(&LdapUserDepartRecord{Name: name, Eid: eid, OldDepart: oldDepart, NewDepart: newDepart, Level: level})
}

// FetchLdapUserDepartRecord 查询一段时间用户架构变化记录
func FetchLdapUserDepartRecord(offsetBefore, offsetAfter int) (ldapUserDepartRecords []LdapUserDepartRecord, err error) {
	begin, _ := time.Parse("2006-01-02", time.Now().AddDate(0, 0, offsetBefore).Format("2006-01-02")) // 开始日期的零点
	end, _ := time.Parse("2006-01-02", time.Now().AddDate(0, 0, offsetAfter).Format("2006-01-02"))    // 结束日期的最后一秒
	_ = DB.Where("created_at BETWEEN ? AND ?", begin, end).Find(&ldapUserDepartRecords)
	return
}
