package bootstrap

import (
	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/email"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	log "github.com/sirupsen/logrus"

	"gitee.com/RandolphCYG/akita/internal/model"
)

var (
	LdapCfg   model.LdapCfg
	LdapField model.LdapField
)

// Init 初始化启动
func Init(cfg string) {
	// 初始化应用
	InitApplication()
	// 初始化系统配置 TODO 各配置热加载
	c, err := conf.Init(cfg)
	if err != nil {
		panic(err)
	}
	// 初始化 db
	model.Init(&c.Database)
	// 数据迁移
	log.Info("Data migration begin ...")
	model.DB.AutoMigrate(&model.LdapCfg{}, &model.LdapField{}, &hr.HrDataConn{}, &model.WeworkCfg{}, &model.WeworkOrder{})
	if result := model.DB.Limit(1).Find(&model.LdapCfg{}); result.RowsAffected == 0 {
		model.DB.Create(&c.LdapCfg)
	}
	log.Info("Data migration end...")

	// 初始化 redis
	cache.Init(&c.Redis)

	// 初始化 email
	email.Init(&c.Email)

	// 初始化ldap连接
	LdapCfg, _ = model.GetAllLdapConn() // 直接查询
	// 初始化ldap字段配置
	LdapField, _ = model.GetLdapFieldByConnUrl(LdapCfg.ConnUrl)

	// 初始化企业微信配置信息
	err = model.GetWeworkOrderCfg()
	if err != nil {
		log.Error("初始化企业微信审批配置信息错误,err: ", err)
	}
	err = model.GetWeworkUuapCfg()
	if err != nil {
		log.Error("初始化企业微信UUAP公告应用配置信息错误,err: ", err)
	}
}
