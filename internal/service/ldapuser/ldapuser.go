package ldapuser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/email"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
)

var (
	DisabledLdapUserCodes = [5]int32{514, 546, 66050, 66080, 66082} // 禁用用户的 UserAccountControl 状态码
	EnabledLdapUserCodes  = [4]int32{512, 544, 66048, 262656}       // 启用用户的 UserAccountControl 状态码
	ErrLdapUserNotFound   = errors.New("fail to fetch ldap user")   // 未找到 LDAP 用户错误
	// LDAP 用户属性
	attrs = []string{
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
)

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
	ProbationFlag  int    `json:"probation_flag" gorm:"-"`                                 // 打试用期标签 1 打标签 0 不打标签
}

// FetchLdapUsers 多条件查询用户 返回符合搜索条件的用户列表
func FetchLdapUsers(user *LdapAttributes) (result []*ldap.Entry) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
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
		log.Log.Error("Fail to search users, err: ", err)
	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries
	}
	return
}

// AddUser 新增用户
func AddUser(user *LdapAttributes) (pwd string, err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

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
			log.Log.Error("Fail to insert ldap user, err: ", err)
		}
		return
	}

	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwd = util.PwdGenerator(8)                                           // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", pwd)) // 密码字符字面值
	if err != nil {
		log.Log.Error("Fail to encode pwd, err: ", err)
	}
	modReq := ldap.NewModifyRequest(user.Dn, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log.Error("Fail to init pwd, err: ", err)
		return
	}
	return
}

// RetrievePwd 密码找回
func (user *LdapAttributes) RetrievePwd() (sam string, newPwd string, err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}

	sam = entry.GetAttributeValue("sAMAccountName")
	// 初始化复杂密码
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	newPwd = util.PwdGenerator(8)                                           // 密码字符串
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newPwd)) // 密码字符字面值
	if err != nil {
		log.Log.Error("Fail to encode pwd, err: ", err)
	}
	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log.Error("Fail to modify pwd, err: ", err)
		return
	}
	return
}

// ModifyPwd 修改用户密码 这种修改密码的方法有延迟性 大约五分钟，新旧密码都能使用
func (user *LdapAttributes) ModifyPwd(newUserPwd string) (err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("%q", newUserPwd))
	if err != nil {
		log.Log.Error("Fail to encode pwd, err: ", err)
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	modReq.Replace("unicodePwd", []string{pwdEncoded})

	if err = LdapConn.Modify(modReq); err != nil {
		log.Log.Error("Fail to set pwd, err: ", err)
		return
	}
	return
}

// FetchUser 根据cn查询用户 注意: cn查询不到则会返回管理员用户
func FetchUser(user *LdapAttributes) (result *ldap.Entry, err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

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

	// search user
	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	// 查询结果判断
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		result = sr.Entries[0]
	} else {
		return nil, ErrLdapUserNotFound
	}
	return
}

func (user *LdapAttributes) ModifyDn(cn string) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}
	cn = "CN=" + cn
	modReq := ldap.NewModifyDNRequest(entry.DN, cn, true, "")
	if err := LdapConn.Conn.ModifyDN(modReq); err != nil {
		log.Log.Error("Fail to modify dn, err: ", err)
	}
}

// MoveDn 移动dn
func (user *LdapAttributes) MoveDn(newOu string) (err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}
	cn := strings.Split(entry.DN, ",")[0]
	movReq := ldap.NewModifyDNRequest(entry.DN, cn, true, newOu)
	if err = LdapConn.Conn.ModifyDN(movReq); err != nil {
		log.Log.Error("Fail to move user dn, err: ", err)
		return
	}
	return

}

// NewUser 将 ldap.Entry 类型转换为自定义类型 LdapAttributes
func NewUser(entry *ldap.Entry) *LdapAttributes {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
	}
	defer LdapConn.Close()

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
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
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
			// 更新用户普通信息
			modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
			modReq.Replace("employeeNumber", []string{user.Num})
			modReq.Replace("mail", []string{user.Email})
			modReq.Replace("mobile", []string{user.Phone})
			modReq.Replace("displayName", []string{user.DisplayName})
			modReq.Replace("department", []string{user.Depart})
			modReq.Replace("company", []string{user.Company})
			modReq.Replace("title", []string{user.Title})
			modReq.Replace("accountExpires", []string{strconv.FormatInt(user.Expire, 10)})

			if err := LdapConn.Modify(modReq); err != nil {
				log.Log.Error("Fail to update user's info: ", err)
			}
		}

		// 若用户部门或状态发生变化 由部门1>>部门2 由部门1>>离职
		if user.Dn != "" {
			if !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
				oldDepart := util.DnToDeparts(strings.Join(strings.Split(entry.DN, ",")[1:], ","))
				oldDeparts := strings.Split(oldDepart, ".")
				newDepart := util.DnToDeparts(user.Dn)
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
				log.Log.Info(user.DisplayName, user.Num, " 岗位变动:[", oldDepart, "]转到[", newDepart, "],类型:", level)
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
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
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
		log.Log.Error("Fail to modify user's information, err: ", err)
	}

	// 对用户DN进行更新 必须放在修改其他普通数据之后
	if user.Dn != "" && !strings.EqualFold(strings.SplitN(entry.DN, ",", 2)[1], user.Dn) {
		CheckOuTree(user.Dn)
		err = user.MoveDn(user.Dn)
	}
	return
}

// IsOuExist 查询OU是否存在
func IsOuExist(newOu string) (isOuExist bool) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
	}
	defer LdapConn.Close()

	searchRequest := ldap.NewSearchRequest(
		model.LdapCfgs.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalUnit)(distinguishedName="+newOu+"))",
		attrs,
		nil,
	)

	sr, err := LdapConn.Search(searchRequest)
	if err != nil {
		log.Log.Error("Fail to fetch ou, err: ", err)
		isOuExist = false

	}
	if len(sr.Entries) > 0 && len(sr.Entries[0].Attributes) > 0 {
		isOuExist = true
	}
	return
}

// AddOu 新增OU 只处理当前OU，不考虑父子OU
func AddOu(newOu string) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
	}
	defer LdapConn.Close()

	// 新增逻辑
	addReq := ldap.NewAddRequest(newOu, []ldap.Control{})
	addReq.Attribute("objectClass", []string{"top", "organizationalUnit"})
	addReq.Attribute("cn", []string{strings.Split(strings.Split(newOu, ",")[0], "=")[1]})

	if err := LdapConn.Add(addReq); err != nil {
		log.Log.Error("Fail to add ou, err: ", err)
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
		depart = util.DnToDepart(model.LdapFields.BasicPullNode) + ".合作伙伴." + depart
	}

	ous := strings.Split(depart, ".")

	var reversedOus []string
	for i := range ous {
		reversedOus = append(reversedOus, ous[len(ous)-i-1])
	}
	dn = strings.Join(reversedOus, ",OU=")
	dn = "OU=" + dn + "," + model.LdapCfgs.BaseDn
	return
}

// Disable ldap用户方法——禁用用户 禁用用户不修改用户的OU
func (user *LdapAttributes) Disable() (err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	// 对用户的普通数据进行选择性更新
	modReq.Replace("userAccountControl", []string{"546"})
	if err = LdapConn.Modify(modReq); err != nil {
		log.Log.Error("Fail to disable user, err: ", err)
		return
	}
	return
}

// Renewal ldap用户方法——账号续期
func (user *LdapAttributes) Renewal() (err error) {
	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}

	modReq := ldap.NewModifyRequest(entry.DN, []ldap.Control{})
	// 修改账号过期时间字段
	expireTimeStampStr := strconv.FormatInt(user.Expire, 10)
	modReq.Replace("accountExpires", []string{expireTimeStampStr})
	if err = LdapConn.Modify(modReq); err != nil {
		log.Log.Error("Fail to renewal user, err: ", err)
		return
	}
	return
}

// LdapUserService LDAP用户查询条件
type LdapUserService struct {
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
}

// ScanExpiredUsersManual 手动触发扫描过期ldap用户
func ScanExpiredUsersManual() serializer.Response {
	go func() {
		ScanExpiredUsers()
	}()
	return serializer.Response{Data: 0}
}

// ScanExpiredUsers 扫描过期ldap用户
func ScanExpiredUsers() {
	LdapUsers := FetchLdapUsers(&LdapAttributes{})
	currentTime := time.Now()
	// expireLdapUsers := make([]*ldap.LdapAttributes, 0, 10)  // TODO 预留防止更改传参
	for _, u := range LdapUsers {
		expire, _ := strconv.ParseInt(u.GetAttributeValue("accountExpires"), 10, 64)
		expireDays := util.FormatLdapExpireDays(util.SubDays(util.NtToUnix(expire), currentTime))
		if expireDays != 106752 { // 排除不过期的账号
			if expireDays >= -7 && expireDays <= 14 { // 未/已经过期 7 天内的账号
				ldapUser := NewUser(u) // 初始化速度较慢 适用定时异步任务处理少量数据
				// expireLdapUsers = append(expireLdapUsers, ldapUser)
				log.Log.Info(ldapUser, " 过期天数: ", expireDays)
				HandleExpiredLdapUsers(ldapUser, expireDays) // 处理过期账号
			}
		}

	}
}

// SyncUsersManual 手动触发更新ldap用户
func SyncUsersManual() serializer.Response {
	go func() {
		SyncUsers()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发更新ldap用户成功!"}
}

// SyncUsers 更新ldap用户
func SyncUsers() {
	// 从缓存取HR元数据
	ldapUsers, err := cache.HGetAll("hr_users")
	if err != nil {
		log.Log.Error("Fail to fetch ldap users cache,:", err)
	}

	log.Log.Info("开始更新ldap用户...")
	var wg sync.WaitGroup
	ch := make(chan struct{}, 20)
	for cn, u := range ldapUsers {
		ch <- struct{}{}
		wg.Add(1)
		go func(cn string, u string) {
			defer wg.Done()
			var userStat, dn string
			var expire int64
			var user hr.User

			json.Unmarshal([]byte(u), &user) // 反序列化
			if user.Stat == "离职" {
				userStat = "546"
				dn = model.LdapFields.BaseDnDisabled // 禁用部门
				expire = 0                           // 账号失效
			} else { // 在职员工
				userStat = "544"                 // 账号有效
				dn = DepartToDn(user.Department) // 将部门转换为DN
				expire = util.ExpireTime(-1)     // 账号永久有效
			}
			depart := strings.Split(user.Department, ".")[len(strings.Split(user.Department, "."))-1]
			name := []rune(user.Name)

			// 将hr数据转换为ldap信息格式
			ldapUser := &LdapAttributes{
				Num:         user.Eid,
				Sam:         user.Eid,
				DisplayName: user.Name,
				Email:       user.Mail,
				Phone:       user.Mobile,
				Dn:          dn,
				PwdLastSet:  "0", // 用户下次必须修改密码 0
				AccountCtl:  userStat,
				Expire:      expire,
				Sn:          string(name[0]),
				Name:        user.Name,
				GivenName:   string(name[1:]),
				Company:     user.CompanyName,
				Depart:      depart,
				Title:       user.Title,
			}
			// 更新用户操作
			err := ldapUser.Update()
			if err != nil {
				if err == ErrLdapUserNotFound {
					// Do nothing
				} else {
					log.Log.Error("Fail to update user form cache to conn server,: ", err)
				}
			}
			<-ch
		}(cn, u)
	}
	wg.Wait() // 等待
	log.Log.Info("更新ldap用户成功!")

	// 更新完成后，将数据库更改记录统一公布
	now := time.Now()
	// 若是周一 则将周末的处理结果一并发出
	var ldapUserDepartRecords []model.LdapUserDepartRecord
	if util.IsMonday(now) {
		ldapUserDepartRecords, _ = model.FetchLdapUserDepartRecord(-2, 1)
	} else {
		ldapUserDepartRecords, _ = model.FetchLdapUserDepartRecord(0, 1)
	}

	today := now.Format("2006年01月02日")
	tempTitle := `<font color="warning"> ` + today + ` </font>LDAP用户架构变化：`
	temp := `>%s. <font color="warning"> %s </font>岗位变动:<font color="comment"> %s </font>到<font color="info"> %s </font>级别<font color="warning"> %s </font>`
	var msgs string
	for i, r := range ldapUserDepartRecords {
		if i != len(ldapUserDepartRecords) {
			msgs += "\n\n"
		}
		msgs += fmt.Sprintf(temp, strconv.Itoa(i+1), r.Name, r.OldDepart, r.NewDepart, r.Level)
	}

	// 根据是否为节假日决定是否发消息
	if isSilent, festival := util.IsHolidaySilentMode(now); isSilent {
		if festival != "" {
			util.SendRobotMsg(`<font color="warning"> ` + festival + "快乐！祝各位阖家团圆、岁岁平安~" + ` </font>`)
		} else {
			// 不是节日 只是周末静默
		}
	} else {
		// 工作日正常发送通知
		if len(ldapUserDepartRecords) == 0 {
			util.SendRobotMsg(`<font color="warning"> ` + today + ` </font>LDAP用户架构无变化`)
		} else {
			// 消息过长 作剪裁处理
			msgs := util.TruncateMsg(tempTitle+msgs, "\n\n")
			for _, m := range msgs {
				util.SendRobotMsg(m)
			}
		}
	}

	log.Log.Info("汇总通知发送成功!")
}

// FormatData 校验邮箱和手机号格式
func FormatData(mail string, mobile string) (err error) {
	if strings.Contains(mail, " ") || strings.Contains(mobile, " ") || len(mobile) != 11 {
		err = errors.New("手机号或邮箱不符合规范! 1. 手机号11位且中间不允许有空格 2. 邮箱中间不允许有空格")
	}
	return
}

// CreateLdapUser 生成ldap用户操作
func CreateLdapUser(o order.WeworkOrderDetailsAccountsRegister, user *LdapAttributes) (err error) {
	// 数据校验 1. 手机号11位 中间不允许有空格 2. 邮箱中间不允许有空格
	err = FormatData(user.Email, user.Phone)
	if err != nil {
		// 创建成功发送企业微信消息
		createAccountsWeworkMsgTemplate, _ := cache.HGet("wework_msg_templates", "wework_template_uuap_register_err")
		msg := map[string]interface{}{
			"touser":  o.Userid,
			"msgtype": "markdown",
			"agentid": model.WeworkUuapCfg.AppId,
			"markdown": map[string]interface{}{
				"content": fmt.Sprintf(createAccountsWeworkMsgTemplate, o.SpName, user.DisplayName, err),
			},
		}
		_, err = model.CorpAPIMsg.MessageSend(msg)
		if err != nil {
			log.Log.Error("Fail to send wework msg, err: ", err)
		}
		log.Log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【初次注册-手机号|邮箱格式错误】")
		return
	}

	// 创建LDAP用户 生成初始密码
	pwd, err := AddUser(user)
	if err != nil {
		log.Log.Error("Fail to create ldap user, err: ", err)
		// 此处的错误一般是账号已经存在 为了防止其他错误，这里输出日志
		err = HandleUuapDuplicateRegister(user, o)
		return
	}

	// 创建成功发送企业微信消息
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_uuap_register")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}

	msg := map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, o.SpName, user.Sam, pwd),
		},
	}
	_, err = model.CorpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【初次注册】")
	return
}

// HandleExpiredLdapUsers 处理过期用户
func HandleExpiredLdapUsers(user *LdapAttributes, expireDays int) (err error) {
	emailTempUuaplateExpiring, err := cache.HGet("email_templates", "email_template_uuap_expiring")
	if err != nil {
		log.Log.Error("读取即将过期邮件消息模板错误: ", err)
	}
	emailTemplateUuapExpired, err := cache.HGet("email_templates", "email_template_uuap_expired")
	if err != nil {
		log.Log.Error("读取已过期邮件消息模板错误: ", err)
	}
	emailTemplateUuapExpiredDisabled, err := cache.HGet("email_templates", "email_template_uuap_expired_disabled")
	if err != nil {
		log.Log.Error("读取已过期禁用邮件消息模板错误: ", err)
	}

	if expireDays == 7 || expireDays == 14 { // 即将过期提醒 7|14天前
		address := []string{user.Email}
		htmlContent := fmt.Sprintf(emailTempUuaplateExpiring, user.DisplayName, user.Sam, strconv.Itoa(expireDays))
		err = email.SendMailHtml(address, "UUAP账号即将过期通知", htmlContent)
		if err != nil {
			log.Log.Error("发邮件错误: ", err)
			return
		}
		log.Log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【即将过期】")
	} else if expireDays == -7 { // 已经过期提醒 7天后
		address := []string{user.Email}
		htmlContent := fmt.Sprintf(emailTemplateUuapExpired, user.DisplayName, user.Sam, strconv.Itoa(-expireDays)) // 此时过期天数为负值
		email.SendMailHtml(address, "UUAP账号已过期通知", htmlContent)
		log.Log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【已经过期】")
	} else if expireDays == -30 { // 已经过期且禁用提醒 30天后
		fmt.Println(emailTemplateUuapExpiredDisabled)
		log.Log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【已经过期禁用】")
	}
	return
}

// HandleUuapDuplicateRegister 处理重复注册
func HandleUuapDuplicateRegister(user *LdapAttributes, order order.WeworkOrderDetailsAccountsRegister) (err error) {
	duplicateRegisterUuapUserWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_uuap_user_duplicate_register")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}

	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := FetchUser(user)
	if err != nil {
		return
	}
	sam := entry.GetAttributeValue("sAMAccountName")

	_, err = model.CorpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(duplicateRegisterUuapUserWeworkMsgTemplate, order.SpName, user.DisplayName, sam),
		},
	})
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Log.Info("企业微信回执消息:工单【" + order.SpName + "】用户【" + order.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【已注册过的UUAP用户】")
	return
}
