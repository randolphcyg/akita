package user

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/email"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"

	log "github.com/sirupsen/logrus"
)

/*
* 这里是外部接口(HR数据)的模型
 */
// HrDataService HR数据查询条件
type HrDataService struct {
	// 获取 token 的 URL
	UrlGetToken string `json:"url_get_token" gorm:"type:varchar(255);not null;comment:获取token的地址"`
	// 获取 数据 的URL
	UrlGetData string `json:"url_get_data" gorm:"type:varchar(255);not null;comment:获取数据的地址"`
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

// 查
func (service *LdapUserService) FetchUser(url string) serializer.Response {
	// 初始化连接
	user := &ldap.LdapAttributes{}
	LdapUsers := ldap.FetchLdapUsers(user)
	jndex := 1
	for _, user := range LdapUsers {
		// 速度慢
		// ldapUser := ldap.NewUser(user)
		// fmt.Println(ldapUser.DisplayName, ldapUser.Num, ldapUser.Sam, ldapUser.Email, ldapUser.Phone)
		// 速度快 查询激活用户 判断带前缀的账号
		if user.GetAttributeValue("userAccountControl") == "544" && (strings.HasPrefix(user.GetAttributeValue("sAMAccountName"), "X") || strings.HasPrefix(user.GetAttributeValue("sAMAccountName"), "XXXX")) {
			fmt.Println(jndex,
				user.GetAttributeValue("displayName"),
				user.GetAttributeValue("employeeNumber"),
				user.GetAttributeValue("sAMAccountName"),
				strings.ToUpper(user.GetAttributeValue("mail")),
				user.GetAttributeValue("mobile"),
				user.GetAttributeValue("userAccountControl"))
			jndex += 1
		}
	}
	return serializer.Response{Data: LdapUsers}
}

// TODO 创建用户-管理员接口
func (service *LdapUserService) CreateUser(u LdapUserService) serializer.Response {

	return serializer.Response{Data: 0, Msg: "UUAP账号创建成功"}
}

// TODO 查询HR数据
func (service *HrDataService) FetchHrData(h HrDataService) serializer.Response {
	// 初始化连接
	// var c conn.LdapConnService

	// conn, err := c.FetchByConnUrl(url)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// LdapUsers := ldap.FetchUser(&conn)
	// for _, user := range LdapUsers {
	// 	user.PrettyPrint(2)
	// 	fmt.Println(user.GetAttributeValue("displayName"))
	// 	break
	// }
	return serializer.Response{Data: 1111}
}

// ScanExpiredLdapUsersManual 手动扫描过期ldap用户
func ScanExpiredLdapUsersManual() serializer.Response {
	go func() {
		HandleExpiredLdapUsers()
	}()
	return serializer.Response{Data: 0}
}

func TestTask() {
	fmt.Println("执行定任务")
}

// HandleExpiredLdapUsers 处理过期用户 (expireLdapUsers []*ldap.LdapAttributes)
func HandleExpiredLdapUsers() {
	// 初始化连接
	LdapUsers := ldap.FetchLdapUsers(&ldap.LdapAttributes{})
	currentTime := time.Now()
	// expireLdapUsers := make([]*ldap.LdapAttributes, 0, 10)  // TODO 预留防止更改传参
	for _, u := range LdapUsers {
		expire, _ := strconv.ParseInt(u.GetAttributeValue("accountExpires"), 10, 64)
		expireDays := util.FormatLdapExpireDays(util.SubDays(util.NtToUnix(expire), currentTime))
		if expireDays != 106752 { // 排除不过期的账号
			if expireDays >= -7 && expireDays <= 14 { // 未/已经过期 7 天内的账号
				ldapUser := ldap.NewUser(u) // 初始化速度较慢 适用定时异步任务处理少量数据
				// expireLdapUsers = append(expireLdapUsers, ldapUser)
				log.Info(ldapUser, "过期天数: ", expireDays)
				HandleOrderUuapExpired(ldapUser, expireDays) // 处理过期账号
			}
		}
	}
	// return
}

// UpdateCacheUsersManual 更新HR元数据到缓存
func UpdateCacheUsersManual() serializer.Response {
	go func() {
		HrToCache() // 更新HR元数据到缓存
	}()
	return serializer.Response{Data: 0, Msg: "从HR数据接口更新用户到缓存成功!"}
}

// UpdateLdapUsersManual 手动更新用户
func UpdateLdapUsersManual() serializer.Response {
	go func() {
		CacheToLdap()
	}()
	return serializer.Response{Data: 0, Msg: "从缓存更新用户到ldap成功!"}
}

// HrToCache 将HR元数据存到缓存
func HrToCache() {
	log.Info("获取HR接口数据中......")
	var hrDataConn hr.HrDataConn
	if result := model.DB.First(&hrDataConn); result.Error != nil {
		log.Error("Fail to get HR data connection cfg!")
	}
	hrUsers := hr.FetchHrData(&hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	})

	// 先清空缓存
	_, err := cache.HDel("ldap_users")
	if err != nil {
		log.Error("Fail to clean ldap users cache,:", err)
	}

	// 将HR接口元数据写入缓存
	for _, user := range hrUsers {
		userData, _ := json.Marshal(user)
		_, err := cache.HSet("ldap_users", user.Name+user.Eid, userData)
		if err != nil {
			log.Error("Fail to update ldap users to cache,:", err)
		}
	}
	log.Info("从HR数据接口更新用户到缓存成功!")
}

// CacheToLdap 将HR缓存数据更新到ldap
func CacheToLdap() {
	// 从缓存取所有HR元数据
	ldapUsers, err := cache.HGetAll("ldap_users")
	if err != nil {
		log.Error("Fail to fetch ldap users cache,:", err)
	}

	// TODO 待优化 目前速度提升没有
	done := make(chan int, 30) // 带 20 个缓存
	for cn, u := range ldapUsers {
		go func(cn string, u string) {
			var userStat, dn string
			var expire int64
			var user hr.HrUser

			json.Unmarshal([]byte(u), &user) // 反序列化
			if user.Stat == "离职" {
				userStat = "546"
				dn = bootstrap.LdapField.BaseDnDisabled // 禁用部门
				expire = 0                              // 账号失效
			} else { // 在职员工
				userStat = "544"                      // 账号有效
				dn = ldap.DepartToDn(user.Department) // 将部门转换为DN
				expire = ldap.ExpireTime(-1)          // 账号永久有效
			}
			depart := strings.Split(user.Department, ".")[len(strings.Split(user.Department, "."))-1]
			name := []rune(user.Name)

			// 将hr数据转换为ldap信息格式
			ldapUser := &ldap.LdapAttributes{
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
			fmt.Println(cn)
			// 更新用户操作
			err := ldapUser.Update()
			if err != nil {
				log.Error("Fail to update user form cache to ldap server,: ", err)
			}
			done <- 30
		}(cn, u)
		<-done
	}

	log.Info("从缓存更新用户到ldap成功!")
}

func CreateLdapUser(o order.WeworkOrderDetailsAccountsRegister, user *ldap.LdapAttributes) (err error) {
	// 执行生成UUAP操作
	// 创建LDAP用户 生成初始密码
	pwd, err := ldap.AddUser(user)
	if err != nil {
		log.Error("Fail to create ldap user, err: ", err)
		// 此处的错误一般是账号已经存在 为了防止其他错误，这里输出日志
		err = HandleOrderUuapDuplicateRegister(user, o)
		return
	}

	// 创建成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_register")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}

	msg := map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, o.SpName, user.Sam, pwd),
		},
	}
	_, err = corpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Error("Fail to send wework msg, err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【初次注册】")
	return
}

// 过期用户处理
func HandleOrderUuapExpired(user *ldap.LdapAttributes, expireDays int) (err error) {
	emailTempUuaplateExpiring, err := cache.HGet("email_templates", "email_template_uuap_expiring")
	if err != nil {
		log.Error("读取即将过期邮件消息模板错误: ", err)
	}
	emailTemplateUuapExpired, err := cache.HGet("email_templates", "email_template_uuap_expired")
	if err != nil {
		log.Error("读取已过期邮件消息模板错误: ", err)
	}
	emailTemplateUuapExpiredDisabled, err := cache.HGet("email_templates", "email_template_uuap_expired_disabled")
	if err != nil {
		log.Error("读取已过期禁用邮件消息模板错误: ", err)
	}

	if expireDays == 7 || expireDays == 14 { // 即将过期提醒 7|14天前
		// fmt.Println(emailTempUuaplateExpiring)
		address := []string{user.Email}
		htmlContent := fmt.Sprintf(emailTempUuaplateExpiring, user.DisplayName, user.Sam, strconv.Itoa(expireDays))
		email.SendMailHtml(address, "UUAP账号即将过期通知", htmlContent)
		log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【即将过期】")
	} else if expireDays == -7 { // 已经过期提醒 7天后
		// fmt.Println(emailTemplateUuapExpired)
		address := []string{user.Email}
		htmlContent := fmt.Sprintf(emailTemplateUuapExpired, user.DisplayName, user.Sam, strconv.Itoa(-expireDays)) // 此时过期天数为负值
		email.SendMailHtml(address, "UUAP账号已过期通知", htmlContent)
		log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【已经过期】")
	} else if expireDays == -30 { // 已经过期且禁用提醒 30天后
		fmt.Println(emailTemplateUuapExpiredDisabled)
		log.Info("邮件发送成功！用户【" + user.DisplayName + "】账号【" + user.Sam + "】状态【已经过期禁用】")
	}
	return
}

// 重复提交注册申请
func HandleOrderUuapDuplicateRegister(user *ldap.LdapAttributes, order order.WeworkOrderDetailsAccountsRegister) (err error) {
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	duplicateRegisterUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_duplicate_register")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}

	// 初始化连接
	err = ldap.Init(&bootstrap.LdapCfg)
	if err != nil {
		log.Error("Fail to get ldap connection, err: ", err)
		return
	}

	entry, _ := ldap.FetchUser(user)
	sam := entry.GetAttributeValue("sAMAccountName")

	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(duplicateRegisterUuapWeworkMsgTemplate, order.SpName, user.DisplayName, sam),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg, err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Info("企业微信回执消息:工单【" + order.SpName + "】用户【" + order.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【已注册过的用户】")
	return
}
