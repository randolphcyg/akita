package user

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/order"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"

	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
)

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

// 查询并处理过期用户逻辑 【定时任务】
func FetchAndHandleExpireUser() (handledExpireUsers []*ldap.LdapAttributes) {
	// 初始化连接
	user := &ldap.LdapAttributes{}
	LdapUsers := ldap.FetchLdapUsers(user)
	currentTime := time.Now()
	// handledExpireUsers := make([]*ldap.LdapAttributes, 0, 10)  // TODO 预留防止更改传参
	for _, user := range LdapUsers {
		expire, _ := strconv.ParseInt(user.GetAttributeValue("accountExpires"), 10, 64)
		unixTime := util.NtToUnix(expire)
		expireDays := util.FormatLdapExpireDays(util.SubDays(unixTime, currentTime))
		if expireDays != 106752 { // 排除不过期的账号
			if expireDays >= -7 && expireDays <= 14 { // 未/已经过期 7 天内的账号
				ldapUser := ldap.NewUser(user) // 初始化速度较慢 适用定时异步任务处理少量数据
				handledExpireUsers = append(handledExpireUsers, ldapUser)
				fmt.Println(ldapUser, "过期天数: ", expireDays)
				// log.Info(ldapUser, "过期天数: ", expireDays)
				// order.HandleOrderUuapExpired(ldapUser, expireDays) // 处理过期账号
				// if ldapUser.DisplayName == "蔡迎港" {
				// 	log.Info(ldapUser, "过期天数: ", expireDays)
				order.HandleOrderUuapExpired(ldapUser, expireDays) // 处理过期账号
				// }
			}
		}
	}
	return
}

// 触发定时任务
func (service *LdapUserService) FetchExpireUser() serializer.Response {
	go func() {
		gocron.Every(1).Day().At("10:00").Do(FetchAndHandleExpireUser) // 每天早上10点扫一遍
		// gocron.Every(30).Second().Do(FetchAndHandleExpireUser)
		// gocron.Every(1).Day().At("11:00").Do(FetchAndHandleExpireUser)
		_, nextTime := gocron.NextRun()
		log.Info("定时任务下次触发时间：", nextTime)
		<-gocron.Start()
	}()
	return serializer.Response{Data: 0}
}

// 创建用户-管理员接口
func (service *LdapUserService) AddUser(u LdapUserService) serializer.Response {

	return serializer.Response{Data: 0, Msg: "UUAP账号创建成功"}
}

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

// 查询HR数据
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

// 更新HR数据到LDAP实现逻辑
// 此部分逻辑将最终修改为手动和自动两种调用方式
func (service *HrDataService) HrToLdap(h HrDataService) serializer.Response {
	// 获取HR接口数据
	var hrDataConn hr.HrDataConn
	if result := model.DB.First(&hrDataConn); result.Error != nil {
		log.Error("Fail to get HR data connection cfg!")
	}

	hrConn := &hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	}

	hrUsers := hr.FetchHrData(hrConn)

	// 全量遍历HR接口数据用户 并 更新LDAP用户
	var userStat, dn string
	var expire int64

	for _, user := range hrUsers {
		if user.Stat == "离职" { //离职员工
			userStat = "546"
			dn = "OU=disabled," + bootstrap.LdapCfg.BaseDn
			expire = 0 // 账号失效
		} else { // 在职员工
			userStat = "544"
			dn = ldap.DepartToDn(user.Department)
			expire = ldap.ExpireTime(-1) // 账号永久有效
		}
		depart := strings.Split(user.Department, ".")[len(strings.Split(user.Department, "."))-1]
		name := []rune(user.Name)

		// 将hr数据转换为ldap信息格式
		user := &ldap.LdapAttributes{
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
		user.Update()
	}
	return serializer.Response{Data: 1, Msg: "更新成功"}
}
