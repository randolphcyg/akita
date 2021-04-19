package ldap

import (
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gorm.io/gorm"
)

// LdapConnService 查询条件
type LdapConnService struct {
	gorm.Model
	// 连接地址
	ConnUrl string `json:"conn_url" gorm:"type:varchar(255);unique_index"`
	// SSL加密方式
	SslEncryption bool `json:"ssl_encryption" gorm:"type:tinyint;length:1"`
	// 超时设置
	Timeout int `json:"timeout" gorm:"type:int"`
	// 根目录
	BaseDn string `json:"base_dn" gorm:"type:varchar(255)"`
	// 用户名
	AdminAccount string `json:"admin_account" gorm:"type:varchar(255)"`
	// 密码
	Password string `json:"password" gorm:"type:varchar(255)"`
}

// 增
func (service *LdapConnService) Add(s *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.AdminAccount = s.AdminAccount
	conn.BaseDn = s.BaseDn
	conn.ConnUrl = s.ConnUrl
	conn.Password = s.Password
	conn.SslEncryption = s.SslEncryption
	conn.Timeout = s.Timeout

	if err := model.DB.Create(&conn).Error; err != nil {
		return serializer.DBErr("增加记录失败", err)
	} else {
		return serializer.Response{Data: conn, Msg: "增加成功!"}
	}
}

// 删
func (service *LdapConnService) Delete(s *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.ID = s.ID
	if err := model.DB.Delete(&conn).Error; err != nil {
		return serializer.DBErr("删除记录失败", err)
	} else {
		return serializer.Response{Data: conn, Msg: "删除成功!"}
	}
}

// 改
func (service *LdapConnService) Update(s *LdapConnService) serializer.Response {
	conn := model.NewLdapConn()
	conn.ID = s.ID
	conn.AdminAccount = s.AdminAccount
	conn.BaseDn = s.BaseDn
	conn.ConnUrl = s.ConnUrl
	conn.Password = s.Password
	conn.SslEncryption = s.SslEncryption
	conn.Timeout = s.Timeout

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
