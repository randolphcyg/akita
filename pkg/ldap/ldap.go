package ldap

import (
	"crypto/tls"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/log"

	"github.com/go-ldap/ldap"
)

// Init 实例化一个 ldapConn
func Init(c *model.LdapConn) *model.LdapConn {
	return &model.LdapConn{
		ConnUrl:       c.ConnUrl,
		SslEncryption: c.SslEncryption,
		Timeout:       c.Timeout,
		BaseDn:        c.BaseDn,
		AdminAccount:  c.AdminAccount,
		Password:      c.Password,
	}
}

// 获取ldap连接
func NewLdapConn(conn *model.LdapConn) (l *ldap.Conn, err error) {
	// 建立ldap连接
	l, err = ldap.DialURL(conn.ConnUrl)
	if err != nil {
		log.Log().Error("dial ldap url failed,err:%v", err)
		return
	}
	// defer l.Close()

	// 重新连接TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Log().Error("start tls failed,err:%v", err)
		return
	}

	// 首先与只读用户绑定
	err = l.Bind(conn.AdminAccount, conn.Password)
	if err != nil {
		log.Log().Error("admin user auth failed,err:%v", err)
		return
	}
	return
}

func FetchLdapUsers(conn *model.LdapConn) (LdapUsers []*ldap.Entry) {
	// 建立ldap连接
	ldap_conn, err := NewLdapConn(conn)
	if err != nil {
		log.Log().Error("setup ldap connect failed,err:%v\n", err)
	}
	defer ldap_conn.Close()

	attrs := []string{
		"employeeNumber",     // 工号
		"sAMAccountName",     // SAM账号
		"distinguishedName",  // dn
		"UserAccountControl", // 用户账户控制
		"accountExpires",     // 账户过期时间
		"pwdLastSet",         // 用户下次登录必须修改密码
		"whenCreated",        // 创建时间
		"whenChanged",        // 修改时间
		"displayName",        // 显示名
		"sn",                 // 姓
		"name",
		"givenName",       // 名
		"mail",            // 邮箱
		"mobile",          // 移动电话
		"telephoneNumber", // 电话号码
		"company",         // 公司
		"department",      // 部门
		"title",           // 职务
	}

	searchRequest := ldap.NewSearchRequest(
		conn.BaseDn, // 待查询的base dn
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=user)", // 过滤规则
		attrs,                // 待查询属性列表
		nil,
	)

	sr, err := ldap_conn.Search(searchRequest)
	if err != nil {
		log.Log().Error("查询用户出错:%v", err)
	}
	return sr.Entries
}

func AddLdapUsers(conn *model.LdapConn) (LdapUsers []*ldap.Entry) {
	// 建立ldap连接
	ldap_conn, err := NewLdapConn(conn)
	if err != nil {
		log.Log().Error("setup ldap connect failed,err:%v\n", err)
	}
	defer ldap_conn.Close()

	attrs := []string{
		"employeeNumber",     // 工号
		"sAMAccountName",     // SAM账号
		"distinguishedName",  // dn
		"UserAccountControl", // 用户账户控制
		"accountExpires",     // 账户过期时间
		"pwdLastSet",         // 用户下次登录必须修改密码
		"whenCreated",        // 创建时间
		"whenChanged",        // 修改时间
		"displayName",        // 显示名
		"sn",                 // 姓
		"name",
		"givenName",       // 名
		"mail",            // 邮箱
		"mobile",          // 移动电话
		"telephoneNumber", // 电话号码
		"company",         // 公司
		"department",      // 部门
		"title",           // 职务
	}
	// TODO 这里将写成ldap用户的创建通用逻辑
	searchRequest := ldap.NewSearchRequest(
		conn.BaseDn, // 待查询的base dn
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=user)", // 过滤规则
		attrs,                // 待查询属性列表
		nil,
	)

	sr, err := ldap_conn.Search(searchRequest)
	if err != nil {
		log.Log().Error("查询用户出错:%v", err)
	}
	return sr.Entries
}
