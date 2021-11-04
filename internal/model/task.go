package model

import (
	"github.com/robfig/cron/v3"

	"gitee.com/RandolphCYG/akita/internal/middleware/log"
)

var (
	Cron         *cron.Cron
	AllTasks     map[string]JobWrapper
	CurrentTasks map[string]Job // 当前所有任务
)

// JobWrapper 包装塞入方法名
type JobWrapper struct {
	// Id   cron.EntryID `json:"id" comment:"任务ID"`
	Name string `json:"name" comment:"任务名称"`
	Cron string `form:"cron" json:"cron" comment:"cron表达式"`
	Func func()
}

// Job 当前任务
type Job struct {
	Id cron.EntryID `json:"id" comment:"任务ID"`
}

// Run 重写任务执行方法
func (j JobWrapper) Run() {
	j.Func()
}

// InitTasks 注册并启动所有定时任务
func InitTasks() {
	for name, t := range AllTasks {
		enterId, _ := Cron.AddJob(t.Cron, JobWrapper{Name: name, Cron: t.Cron, Func: t.Func})
		CurrentTasks[name] = Job{Id: enterId}
	}
	log.Log.Info("Success to start all crontab tasks, :", CurrentTasks)
}
