package model

import (
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gorm.io/gorm"
)

const (
	// Active 账户正常状态
	Active = iota
	// NotActivicated 未激活
	NotActivicated
	// Baned 被封禁
	Baned
	// OveruseBaned 超额使用被封禁
	OveruseBaned
)

// User 用户模型
type User struct {
	gorm.Model
	// ldap字段

	// 补充字段
	IsLdap    bool `json:"isLdap" gorm:"type:tinyint(3)"`    // 是否ldap
	IsEnabled bool `json:"isEnabled" gorm:"type:tinyint(3)"` // 是否被启用
	IsAdmin   bool `json:"isAdmin" gorm:"type:tinyint(3)"`   // 是否管理员
}

func GetUserByEmail(email string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("email = ?", email).First(&user)
	return user, result.Error
}

// TaskSyncHrUsers2Cache 定时任务 将用户定时更新到缓存库
func TaskSyncHrUsers2Cache() {
	log.Log().Debug("开始执行定时同步用户定时任务...")
	var hrDataConn hr.HrDataConn
	if result := DB.First(&hrDataConn); result.Error != nil {
		log.Log().Error("没有外部HR数据连接信息")
	}

	hrConn := &hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	}

	hrUsers := hr.FetchHrData(hrConn)
	data, _ := cache.Serializer(hrUsers)
	cache.Set("hrUsers", data)
}
