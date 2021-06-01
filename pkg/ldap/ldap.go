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

var (
	LdapConn  *ldap.Conn
	LdapCfg   *model.LdapCfg
	LdapField *model.LdapField
)

// Init 实例化一个 ldapConn
func Init(c *model.LdapCfg) (err error) {
	LdapCfg = &model.LdapCfg{
		ConnUrl:       c.ConnUrl,
		SslEncryption: c.SslEncryption,
		Timeout:       c.Timeout,
		BaseDn:        c.BaseDn,
		AdminAccount:  c.AdminAccount,
		Password:      c.Password,
	}
	// 建立ldap连接
	LdapConn, err = ldap.DialURL(c.ConnUrl)
	// 设置超时时间
	// LdapConn.SetTimeout(time.Duration(c.Timeout))
	if err != nil {
		log.Log().Error("dial ldap url failed,err:%v", err)
		return
	}
	// defer l.Close()
	// 重新连接TLS
	if err = LdapConn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		log.Log().Error("start tls failed,err:%v", err)
		return
	}
	// 与只读用户绑定
	if err = LdapConn.Bind(LdapCfg.AdminAccount, LdapCfg.Password); err != nil {
		log.Log().Error("admin user auth failed,err:%v", err)
		return
	}
	return
}

// 多条件查询用户 返回符合搜索条件的用户列表
func FetchLdapUsers(user *LdapAttributes) (result []*ldap.Entry) {
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
		LdapCfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Log().Error("search users error%s", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries
	}
	return
}

// 用户过期期限处理 天数为-1 则过期时间为永久;否则 当前时间往后推迟 expireDays 天
func ExpireTime(expireDays int64) (expireTimestamp int64) {
	expireTimestamp = 9223372036854775807
	if expireDays != -1 {
		expireTimestamp = util.UnixToNt(time.Now().AddDate(0, 0, int(expireDays)))
	}
	return
}

// 新增用户
func AddUser(user *LdapAttributes) (pwd string, err error) {
	// 初始化创建用户请求
	addReq := ldap.NewAddRequest(user.Dn, nil)                                                 // 指定新用户的dn 会同时给cn name字段赋值
	addReq.Attribute("objectClass", []string{"top", "organizationalPerson", "user", "person"}) // 必填字段 否则报错 LDAP Result Code 65 "Object Class Violation"
	addReq.Attribute("employeeNumber", []string{user.Num})                                     // 工号 暂时没用到
	addReq.Attribute("sAMAccountName", []string{user.Sam})                                     // 登录名 必填
	addReq.Attribute("UserAccountControl", []string{user.AccountCtl})                          // 账号控制 544 是启用用户
	addReq.Attribute("accountExpires", []string{strconv.FormatInt(user.Expire, 10)})           // 账号过期时间 当前时间加一个时间差并转换为NT时间
	addReq.Attribute("pwdLastSet", []string{user.PwdLastSet})                                  // 用户下次登录必须修改密码 0是永不过期
	addReq.Attribute("displayName", []string{user.DisplayName})                                // 真实姓名 某些系统需要
	addReq.Attribute("sn", []string{user.Sn})                                                  // 姓
	addReq.Attribute("givenName", []string{user.GivenName})                                    // 名
	addReq.Attribute("mail", []string{user.Email})                                             // 邮箱 必填
	addReq.Attribute("mobile", []string{user.Phone})                                           // 手机号 必填 某些系统需要
	addReq.Attribute("company", []string{user.Company})
	addReq.Attribute("department", []string{user.Depart})
	addReq.Attribute("title", []string{user.Title})

	if err = LdapConn.Add(addReq); err != nil {
		if ldap.IsErrorWithCode(err, 68) {
			return
		} else {
			log.Log().Error("User insert error: %s", err)
		}
		return
	}

	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwd = util.RandStringBytesMaskImprSrcUnsafe(8)                       // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", pwd)) // 密码字符字面值
	if err != nil {
		log.Log().Error("初始化复杂密码时转码错误:%s\n", err)
	}
	modReq := ldap.NewModifyRequest(user.Dn, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log().Error("初始化复杂密码执行报错:%s\n", err)
		return
	}
	return
}

// 密码找回
func (user *LdapAttributes) ResetPwd() (sam string, newPwd string, err error) {
	entry, _ := FetchUser(user)
	sam = entry.GetAttributeValue("sAMAccountName")
	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	newPwd = util.RandStringBytesMaskImprSrcUnsafe(8)                       // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newPwd)) // 密码字符字面值
	if err != nil {
		log.Log().Error("重置密码时转码错误:%s\n", err)
	}
	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log().Error("重置密码执行报错:%s\n", err)
		return
	}
	return
}

// 修改用户密码 这种修改密码的方法有延迟性 大约五分钟，新旧密码都能使用
func (user *LdapAttributes) ModifyPwd(newUserPwd string) (err error) {
	entry, _ := FetchUser(user)
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newUserPwd))
	if err != nil {
		log.Log().Error("转码错误:%s\n", err)
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log().Error("error setting user password:%s\n", err)
		return
	}
	return
}

/* 根据cn查询用户 注意: cn查询不到则会返回管理员用户
 * 这里的查询条件必须保证每个用户必须有
 * 根据cn查询用户 [sam登录名字段也出现了不同的版本 邮箱\手机号都可能更换掉 真实姓名存在重复可能]
 */
func FetchUser(user *LdapAttributes) (result *ldap.Entry, err error) {
	ldapFilterCn := "(cn=" + user.Name + user.Num + ")"
	log.Log().Error(ldapFilterCn)
	searchFilter := "(objectClass=organizationalPerson)"

	if user.Name != "" && user.Num != "" {
		searchFilter += ldapFilterCn
	}
	searchFilter = "(&" + searchFilter + ")"

	searchRequest := ldap.NewSearchRequest(
		LdapCfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	// 这里LdapConn 为nil
	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Log().Error("fetch user error%s", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries[0]
	}
	return
}

// ldap用户方法——修改dn
func (user *LdapAttributes) ModifyDn(cn string) {
	entry, _ := FetchUser(user)
	cn = "CN=" + cn
	modReq := ldap.NewModifyDNRequest(entry.DN, cn, true, "")
	if err := LdapConn.ModifyDN(modReq); err != nil {
		log.Log().Error("Failed to modify DN: %s\n", err)
	}
}

// ldap用户方法——移动dn
func (user *LdapAttributes) MoveDn(newOu string) {
	entry, _ := FetchUser(user)
	cn := strings.Split(entry.DN, ",")[0]
	movReq := ldap.NewModifyDNRequest(entry.DN, cn, true, newOu)
	if err := LdapConn.ModifyDN(movReq); err != nil {
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

// Updateldap用户方法——更新用户信息
func (user *LdapAttributes) Update() {
	entry, err := FetchUser(user)
	if err != nil {
		log.Log().Error("fetch user failed:%s", err)
		panic(err)
	}
	if entry != nil {
		modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
		// 对用户的普通数据进行选择性更新
		if user.Num != "" {
			modReq.Replace("employeeNumber", []string{user.Num})
		}
		// 更新用户登录名 [涉及平台较多 需统一更新!!]
		// if user.Sam != "" {
		// 	modReq.Replace("sAMAccountName", []string{user.Sam})
		// }
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
		if user.AccountCtl != "" {
			modReq.Replace("accountExpires", []string{strconv.FormatInt(user.Expire, 10)})
		}

		if err := LdapConn.Modify(modReq); err != nil {
			log.Log().Error("error update user information:%s\n", err)
		}
		// 对用户DN进行更新 必须放在修改其他普通数据之后 [不会影响用户使用体验]
		if user.Dn != "" && !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
			CheckOuTree(user.Dn)
			user.MoveDn(user.Dn)
		}
	} else {
		log.Log().Error("无此用户")
	}
}

// ldap用户方法——手动修改用户信息
func (user *LdapAttributes) ModifyInfo() {
	entry, err := FetchUser(user)

	if err != nil {
		log.Log().Error("fetch user failed:%s", err)
	}
	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	// 对用户的普通数据进行选择性更新
	if user.Num != "" {
		modReq.Replace("employeeNumber", []string{user.Num})
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

	if err := LdapConn.Modify(modReq); err != nil {
		log.Log().Error("error modify user information:%s\n", err)
	}

	// 对用户DN进行更新 必须放在修改其他普通数据之后
	if user.Dn != "" && !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
		CheckOuTree(user.Dn)
		user.MoveDn(user.Dn)
	}
}

// 查询OU是否存在
func IsOuExist(newOu string) (isOuExist bool) {
	searchRequest := ldap.NewSearchRequest(
		LdapCfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalUnit)(distinguishedName="+newOu+"))",
		attrs,
		nil,
	)

	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Log().Error("error fetch ou:%s\n", err)
		isOuExist = false

	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		isOuExist = true
	}
	return
}

// 新增OU 只处理当前OU，不考虑父子OU
func AddOu(newOu string) {
	// 新增逻辑
	addReq := ldap.NewAddRequest(newOu, []ldap.Control{})
	addReq.Attribute("objectClass", []string{"top", "organizationalUnit"})
	addReq.Attribute("cn", []string{strings.Split(strings.Split(newOu, ",")[0], "=")[1]})

	if err := LdapConn.Add(addReq); err != nil {
		log.Log().Error("error add ou:%s\n", err)
	}
}

// 新增OU树逻辑 判断OU树是否存在，若不存在 则层层新增
func CheckOuTree(newOu string) {
	ous := strings.SplitN(newOu, ",", len(strings.Split(newOu, ","))-1)
	for i := range ous {
		if i != 0 {
			dn := strings.Join(ous[len(ous)-i-1:], ",") // 获取每层DN地址
			// 查询dn树中每一层是否都存在
			isOuExist := IsOuExist(dn)
			if !isOuExist { // 如果不存在则新增
				AddOu(dn) // 为了安全 充分测试后再启用
			}
		}
	}
}

// 将部门架构转换为LDAP的DN地址
func DepartToDn(depart string) (dn string) {
	ous := strings.Split(depart, ".")
	// if !strings.Contains(depart, ".") {
	// 	ous = strings.Split("合作伙伴."+depart, ".")
	// }

	var reversedOus []string = []string{}
	for i := range ous {
		reversedOus = append(reversedOus, ous[len(ous)-i-1])
	}
	dn = strings.Join(reversedOus, ",OU=")
	dn = "OU=" + dn + "," + LdapCfg.BaseDn
	return
}
