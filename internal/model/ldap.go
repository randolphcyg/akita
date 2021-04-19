package model

import (
	"fmt"

	"gorm.io/gorm"
)

// LdapConn LDAP服务器连接配置
type LdapConn struct {
	gorm.Model
	// 连接地址
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index"`
	// SSL加密方式
	SslEncryption bool `json:"ssl_encryption" gorm:"type:tinyint;length:1"`
	// 超时设置
	Timeout int `json:"timeout" gorm:"type:int"`
	// 根目录
	BaseDn string `json:"base_dn" gorm:"type:varchar(255)"`
	// 用户名
	AdminAccount string `json:"admin_account" gorm:"type:varchar(255)"`
	// 密码
	Password string `json:"password" gorm:"type:varchar(255)"`
}

// LdapFields LDAP服务器字段配置
type LdapFields struct {
	// 用户基础字段
	// 拉取结点 默认不填表示从根目录开始搜索
	BasicPullNode string
	// 用户对象类
	UserClass string
	// 用户对象过滤
	UserFilter string
	// 组织架构类
	OrganizationClass string
	// 用户名字段
	Username string
	// 中文名字段
	DisplayName string
	// 邮箱字段
	Email string
	// 手机号字段
	Mobile string

	// 用户拓展字段
	// 组织过滤规则
	SearchFilterOu string
	// 禁用用户DN
	BaseDnDisabled string

	// 用户组字段
	// 用户组对象类
	UserGroupClass string
	// 用户组对象过滤
	UserGroupFilter string
	// 用户组名
	UserGroupName string
	// 用户组描述
	UserGroupDescription string
}

// GetAllLdapConn 查询所有ldap连接
func GetAllLdapConn() (LdapConn, error) {
	var conn LdapConn
	result := DB.Find(&conn)
	fmt.Println(result.RowsAffected)
	return conn, result.Error
}

// NewLdapConn 返回一个新的空 LdapConn
func NewLdapConn() LdapConn {
	return LdapConn{}
}
