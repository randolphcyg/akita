package model

import (
	"fmt"

	"gorm.io/gorm"
)

// LdapConn LDAP服务器连接配置
type LdapConn struct {
	gorm.Model
	// 连接地址
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index;not null;comment:连接地址 逻辑外键"`
	// SSL加密方式
	SslEncryption bool `json:"ssl_encryption" gorm:"type:tinyint;length:1;comment:SSL加密方式"`
	// 超时设置
	Timeout int `json:"timeout" gorm:"type:int;comment:超时设置"`
	// 根目录
	BaseDn string `json:"base_dn" gorm:"type:varchar(255);not null;comment:根目录"`
	// 用户名
	AdminAccount string `json:"admin_account" gorm:"type:varchar(255);not null;comment:用户名"`
	// 密码
	Password string `json:"password" gorm:"type:varchar(255);not null;comment:密码"`
}

// 公司名-英文前缀映射关系
type CompPreMap struct {
	Company string `json:"company"`
	Prefix  string `json:"prefix"`
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
	// 公司英文前缀
	CompanyPrefixs      string       `json:"company_prefixs" gorm:"type:varchar(255);comment:公司英文前缀"`
	CompanyPrefixsSlice []CompPreMap `gorm:"-"` // 非数据库字段 用来处理复杂数据结构

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

// GetAllLdapConn 查询所有ldap连接
func GetAllLdapConn() (LdapConn, error) {
	var conn LdapConn
	result := DB.Find(&conn)
	fmt.Println(result.RowsAffected)
	return conn, result.Error
}

// GetLdapConn 查询一个ldap连接
func GetLdapConn(id uint) (LdapConn, error) {
	var conn LdapConn
	result := DB.First(&conn, id)
	return conn, result.Error
}

// GetLdapConnByConnUrl  根据conn_url查询一个ldap连接
func GetLdapConnByConnUrl(url string) (LdapConn, error) {
	var conn LdapConn
	result := DB.Where("conn_url = ?", url).First(&conn)
	return conn, result.Error
}

// NewLdapConn 返回一个新的空 LdapConn
func NewLdapConn() LdapConn {
	return LdapConn{}
}

// NewLdapField 返回一个新的空 LdapField
func NewLdapField() LdapField {
	return LdapField{}
}

// GetLdapFieldByConn 根据连接的URL查询ldap连接的字段明细
func GetLdapFieldByConnUrl(url string) (LdapField, error) {
	var field LdapField
	result := DB.Where("conn_url = ?", url).First(&field)
	fmt.Println(result.RowsAffected)
	return field, result.Error
}
