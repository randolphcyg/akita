package conf

import (
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/db"
	"gitee.com/RandolphCYG/akita/pkg/email"
)

// Config 全局配置文件 结构体的名称对应yaml文件中各配置的平台
type Config struct {
	System   System
	Database db.Config
	Redis    cache.Config
	LdapCfg  model.LdapCfg
	Email    email.Config
}

// System 系统配置
type System struct {
	Addr  string
	Mode  string
	Debug bool
}
