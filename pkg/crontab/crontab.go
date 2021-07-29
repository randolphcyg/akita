package crontab

import (
	"fmt"
	"time"

	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

// 全局
var Scheduler *gocron.Scheduler
var JobsInfos []JobsInfo

type JobsInfo struct {
	Tags    []string `json:"tags"`
	NextRun time.Time
}

// 初始化全局变量
func Init() {
	Scheduler = gocron.NewScheduler(time.Local)
	Scheduler.TagsUnique()
}

// TaskSart 启动定时任务
func TaskSart(taskName string) serializer.Response {
	FreshCurrentJobs()
	logrus.Info("当前所有的定时任务:", JobsInfos)
	logrus.Info("启动定时任务:", taskName)
	err := Scheduler.RunByTag(taskName)
	if err != nil {
		fmt.Println(err)
	}
	FreshCurrentJobs()
	logrus.Info("当前所有的定时任务:", JobsInfos)
	return serializer.Response{Data: err}
}

// TaskStop 停止定时任务
func TaskStop(taskName string) serializer.Response {
	FreshCurrentJobs()
	logrus.Info("当前所有的定时任务:", JobsInfos)
	logrus.Info("停止定时任务:", taskName)
	Scheduler.RemoveByTag(taskName)
	FreshCurrentJobs()
	logrus.Info("当前所有的定时任务:", JobsInfos)
	return serializer.Response{Data: 0}
}

// 刷新现有的任务
func FreshCurrentJobs() {
	// 清空切片
	JobsInfos = JobsInfos[0:0]
	// 每次调用crontab包都将正在执行的任务刷到全局变量中
	for _, j := range Scheduler.Jobs() {
		JobsInfos = append(JobsInfos, JobsInfo{
			Tags:    j.Tags(),
			NextRun: j.NextRun(),
		})
	}
}
