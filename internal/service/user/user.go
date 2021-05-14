package user

import (
	"fmt"
	"strings"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
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
	for _, user := range LdapUsers {
		fmt.Println(user.DN)
		break
	}
	return serializer.Response{Data: LdapUsers}
}

// 创建用户
func (service *LdapUserService) AddUser(u LdapUserService) serializer.Response {
	// service层处理前端数据，并将数据传给pkg的ldap组件，然后ldap组件处理有关ldap用户的通用逻辑

	return serializer.Response{Data: 111}
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
		log.Log().Error("没有外部HR数据连接信息")
	}

	hrConn := &hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	}

	hrUsers := hr.FetchHrData(hrConn)

	// 全量遍历HR接口数据用户 并 更新LDAP用户
	for _, user := range hrUsers {
		if strings.EqualFold(user.Name, "张三") { // 暂时测试一位用户
			// 取ldap连接
			var userStat string
			if user.Stat == "在职" {
				userStat = "544"
			} else {
				userStat = "546"
			}
			depart := strings.Split(user.Department, ".")[len(strings.Split(user.Department, "."))-1]
			user := &ldap.LdapAttributes{
				Num:         user.Eid,
				Sam:         user.Eid,
				DisplayName: user.Name,
				Email:       user.Mail,
				Phone:       user.Mobile,
				Dn:          ldap.DepartToDn(user.Department),
				AccountCtl:  userStat,
				Expire:      ldap.ExpireTime(int64(30)),
				Sn:          user.Name,
				Name:        user.Name,
				GivenName:   user.Name,
				Company:     user.CompanyName,
				Depart:      depart,
				Title:       user.Title,
			}
			user.ModifyInfo() // 修改用户信息
			// fmt.Println(addRes)
		}
	}
	return serializer.Response{Data: 1111, Msg: "更新成功"}
}
