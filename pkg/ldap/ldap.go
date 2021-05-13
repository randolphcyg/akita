package ldap

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"github.com/go-ldap/ldap"
	"golang.org/x/text/encoding/unicode"
)

// 禁用/启用用户的 UserAccountControl 状态码
var DisabledLdapUserCodes = [5]int32{514, 546, 66050, 66080, 66082}
var EnabledLdapUserCodes = [4]int32{512, 544, 66048, 262656}

type LdapAttributes struct {
	// ldap字段
	Num         string `json:"employeeNumber" gorm:"type:varchar(100);unique_index"`    // 工号
	Sam         string `json:"sAMAccountName" gorm:"type:varchar(128);unique_index"`    // SAM账号
	Dn          string `json:"distinguishedName" gorm:"type:varchar(100);unique_index"` // dn
	AccountCtl  string `json:"UserAccountControl" gorm:"type:varchar(100)"`             // 用户账户控制
	Expire      int64  `json:"accountExpires" gorm:"type:int(30)"`                      // 账户过期时间
	PwdLastSet  string `json:"pwdLastSet" gorm:"type:varchar(100)"`                     // 用户下次登录必须修改密码
	WhenCreated string `json:"whenCreated" gorm:"type:varchar(100)"`                    // 创建时间
	WhenChanged string `json:"whenChanged" gorm:"type:varchar(100)"`                    // 修改时间
	DisplayName string `json:"displayName" gorm:"type:varchar(32)"`                     // 真实姓名
	Sn          string `json:"sn" gorm:"type:varchar(100)"`                             // 姓
	Name        string `json:"name" gorm:"type:varchar(100)"`                           // 姓名
	GivenName   string `json:"givenName" gorm:"type:varchar(100)"`                      // 名
	Email       string `json:"mail" gorm:"type:varchar(128);unique_index"`              // 邮箱
	Phone       string `json:"mobile" gorm:"type:varchar(32);unique_index"`             // 移动电话
	Company     string `json:"company" gorm:"type:varchar(128)"`                        // 公司
	Depart      string `json:"department" gorm:"type:varchar(128)"`                     // 部门
	Title       string `json:"title" gorm:"type:varchar(100)"`                          // 职务
}

var attrs = []string{
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
	"givenName",  // 名
	"mail",       // 邮箱
	"mobile",     // 手机号
	"company",    // 公司
	"department", // 部门
	"title",      // 职务
}

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
	// 设置超时时间
	l.SetTimeout(time.Duration(conn.Timeout))
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

// 多条件查询用户 返回符合搜索条件的用户列表
func FetchLdapUsers(ldap_conn *ldap.Conn, conn *model.LdapConn, user *LdapAttributes) (result []*ldap.Entry) {
	// 多查询条件
	ldapFilterNum := "(employeeNumber=" + user.Num + ")"
	ldapFilterSam := "(sAMAccountName=" + user.Sam + ")"
	ldapFilterEmail := "(mail=" + user.Email + ")"
	ldapFilterPhone := "(mobile=" + user.Phone + ")"
	ldapFilterName := "(displayName=" + user.DisplayName + ")"
	ldapFilterDepart := "(department=" + user.Depart + ")"
	ldapFilterCompany := "(company=" + user.Company + ")"
	ldapFilterTitle := "(title=" + user.Title + ")"

	searchFilter := "(objectClass=organizationalPerson)"

	if user.Num != "" {
		searchFilter += ldapFilterNum
	}
	if user.Sam != "" {
		searchFilter += ldapFilterSam
	}
	if user.Email != "" {
		searchFilter += ldapFilterEmail
	}
	if user.Phone != "" {
		searchFilter += ldapFilterPhone
	}
	if user.DisplayName != "" {
		searchFilter += ldapFilterName
	}
	if user.Depart != "" {
		searchFilter += ldapFilterDepart
	}
	if user.Company != "" {
		searchFilter += ldapFilterCompany
	}
	if user.Title != "" {
		searchFilter += ldapFilterTitle
	}
	searchFilter = "(&" + searchFilter + ")"

	searchRequest := ldap.NewSearchRequest(
		conn.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	sr, err := ldap_conn.Search(searchRequest)
	if err != nil {
		log.Log().Error("%s", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries
	}
	return
}

// 用户过期期限处理 月份为-1 则过期时间为永久;否则 当前时间往后推迟expireMouths个月
func expireTime(expireMouths int64) (expireTimestamp int64) {
	expireTimestamp = 9223372036854775807
	if expireMouths != -1 {
		expireTimestamp = util.UnixToNt(time.Now().AddDate(0, int(expireMouths), 0))
	}
	return
}

// 批量新增用户
func AddLdapUsers(ldap_conn *ldap.Conn, LdapUsers []*LdapAttributes) (AddLdapUsersRes []bool) {
	// 批量处理
	for _, user := range LdapUsers {
		addReq := ldap.NewAddRequest(user.Dn, nil)                                                   // 指定新用户的dn 会同时给cn name字段赋值
		addReq.Attribute("objectClass", []string{"top", "organizationalPerson", "user", "person"})   // 必填字段 否则报错 LDAP Result Code 65 "Object Class Violation"
		addReq.Attribute("employeeNumber", []string{user.Num})                                       // 工号 暂时没用到
		addReq.Attribute("sAMAccountName", []string{user.Sam})                                       // 登录名 必填
		addReq.Attribute("UserAccountControl", []string{user.AccountCtl})                            // 账号控制 544 是启用用户
		addReq.Attribute("accountExpires", []string{strconv.FormatInt(expireTime(user.Expire), 10)}) // 账号过期时间 当前时间加一个时间差并转换为NT时间
		addReq.Attribute("pwdLastSet", []string{user.PwdLastSet})                                    // 用户下次登录必须修改密码 0是永不过期
		addReq.Attribute("displayName", []string{user.DisplayName})                                  // 真实姓名 某些系统需要
		addReq.Attribute("sn", []string{user.Sn})                                                    // 姓
		addReq.Attribute("givenName", []string{user.GivenName})                                      // 名
		addReq.Attribute("mail", []string{user.Email})                                               // 邮箱 必填
		addReq.Attribute("mobile", []string{user.Phone})                                             // 手机号 必填 某些系统需要
		addReq.Attribute("company", []string{user.Company})
		addReq.Attribute("department", []string{user.Depart})
		addReq.Attribute("title", []string{user.Title})

		if err := ldap_conn.Add(addReq); err != nil {
			if ldap.IsErrorWithCode(err, 68) {
				log.Log().Error("User already exist: %s", err)
			} else {
				log.Log().Error("User insert error: %s", err)
			}
			AddLdapUsersRes = append(AddLdapUsersRes, false)
			return
		}
		AddLdapUsersRes = append(AddLdapUsersRes, true)
	}
	return
}

// 修改用户密码 这种修改密码的方法有延迟性 大约五分钟，新旧密码都能使用
func (user *LdapAttributes) ModifyPwd(ldap_conn *ldap.Conn, conn *model.LdapConn, newUserPwd string) (err error) {
	u := FetchUser(ldap_conn, conn, user)
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newUserPwd))
	if err != nil {
		log.Log().Error("转码错误:%s\n", err)
	}

	modReq := ldap.NewModifyRequest(u.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = ldap_conn.Modify(modReq); err != nil {
		log.Log().Error("error setting user password:%s\n", err)
		return
	}
	return
}

// 多条件查询单个用户
func FetchUser(ldap_conn *ldap.Conn, conn *model.LdapConn, user *LdapAttributes) (result *ldap.Entry) {
	// 这里的查询条件必须保证每个用户必须有
	ldapFilterSam := "(sAMAccountName=" + user.Sam + ")"
	ldapFilterEmail := "(mail=" + user.Email + ")"

	searchFilter := "(objectClass=organizationalPerson)"

	if user.Sam != "" {
		searchFilter += ldapFilterSam
	}
	if user.Email != "" {
		searchFilter += ldapFilterEmail
	}
	searchFilter = "(&" + searchFilter + ")"

	searchRequest := ldap.NewSearchRequest(
		conn.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	sr, err := ldap_conn.Search(searchRequest)
	if err != nil {
		log.Log().Error("%s", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries[0]
	}
	return
}

// ldap用户方法——修改dn
func (user *LdapAttributes) ModifyDn(ldap_conn *ldap.Conn, conn *model.LdapConn, cn string) {
	u := FetchUser(ldap_conn, conn, user)
	cn = "CN=" + cn
	modReq := ldap.NewModifyDNRequest(u.DN, cn, true, "")
	if err := ldap_conn.ModifyDN(modReq); err != nil {
		log.Log().Error("Failed to modify DN: %s\n", err)
	}
}

// ldap用户方法——移动dn
func (user *LdapAttributes) MoveDn(ldap_conn *ldap.Conn, conn *model.LdapConn, newOu string) {
	u := FetchUser(ldap_conn, conn, user)
	cn := strings.Split(u.DN, ",")[0]
	movReq := ldap.NewModifyDNRequest(u.DN, cn, true, newOu)
	if err := ldap_conn.ModifyDN(movReq); err != nil {
		log.Log().Error("Failed to move userDN: %s\n", err)
	}

}

// 将ldap库的entry类型转换为自定义类型LdapAttributes
func NewUser(entry *ldap.Entry) *LdapAttributes {
	expire, _ := strconv.ParseInt(entry.GetAttributeValue("accountExpires"), 10, 64)
	return &LdapAttributes{
		Num:         entry.GetAttributeValue("employeeNumber"),
		Sam:         entry.GetAttributeValue("sAMAccountName"),
		DisplayName: entry.GetAttributeValue("distinguishedName"),
		AccountCtl:  entry.GetAttributeValue("UserAccountControl"),
		Expire:      expire,
		PwdLastSet:  entry.GetAttributeValue("pwdLastSet"),
		WhenCreated: entry.GetAttributeValue("whenCreated"),
		WhenChanged: entry.GetAttributeValue("whenChanged"),
		Email:       entry.GetAttributeValue("mail"),
		Phone:       entry.GetAttributeValue("mobile"),
		Name:        entry.GetAttributeValue("displayName"),
		Sn:          entry.GetAttributeValue("sn"),
		GivenName:   entry.GetAttributeValue("givenName"),
		Company:     entry.GetAttributeValue("company"),
		Depart:      entry.GetAttributeValue("department"),
		Title:       entry.GetAttributeValue("title"),
	}
}

// ldap用户方法——修改用户信息
func (user *LdapAttributes) ModifyInfo(ldap_conn *ldap.Conn, conn *model.LdapConn) {
	u := FetchUser(ldap_conn, conn, user)
	modReq := ldap.NewModifyRequest(u.DN, []ldap.Control{})
	// 对用户的普通数据进行选择性更新
	if user.Num != "" {
		modReq.Replace("employeeNumber", []string{user.Company})
	}
	if user.Sam != "" {
		modReq.Replace("sAMAccountName", []string{user.Sam})
	}
	if user.Email != "" {
		modReq.Replace("mail", []string{user.Email})
	}
	if user.Phone != "" {
		modReq.Replace("mobile", []string{user.Phone})
	}
	if user.DisplayName != "" {
		modReq.Replace("displayName", []string{user.DisplayName})
	}
	if user.Depart != "" {
		modReq.Replace("department", []string{user.Depart})
	}
	if user.Company != "" {
		modReq.Replace("company", []string{user.Company})
	}
	if user.Title != "" {
		modReq.Replace("title", []string{user.Title})
	}

	if err := ldap_conn.Modify(modReq); err != nil {
		log.Log().Error("error modify user information:%s\n", err)
	}

	// 对用户DN进行更新 必须放在修改其他普通数据之后
	if user.Dn != "" && strings.SplitN(u.DN, ",", 2)[1] != user.Dn {
		// 将用户转类型后处理
		NewUser(u).MoveDn(ldap_conn, conn, user.Dn)
	}
}
