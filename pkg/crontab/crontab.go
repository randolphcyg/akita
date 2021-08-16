package crontab

import (
	"github.com/robfig/cron/v3"
)

// 全局变量
var C *cron.Cron

// 初始化全局变量
func Init() {
	C = cron.New()
	C.Start()
}
