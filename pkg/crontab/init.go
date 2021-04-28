package crontab

import (
	"time"

	"gitee.com/RandolphCYG/akita/pkg/log"
	"github.com/robfig/cron/v3"
)

/*
TODO 可配置性，将页面上配置的定时任务缓存到 redis
然后系统启动刷一遍所有的定时任务
*/

// Cron 定时任务
var Cron *cron.Cron

// Init 初始化定时任务
func Init() {
	log.Log().Info("初始化定时任务...")
	// 先开启秒级，写不标准crontab命令测试
	Cron := cron.New(cron.WithSeconds())
	spec1 := "*/3 * * * * *"
	Cron.AddFunc(spec1, task1)

	spec2 := "*/5 * * * * *"
	Cron.AddFunc(spec2, task2)

	defer Cron.Stop()

	go Cron.Start()
	// select {
	// case <-Cron.Stop().Done():
	// 	return
	// }
	time.Sleep(time.Second * 5)
}

// Reload 重新启动定时任务
func Reload() {
	if Cron != nil {
		Cron.Stop()
		log.Log().Warning("停止定时任务...")
	}
	Init()
}

// 测试秒级别定时任务 回头改成不支持秒级别的
func task1() {
	log.Log().Debug("每隔3秒执行一次")
}

func task2() {
	log.Log().Debug("每隔5秒执行")
}