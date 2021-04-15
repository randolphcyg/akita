package conf

import (
	"gitee.com/RandolphCYG/akita/pkg/database/orm"
	"gitee.com/RandolphCYG/akita/pkg/email"
	"gitee.com/RandolphCYG/akita/pkg/redis"
)

// Config 全局配置文件 结构体的名称对应yaml文件中各配置的平台
type Config struct {
	System   System
	Database orm.Config
	Redis    redis.Config
	Email    email.Config
}

// System 系统配置
type System struct {
	Addr  string
	Mode  string
	Debug bool
}
