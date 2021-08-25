package ldap

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
)

var log = logrus.New()

// 禁用/启用用户的 UserAccountControl 状态码
var DisabledLdapUserCodes = [5]int32{514, 546, 66050, 66080, 66082}
var EnabledLdapUserCodes = [4]int32{512, 544, 66048, 262656}

type LdapAttributes struct {
	// ldap字段
	Num            string `json:"employeeNumber" gorm:"type:varchar(100);unique_index"`    // 工号
	Sam            string `json:"sAMAccountName" gorm:"type:varchar(128);unique_index"`    // SAM账号
	Dn             string `json:"distinguishedName" gorm:"type:varchar(100);unique_index"` // dn
	AccountCtl     string `json:"UserAccountControl" gorm:"type:varchar(100)"`             // 用户账户控制
	Expire         int64  `json:"accountExpires" gorm:"type:int(30)"`                      // 账户过期时间
	PwdLastSet     string `json:"pwdLastSet" gorm:"type:varchar(100)"`                     // 用户下次登录必须修改密码
	WhenCreated    string `json:"whenCreated" gorm:"type:varchar(100)"`                    // 创建时间
	WhenChanged    string `json:"whenChanged" gorm:"type:varchar(100)"`                    // 修改时间
	DisplayName    string `json:"displayName" gorm:"type:varchar(32)"`                     // 真实姓名
	Sn             string `json:"sn" gorm:"type:varchar(100)"`                             // 姓
	Name           string `json:"name" gorm:"type:varchar(100)"`                           // 姓名
	GivenName      string `json:"givenName" gorm:"type:varchar(100)"`                      // 名
	Email          string `json:"mail" gorm:"type:varchar(128);unique_index"`              // 邮箱
	Phone          string `json:"mobile" gorm:"type:varchar(32);unique_index"`             // 移动电话
	Company        string `json:"company" gorm:"type:varchar(128)"`                        // 公司
	Depart         string `json:"department" gorm:"type:varchar(128)"`                     // 部门
	Title          string `json:"title" gorm:"type:varchar(100)"`                          // 职务
	WeworkExpire   string `json:"wework_expire" gorm:"-"`                                  // 企业微信过期日期
	WeworkDepartId int    `json:"wework_depart_id" gorm:"-"`                               // 企业微信部门id
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
	"cn",         // common name
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
	if err != nil {
		log.Error("Fail to dial ldap url, err: ", err)
		return
	}

	// 重新连接TLS
	if err = LdapConn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		log.Error("Fail to start tls, err: ", err)
		return
	}

	// 与只读用户绑定
	if err = LdapConn.Bind(LdapCfg.AdminAccount, LdapCfg.Password); err != nil {
		log.Error("admin user auth failed, err: ", err)
		return
	}
	return
}

// 多条件查询用户 返回符合搜索条件的用户列表
func FetchLdapUsers(user *LdapAttributes) (result []*ldap.Entry) {
	// 初始化连接
	err := Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}
	// 多查询条件
	ldapFilterNum := "(employeeNumber=" + user.Num + ")"
	ldapFilterSam := "(sAMAccountName=" + user.Sam + ")"
	ldapFilterEmail := "(mail=" + user.Email + ")"
	ldapFilterPhone := "(mobile=" + user.Phone + ")"
	ldapFilterName := "(displayName=" + user.DisplayName + ")"
	ldapFilterDepart := "(department=" + user.Depart + ")"
	ldapFilterCompany := "(company=" + user.Company + ")"
	ldapFilterTitle := "(title=" + user.Title + ")"

	searchFilter := "(&(objectClass=user)(mail=*))" // 有邮箱的用户 排除系统级别用户

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
		model.LdapCfgs.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 500, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	sr, err := LdapConn.SearchWithPaging(searchRequest, 100)
	if err != nil {
		log.Error("Fail to search users, err: ", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries
	}
	return
}

// ExpireTime 用户过期期限处理 天数为-1 则过期时间为永久;否则 当前时间往后推迟 expireDays 天
func ExpireTime(expireDays int64) (expireTimestamp int64) {
	expireTimestamp = 9223372036854775807
	if expireDays != -1 {
		expireTimestamp = util.UnixToNt(time.Now().AddDate(0, 0, int(expireDays)))
	}
	return
}

// AddUser 新增用户
func AddUser(user *LdapAttributes) (pwd string, err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err:", err)
		return
	}
	// 初始化创建用户请求
	addReq := ldap.NewAddRequest(user.Dn, nil)                                                 // 指定新用户的dn 会同时给cn name字段赋值
	addReq.Attribute("objectClass", []string{"top", "organizationalPerson", "user", "person"}) // 必填字段 否则报错 LDAP Result Code 65 "Object Class Violation"
	addReq.Attribute("employeeNumber", []string{user.Num})                                     // 工号 必填 与显示姓名联合查询唯一用户
	addReq.Attribute("displayName", []string{user.DisplayName})                                // 真实姓名 必填 与工号联合查询唯一用户
	addReq.Attribute("sAMAccountName", []string{user.Sam})                                     // 登录名 必填
	addReq.Attribute("UserAccountControl", []string{user.AccountCtl})                          // 账号控制 544 是启用用户
	addReq.Attribute("accountExpires", []string{strconv.FormatInt(user.Expire, 10)})           // 账号过期时间 当前时间加一个时间差并转换为NT时间
	addReq.Attribute("pwdLastSet", []string{user.PwdLastSet})                                  // 用户下次登录必须修改密码 0是永不过期
	addReq.Attribute("sn", []string{user.Sn})                                                  // 姓
	addReq.Attribute("givenName", []string{user.GivenName})                                    // 名
	addReq.Attribute("mail", []string{user.Email})                                             // 邮箱 必填
	addReq.Attribute("mobile", []string{user.Phone})                                           // 手机号 必填 某些系统需要
	addReq.Attribute("company", []string{user.Company})

	if err = LdapConn.Add(addReq); err != nil {
		if ldap.IsErrorWithCode(err, 68) {
			return
		} else {
			log.Error("Fail to insert user, err: ", err)
		}
		return
	}

	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwd = util.PwdGenerator(8)                                           // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", pwd)) // 密码字符字面值
	if err != nil {
		log.Error("Fail to encode pwd, err: ", err)
	}
	modReq := ldap.NewModifyRequest(user.Dn, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Error("Fail to init pwd, err: ", err)
		return
	}
	return
}

// RetrievePwd 密码找回
func (user *LdapAttributes) RetrievePwd() (sam string, newPwd string, err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}

	entry, _ := FetchUser(user)
	sam = entry.GetAttributeValue("sAMAccountName")
	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	newPwd = util.PwdGenerator(8)                                           // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newPwd)) // 密码字符字面值
	if err != nil {
		log.Error("Fail to encode pwd, err: ", err)
	}
	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Error("Fail to modify pwd, err: ", err)
		return
	}
	return
}

// ModifyPwd 修改用户密码 这种修改密码的方法有延迟性 大约五分钟，新旧密码都能使用
func (user *LdapAttributes) ModifyPwd(newUserPwd string) (err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	entry, _ := FetchUser(user)
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newUserPwd))
	if err != nil {
		log.Error("Fail to encode pwd, err: ", err)
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Error("Fail to set pwd, err: ", err)
		return
	}
	return
}

/* 根据cn查询用户 注意: cn查询不到则会返回管理员用户
 * 这里的查询条件必须保证每个用户必须有
 * 根据cn查询用户 [sam登录名字段也出现了不同的版本 邮箱\手机号都可能更换掉 真实姓名存在重复可能]
 */
func FetchUser(user *LdapAttributes) (result *ldap.Entry, err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}

	ldapFilterCn := "(cn=" + user.DisplayName + user.Num + ")"
	searchFilter := "(objectClass=organizationalPerson)"

	if user.DisplayName != "" && user.Num != "" {
		searchFilter += ldapFilterCn
	}
	searchFilter = "(&" + searchFilter + ")"

	searchRequest := ldap.NewSearchRequest(
		model.LdapCfgs.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		attrs,
		nil,
	)

	// 这里LdapConn 为nil
	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Error("Fail to fetch user, err: ", err)
		return
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries[0]
	}
	return
}

// ModifyDn 修改dn
func (user *LdapAttributes) ModifyDn(cn string) {
	// 初始化连接
	err := Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}

	entry, _ := FetchUser(user)
	cn = "CN=" + cn
	modReq := ldap.NewModifyDNRequest(entry.DN, cn, true, "")
	if err := LdapConn.ModifyDN(modReq); err != nil {
		log.Error("Fail to modify dn, err: ", err)
	}
}

// MoveDn 移动dn
func (user *LdapAttributes) MoveDn(newOu string) (err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	entry, _ := FetchUser(user)
	cn := strings.Split(entry.DN, ",")[0]
	movReq := ldap.NewModifyDNRequest(entry.DN, cn, true, newOu)
	if err = LdapConn.ModifyDN(movReq); err != nil {
		log.Error("Fail to move user dn, err: ", err)
		return
	}
	return

}

// NewUser 将 ldap.Entry 类型转换为自定义类型 LdapAttributes
func NewUser(entry *ldap.Entry) *LdapAttributes {
	// 初始化连接
	err := Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}
	// 将用户过期字段转换为int64
	expire, _ := strconv.ParseInt(entry.GetAttributeValue("accountExpires"), 10, 64)
	return &LdapAttributes{
		Num:         entry.GetAttributeValue("employeeNumber"),
		Sam:         entry.GetAttributeValue("sAMAccountName"),
		DisplayName: entry.GetAttributeValue("displayName"),
		AccountCtl:  entry.GetAttributeValue("UserAccountControl"),
		Expire:      expire,
		PwdLastSet:  entry.GetAttributeValue("pwdLastSet"),
		WhenCreated: entry.GetAttributeValue("whenCreated"),
		WhenChanged: entry.GetAttributeValue("whenChanged"),
		Email:       entry.GetAttributeValue("mail"),
		Phone:       entry.GetAttributeValue("mobile"),
		Sn:          entry.GetAttributeValue("sn"),
		GivenName:   entry.GetAttributeValue("givenName"),
		Company:     entry.GetAttributeValue("company"),
		Depart:      entry.GetAttributeValue("department"),
		Title:       entry.GetAttributeValue("title"),
	}
}

// Update 更新用户信息
func (user *LdapAttributes) Update() (err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	entry, err := FetchUser(user)
	if err != nil {
		log.Error("Fail to fetch user, err: ", err)
		return
	}

	if entry != nil { // 当用户记录存在时
		if user.Num != entry.GetAttributeValue("employeeNumber") &&
			// user.Sam != entry.GetAttributeValue("sAMAccountName") &&
			user.Email != entry.GetAttributeValue("mail") &&
			user.Phone != entry.GetAttributeValue("mobile") &&
			user.DisplayName != entry.GetAttributeValue("displayName") &&
			user.Depart != entry.GetAttributeValue("department") &&
			user.Company != entry.GetAttributeValue("company") &&
			user.Title != entry.GetAttributeValue("title") &&
			user.AccountCtl != entry.GetAttributeValue("UserAccountControl") {
			// fmt.Println(user.DisplayName, strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn))
			fmt.Println("更新用户普通信息")
			modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
			modReq.Replace("employeeNumber", []string{user.Num})
			// modReq.Replace("sAMAccountName", []string{user.Sam})
			modReq.Replace("mail", []string{user.Email})
			modReq.Replace("mobile", []string{user.Phone})
			modReq.Replace("displayName", []string{user.DisplayName})
			modReq.Replace("department", []string{user.Depart})
			modReq.Replace("company", []string{user.Company})
			modReq.Replace("title", []string{user.Title})
			modReq.Replace("accountExpires", []string{strconv.FormatInt(user.Expire, 10)})

			if err := LdapConn.Modify(modReq); err != nil {
				log.Error("Fail to update user's infor: ", err)
			}
		}

		// 若用户部门或状态发生变化 由部门1>>部门2 由部门1>>离职
		if user.Dn != "" {
			if !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
				oldDepart := DnToDeparts(strings.Join(strings.Split(entry.DN, ",")[1:], ","))
				oldDeparts := strings.Split(oldDepart, ".")
				newDepart := DnToDeparts(user.Dn)
				newDeparts := strings.Split(newDepart, ".")
				var level string
				// 若新或旧部门 有一个是外部公司 另一个是内部公司
				if (strings.Contains(oldDepart, "合作伙伴") && !strings.Contains(newDepart, "合作伙伴")) ||
					(!strings.Contains(oldDepart, "合作伙伴") && strings.Contains(newDepart, "合作伙伴")) {
					level = "公司级别"
				} else {
					if oldDeparts[len(oldDeparts)-1] != newDeparts[len(newDeparts)-1] {
						level = "部门级别"
					} else {
						level = "结构级别"
					}
				}
				fmt.Println(user.DisplayName, user.Num, " 岗位变动:[", oldDepart, "]转到[", newDepart, "],类型:", level)
				model.CreateLdapUserDepartRecord(user.DisplayName, user.Num, oldDepart, newDepart, level)
				CheckOuTree(user.Dn)
				err = user.MoveDn(user.Dn)
			}
			return
		}
	}
	return
}

// ModifyInfo 人工修改用户信息
func (user *LdapAttributes) ModifyInfo() (err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}
	entry, err := FetchUser(user)

	if err != nil {
		log.Error("Fail to fetch user, err: ", err)
		return
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
		log.Error("Fail to modify user's information, err: ", err)
	}

	// 对用户DN进行更新 必须放在修改其他普通数据之后
	if user.Dn != "" && !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
		CheckOuTree(user.Dn)
		err = user.MoveDn(user.Dn)
	}
	return
}

// 查询OU是否存在
func IsOuExist(newOu string) (isOuExist bool) {
	// 初始化连接
	err := Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}
	searchRequest := ldap.NewSearchRequest(
		model.LdapCfgs.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalUnit)(distinguishedName="+newOu+"))",
		attrs,
		nil,
	)

	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Error("Fail to fetch ou, err: ", err)
		isOuExist = false

	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		isOuExist = true
	}
	return
}

// AddOu 新增OU 只处理当前OU，不考虑父子OU
func AddOu(newOu string) {
	// 初始化连接
	err := Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
	}
	// 新增逻辑
	addReq := ldap.NewAddRequest(newOu, []ldap.Control{})
	addReq.Attribute("objectClass", []string{"top", "organizationalUnit"})
	addReq.Attribute("cn", []string{strings.Split(strings.Split(newOu, ",")[0], "=")[1]})

	if err := LdapConn.Add(addReq); err != nil {
		log.Error("Fail to add ou, err: ", err)
	}
}

// CheckOuTree 新增OU树逻辑 判断OU树是否存在，若不存在 则层层新增
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

// DepartToDn 将部门架构 aaa.bbb.ccc 转换为LDAP的DN地址 OU=ccc,OU=bbb,OU=aaa,DC=XXX,DC=COM
func DepartToDn(depart string) (dn string) {
	// 从内存中获取公司列表
	var companies map[string]model.CompanyType
	json.Unmarshal([]byte(model.LdapFields.CompanyType), &companies)
	// 如果是外部公司用户
	if companies[strings.Split(depart, ".")[0]].IsOuter {
		depart = DnToDepart(model.LdapFields.BasicPullNode) + ".合作伙伴." + depart
	}

	ous := strings.Split(depart, ".")

	var reversedOus []string = []string{}
	for i := range ous {
		reversedOus = append(reversedOus, ous[len(ous)-i-1])
	}
	dn = strings.Join(reversedOus, ",OU=")
	dn = "OU=" + dn + "," + model.LdapCfgs.BaseDn
	return
}

// DnToDepart 将DN地址转换为部门架构
func DnToDepart(dn string) (depart string) {
	rawDn := strings.Split(dn, ",")
	rawDn = Reverse(rawDn[:len(rawDn)-2]) // 去掉DC 逆序
	// 元素拼接，用.替换所有的OU=，去掉开始的.
	depart = strings.Replace(strings.Join(rawDn, ""), "OU=", ".", -1)[1:]
	return
}

// DnToDeparts 将DN地址转换为多级部门切片
func DnToDeparts(dn string) (departs string) {
	// var temp []string
	rawDn := strings.Split(dn, ",")
	rawDn = Reverse(rawDn[:len(rawDn)-2]) // 去掉DC 逆序
	for i, d := range rawDn {
		rawDn[i] = strings.Trim(strings.ToUpper(d), "OU=")
	}
	departs = strings.Join(rawDn, ".")
	return
}

// 切片逆序
func Reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// ldap用户方法——禁用用户
func (user *LdapAttributes) Disable() (err error) {
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}

	entry, err := FetchUser(user)
	if err != nil {
		log.Error("Fail to fetch user, err: ", err)
		return
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	// 对用户的普通数据进行选择性更新
	modReq.Replace("userAccountControl", []string{"546"})
	if err = LdapConn.Modify(modReq); err != nil {
		log.Error("Fail to disable user, err: ", err)
		return
	}

	// 对用户DN进行更新
	CheckOuTree(model.LdapFields.BaseDnDisabled)
	err = user.MoveDn(model.LdapFields.BaseDnDisabled)
	return
}

// ldap用户方法——账号续期
func (user *LdapAttributes) Renewal() (err error) {
	expireTimeStampStr := strconv.FormatInt(user.Expire, 10)
	// 初始化连接
	err = Init(&model.LdapCfgs)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}

	entry, err := FetchUser(user)
	if err != nil {
		log.Error("Fail to fetch user, err: ", err)
		return
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	// 修改账号过期时间字段
	modReq.Replace("accountExpires", []string{expireTimeStampStr})
	if err = LdapConn.Modify(modReq); err != nil {
		log.Error("Fail to renewal user, err: ", err)
		return
	}
	return
}
