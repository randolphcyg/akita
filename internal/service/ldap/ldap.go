package ldap

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gorm.io/gorm"
)

// LdapConnService 查询条件
type LdapConnService struct {
	gorm.Model
	// 连接地址
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index;not null;comment:连接地址 逻辑外键"`
	// SSL加密方式
	SslEncryption bool `json:"ssl_encryption" gorm:"type:tinyint;length:1;comment:SSL加密方式"`
	// 超时设置
	Timeout int `json:"timeout" gorm:"type:int;comment:超时设置"`
	// 根目录
	BaseDn string `json:"base_dn" gorm:"type:varchar(255);not null;comment:根目录"`
	// 用户名
	AdminAccount string `json:"admin_account" gorm:"type:varchar(255);not null;comment:用户名"`
	// 密码
	Password string `json:"password" gorm:"type:varchar(255);not null;comment:密码"`
}

// 增
func (service *LdapConnService) Add(c *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.AdminAccount = c.AdminAccount
	conn.BaseDn = c.BaseDn
	conn.ConnUrl = c.ConnUrl
	conn.Password = c.Password
	conn.SslEncryption = c.SslEncryption
	conn.Timeout = c.Timeout

	if err := model.DB.Create(&conn).Error; err != nil {
		return serializer.DBErr("增加记录失败", err)
	} else {
		return serializer.Response{Data: conn, Msg: "增加成功!"}
	}
}

// 删
func (service *LdapConnService) Delete(c *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.ID = c.ID
	if err := model.DB.Delete(&conn).Error; err != nil {
		return serializer.DBErr("删除记录失败", err)
	} else {
		return serializer.Response{Data: conn, Msg: "删除成功!"}
	}
}

// 改
func (service *LdapConnService) Update(c *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.ID = c.ID
	conn.AdminAccount = c.AdminAccount
	conn.BaseDn = c.BaseDn
	conn.ConnUrl = c.ConnUrl
	conn.Password = c.Password
	conn.SslEncryption = c.SslEncryption
	conn.Timeout = c.Timeout

	if err := model.DB.Save(&conn).Error; err != nil {
		return serializer.DBErr("修改记录失败", err)
	} else {
		return serializer.Response{Data: conn, Msg: "修改成功!"}
	}
}

// 查
func (service *LdapConnService) Fetch() serializer.Response {
	conn, err := model.GetAllLdapConn()
	if err != nil {
		return serializer.DBErr("不存在任何ldap连接信息", err)
	} else {
		return serializer.Response{Data: conn, Msg: "查询成功!"}
	}

}

// 测试
func (service *LdapConnService) Test(id uint) serializer.Response {
	conn, err := model.GetLdapConn(id)
	fmt.Println(conn)
	if err != nil {
		return serializer.DBErr("不存在任何ldap连接信息", err)
	}

	_, err = ldap.NewLdapConn(&conn)
	if err != nil {
		return serializer.Err(-1, "ldap连接出错", err)
	}
	return serializer.Response{Msg: "ldap连接测试成功!"}
}

// LdapFieldService LDAP字段查询条件
type LdapFieldService struct {
	gorm.Model

	// 连接地址 逻辑外键
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index;not null;comment:连接地址 逻辑外键"`

	// 用户基础字段
	// 拉取结点 默认不填表示从根目录开始搜索
	BasicPullNode string `json:"basic_pull_node" gorm:"type:varchar(255);comment:拉取结点 默认不填表示从根目录开始搜索"`
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

	// 用户拓展字段
	// 组织过滤规则
	SearchFilterOu string `json:"search_filter_ou" gorm:"type:varchar(255);;comment:组织过滤规则"`
	// 禁用用户DN
	BaseDnDisabled string `json:"base_dn_disabled" gorm:"type:varchar(255);;comment:禁用用户DN"`
	// 公司英文前缀
	CompanyPrefixs []string `json:"company_prefixs" gorm:"type:varchar(255);;comment:公司英文前缀"`

	// 用户组字段
	// 用户组对象类
	UserGroupClass string `json:"user_group_class" gorm:"type:varchar(255);not null;comment:用户组对象类"`
	// 用户组对象过滤
	UserGroupFilter string `json:"user_group_filter" gorm:"type:varchar(255);not null;comment:用户组对象过滤"`
	// 用户组名
	UserGroupName string `json:"user_group_name" gorm:"type:varchar(255);not null;comment:用户组名"`
	// 用户组描述
	UserGroupDescription string `json:"user_group_description" gorm:"type:varchar(255);not null;comment:用户组描述"`
}

// 增
func (service *LdapFieldService) AddField(f *LdapFieldService) serializer.Response {
	field := model.NewLdapField()
	field.BaseDnDisabled = f.BaseDnDisabled
	field.BasicPullNode = f.BasicPullNode
	field.CompanyPrefixs = f.CompanyPrefixs
	field.ConnUrl = f.ConnUrl // 这里的连接地址由前端传回 然后后端校验ldap_conns表确实有这个记录在进行增操作
	field.DisplayName = f.DisplayName
	field.Email = f.Email
	field.Mobile = f.Mobile
	field.OrganizationClass = f.OrganizationClass
	field.SearchFilterOu = f.SearchFilterOu
	field.UserClass = f.UserClass
	field.UserFilter = f.UserFilter
	field.UserGroupClass = f.UserGroupClass
	field.UserGroupDescription = f.UserGroupDescription
	field.UserGroupFilter = f.UserGroupFilter
	field.UserGroupName = f.UserGroupName
	field.Username = f.Username

	if err := model.DB.Create(&field).Error; err != nil {
		return serializer.DBErr("增加字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "增加字段明细成功!"}
	}
}

// 删
func (service *LdapFieldService) DeleteField(f *LdapFieldService) serializer.Response {
	field := model.NewLdapField()
	field.ID = f.ID
	if err := model.DB.Delete(&field).Error; err != nil {
		return serializer.DBErr("删除字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "删除字段明细成功!"}
	}
}

// 改
func (service *LdapFieldService) UpdateField(f *LdapFieldService) serializer.Response {
	field := model.NewLdapField()
	field.ID = f.ID
	field.BaseDnDisabled = f.BaseDnDisabled
	field.BasicPullNode = f.BasicPullNode
	field.CompanyPrefixs = f.CompanyPrefixs
	field.ConnUrl = f.ConnUrl // 这里的连接地址由前端传回 然后后端校验ldap_conns表确实有这个记录在进行增操作
	field.DisplayName = f.DisplayName
	field.Email = f.Email
	field.Mobile = f.Mobile
	field.OrganizationClass = f.OrganizationClass
	field.SearchFilterOu = f.SearchFilterOu
	field.UserClass = f.UserClass
	field.UserFilter = f.UserFilter
	field.UserGroupClass = f.UserGroupClass
	field.UserGroupDescription = f.UserGroupDescription
	field.UserGroupFilter = f.UserGroupFilter
	field.UserGroupName = f.UserGroupName
	field.Username = f.Username

	if err := model.DB.Save(&field).Error; err != nil {
		return serializer.DBErr("修改字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "修改字段明细成功!"}
	}
}
