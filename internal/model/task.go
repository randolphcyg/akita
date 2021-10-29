package model

import (
	"gitee.com/RandolphCYG/akita/pkg/log"
	"github.com/robfig/cron/v3"
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
	// 添加所有任务
	for name, t := range AllTasks {
		enterId, _ := Cron.AddJob(t.Cron, JobWrapper{name, t.Cron, t.Func}) // cron.EntryID(len(C.Entries()) - 1),
		CurrentTasks[name] = Job{enterId}
	}
	log.Log.Info("Success to start all crontab tasks, :", CurrentTasks)
}
