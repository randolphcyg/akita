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
func NewLdapConn(conn *model.LdapConn) (l *ldap.Conn) {
	// 建立ldap连接
	l, err := ldap.DialURL(conn.ConnUrl)
	if err != nil {
		log.Log().Error("建立ldap连接时突发错误:%v", err)
	}
	// defer l.Close()

	// 重新连接TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Log().Error("重新连接TLS突发错误:%v", err)
	}

	// 首先与只读用户绑定
	err = l.Bind(conn.AdminAccount, conn.Password)
	if err != nil {
		log.Log().Error("与只读用户绑定时突发错误:%v", err)
	}
	return
}

func FetchLdapUsers(l *ldap.Conn) (LdapUsers []*ldap.Entry) {
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

	var conn model.LdapConn

	searchRequest := ldap.NewSearchRequest(
		conn.BaseDn, // 待查询的base dn
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=user)", // 过滤规则
		attrs,                // 待查询属性列表
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Log().Error("查询用户出错:%v", err)
	}
	return sr.Entries
}

/*
* 查询所有的ldap用户
 */
// func FetchLdapUsers(conn model.ServerLdapConn) *ldap.SearchResult {
// 	// ldap连接
// 	l, _ := LdapConn(conn)
// 	// redis读取AD配置
// 	config := redis.LoadAdServerConfigFromRedis()
// 	conf := model.DB.Config
// 	fmt.Print(config)

// 	attrs := []string{
// 		"employeeNumber",     // 工号
// 		"sAMAccountName",     // SAM账号
// 		"distinguishedName",  // dn
// 		"UserAccountControl", // 用户账户控制
// 		"accountExpires",     // 账户过期时间
// 		"pwdLastSet",         // 用户下次登录必须修改密码
// 		"whenCreated",        // 创建时间
// 		"whenChanged",        // 修改时间
// 		"displayName",        // 显示名
// 		"sn",                 // 姓
// 		"name",
// 		"givenName",       // 名
// 		"mail",            // 邮箱
// 		"mobile",          // 移动电话
// 		"telephoneNumber", // 电话号码
// 		"company",         // 公司
// 		"department",      // 部门
// 		"title",           // 职务
// 	}

// 	searchRequest := ldap.NewSearchRequest(
// 		config["baseDnEnabled"], // 待查询的base dn
// 		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
// 		config["searchFilterUser"], // 过滤规则
// 		attrs,                      // 待查询属性列表
// 		nil,
// 	)

// 	sr, err := l.Search(searchRequest)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	return sr

// 	// user := make(map[string]string)
// 	// for _, v := range sr.Entries {
// 	// 	fmt.Print(reflect.TypeOf(v.GetAttributeValue("displayName")))
// 	// 	break
// 	// }

// 	// for _, v := range sr.Entries[0].Attributes {
// 	// 	user[v.Name] = v.Values[0]
// 	// 	if v.Name == "displayName" {
// 	// 		fmt.Print(fmt.Sprintf(v.Name) + " " + fmt.Sprintf(v.Values[0]) + "\n")
// 	// 	}
// 	// 	// fmt.Print(fmt.Sprintf(v.Name) + "\n")
// 	// 	// fmt.Print(fmt.Sprintf(v.Name) + " " + fmt.Sprintf(v.Values[0]) + "\n")
// 	// 	// fmt.Print(index, v.Values, entry.GetAttributeValues("whenCreated"), entry.GetAttributeValues("whenChanged"), entry.GetAttributeValues("displayName"), entry.GetAttributeValues("mail"), entry.GetAttributeValues("mobile"), entry.GetAttributeValues("title"))
// 	// }
// }
