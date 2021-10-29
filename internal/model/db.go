package model

import (
	"gitee.com/RandolphCYG/akita/pkg/db"
	"gorm.io/gorm"
)

// DB 数据库全局变量
var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(cfg *db.Config) *gorm.DB {
	DB = db.NewMySQL(cfg)
	return DB
}

// GetDB 返回默认的数据库
func GetDB() *gorm.DB {
	return DB
}
