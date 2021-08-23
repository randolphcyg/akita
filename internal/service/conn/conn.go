package conn

import (
	"encoding/json"
	"time"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/ldap"
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
	Timeout time.Duration `json:"timeout" gorm:"type:int;comment:超时设置"`
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
	err1 := model.DB.Where("conn_url = ?", c.ConnUrl).Delete(&conn).Error // 删除LDAP连接记录

	field := model.NewLdapField()
	err2 := model.DB.Where("conn_url = ?", c.ConnUrl).Delete(&field).Error // 删除LDAP连接对应的字段明细记录

	if err1 != nil {
		return serializer.DBErr("删除连接记录失败", err1)
	} else if err2 != nil {
		return serializer.DBErr("删除连接对应明细记录失败", err2)
	} else {
		return serializer.Response{Data: c.ConnUrl, Msg: "删除连接成功!"}
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

// 查 根据conn_url查询ldap连接信息
func (service *LdapConnService) FetchByConnUrl(url string) (conn model.LdapCfg, err error) {
	conn, err = model.GetLdapConnByConnUrl(url)
	if err != nil {
		return
	} else {
		return conn, nil
	}
}

// 测试
func (service *LdapConnService) Test(id uint) serializer.Response {
	conn, err := model.GetLdapConn(id)
	if err != nil {
		return serializer.DBErr("不存在任何ldap连接信息", err)
	}

	err = ldap.Init(&conn)
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
	SearchFilterOu string `json:"search_filter_ou" gorm:"type:varchar(255);comment:组织过滤规则"`
	// 禁用用户DN
	BaseDnDisabled string `json:"base_dn_disabled" gorm:"type:varchar(255);comment:禁用用户DN"`
	// 公司英文前缀
	BaseDnOuter        string                       `json:"base_dn_outer" gorm:"type:varchar(255);comment:外部用户DN"`
	BaseDnToBeAssigned string                       `json:"base_dn_to_be_assigned" gorm:"type:varchar(255);comment:公司内部待分配DN"`
	CompanyType        string                       `json:"company_type" gorm:"type:varchar(255);comment:公司类型"`
	CompanyTypes       map[string]model.CompanyType `json:"company_types" gorm:"-"` // 非数据库字段 用来处理复杂数据结构

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

// 序列化 公司类型 切片转字符串存储
func CompanyTypes2Str(m map[string]model.CompanyType) (string, error) {
	bs, err := json.Marshal(m)
	return string(bs), err
}

// 反序列化 公司类型 字符串转切片使用
func Str2CompanyTypes(m string) (companyType map[string]model.CompanyType, err error) {
	err = json.Unmarshal([]byte(m), &companyType)
	return companyType, err
}

// 新增
func (service *LdapFieldService) AddField(f *LdapFieldService) serializer.Response {
	companyTypesStr, err := CompanyTypes2Str(f.CompanyTypes)
	if err != nil {
		return serializer.Err(-1, "序列化错误", err)
	}

	field := &model.LdapField{
		BaseDnDisabled:       f.BaseDnDisabled,
		BasicPullNode:        f.BasicPullNode,
		ConnUrl:              f.ConnUrl,
		DisplayName:          f.DisplayName,
		Email:                f.Email,
		Mobile:               f.Mobile,
		OrganizationClass:    f.OrganizationClass,
		SearchFilterOu:       f.SearchFilterOu,
		UserClass:            f.UserClass,
		UserFilter:           f.UserFilter,
		UserGroupClass:       f.UserGroupClass,
		UserGroupDescription: f.UserGroupDescription,
		UserGroupFilter:      f.UserGroupFilter,
		UserGroupName:        f.UserGroupName,
		Username:             f.Username,
		BaseDnOuter:          f.BaseDnOuter,
		BaseDnToBeAssigned:   f.BaseDnToBeAssigned,
		CompanyType:          companyTypesStr,
	}

	if err := model.DB.Create(field).Error; err != nil {
		return serializer.DBErr("增加字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "增加字段明细成功!"}
	}
}

// 删
func (service *LdapFieldService) DeleteField(url string) serializer.Response {
	field := model.NewLdapField()
	if err := model.DB.Where("conn_url = ?", url).Delete(&field).Error; err != nil {
		return serializer.DBErr("删除字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "删除字段明细成功!"}
	}
}

// 改
func (service *LdapFieldService) UpdateField(f *LdapFieldService) serializer.Response {
	companyTypesStr, err := CompanyTypes2Str(f.CompanyTypes)
	if err != nil {
		return serializer.Err(-1, "序列化错误", err)
	}
	field := &model.LdapField{
		BaseDnDisabled:       f.BaseDnDisabled,
		BasicPullNode:        f.BasicPullNode,
		ConnUrl:              f.ConnUrl,
		DisplayName:          f.DisplayName,
		Email:                f.Email,
		Mobile:               f.Mobile,
		OrganizationClass:    f.OrganizationClass,
		SearchFilterOu:       f.SearchFilterOu,
		UserClass:            f.UserClass,
		UserFilter:           f.UserFilter,
		UserGroupClass:       f.UserGroupClass,
		UserGroupDescription: f.UserGroupDescription,
		UserGroupFilter:      f.UserGroupFilter,
		UserGroupName:        f.UserGroupName,
		Username:             f.Username,
		BaseDnOuter:          f.BaseDnOuter,
		BaseDnToBeAssigned:   f.BaseDnToBeAssigned,
		CompanyType:          companyTypesStr,
	}

	if err := model.DB.Where("conn_url = ?", f.ConnUrl).Save(&field).Error; err != nil {
		return serializer.DBErr("修改字段明细记录失败", err)
	} else {
		return serializer.Response{Data: field, Msg: "修改字段明细成功!"}
	}
}

// 查
func (service *LdapFieldService) FetchField(url string) serializer.Response {
	var field model.LdapField
	res := model.DB.Where("conn_url = ?", url).First(&field)
	if res.Error != nil {
		return serializer.DBErr("反序列化失败", res.Error)
	}

	companyTypes, err := Str2CompanyTypes(field.CompanyType)
	if err != nil {
		return serializer.Err(-2, "反序列化失败", err)
	}
	field.CompanyTypes = companyTypes

	if err != nil {
		return serializer.DBErr("不存在任何ldap连接的字段明细信息!", err)
	} else {
		return serializer.Response{Data: field, Msg: "查询成功!"}
	}
}
