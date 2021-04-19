package model

import (
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
	// 表字段
	gorm.Model
	Email     string `gorm:"type:varchar(100);unique_index"`
	Nick      string `gorm:"size:50"`
	Password  string `json:"-"`
	Status    int
	GroupID   uint
	Storage   uint64
	TwoFactor string
	Avatar    string
	Options   string `json:"-",gorm:"type:text"`
	Authn     string `gorm:"type:text"`
}

func GetUserByEmail(email string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("email = ?", email).First(&user)
	return user, result.Error
}
