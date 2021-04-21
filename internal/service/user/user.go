package user

import (
	"fmt"

	ldapService "gitee.com/RandolphCYG/akita/internal/service/ldap"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
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
	var ldapService ldapService.LdapConnService
	fmt.Println(url)
	conn, err := ldapService.FetchByConnUrl(url)
	if err != nil {
		fmt.Println(err)
	}
	LdapUsers := ldap.FetchLdapUsers(&conn)
	for _, user := range LdapUsers {
		user.PrettyPrint(2)
		fmt.Println(user.GetAttributeValue("displayName"))
		break
	}
	return serializer.Response{Data: LdapUsers}
}

// 创建用户
func (service *LdapUserService) AddUser(u LdapUserService) serializer.Response {
	// service层处理前端数据，并将数据传给pkg的ldap组件，然后ldap组件处理有关ldap用户的通用逻辑

	return serializer.Response{Data: 111}
}
