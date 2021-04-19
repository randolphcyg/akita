package bootstrap

import (
	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/redis"

	"gitee.com/RandolphCYG/akita/internal/model"
)

// Init 初始化启动
func Init(cfg string) {
	// 初始化应用 等出releases版本再写版本检查
	InitApplication()
	// 初始化系统配置
	c, err := conf.Init(cfg)
	if err != nil {
		panic(err)
	}
	log.Log().Info("#######后端:%v", &c.System)
	// 初始化 db
	log.Log().Info("#######初始化数据库:%v", &c.Database)
	model.Init(&c.Database)
	// 数据迁移
	model.DB.AutoMigrate(&model.LdapConn{})
	if err := model.DB.First(&model.LdapConn{}).Error; err != nil {
		model.DB.Create(&c.LdapConn)
	}

	// 初始化 redis
	log.Log().Info("#######初始化缓存库库:%v", &c.Redis)
	redis.Init(&c.Redis)
	log.Log().Info("#######初始化其他组件...")
}
