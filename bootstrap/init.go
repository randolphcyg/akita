package bootstrap

import (
	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/log"

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
	model.DB.AutoMigrate(&model.LdapCfg{}, &model.LdapField{}, &hr.HrDataConn{})
	if result := model.DB.Limit(1).Find(&model.LdapCfg{}); result.RowsAffected == 0 {
		log.Log().Info("#######数据迁移...")
		model.DB.Create(&c.LdapCfg)
	}

	// 初始化 redis
	log.Log().Info("#######初始化缓存库库:%v", &c.Redis)
	cache.Init(&c.Redis)

	// 初始化ldap连接
	log.Log().Info("#######初始化ldap连接...")
	Conn, _ := model.GetAllLdapConn() // 直接查询
	ldap.Init(&Conn)
}
