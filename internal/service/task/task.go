package task

import (
	"gitee.com/RandolphCYG/akita/internal/service/user"
	"gitee.com/RandolphCYG/akita/pkg/crontab"
	"gitee.com/RandolphCYG/akita/pkg/serializer"

	"github.com/sirupsen/logrus"
)

type Task struct {
	Tags []string `json:"tags" comment:"标签"`
	Cron string   `form:"cron" json:"cron" comment:"cron表达式"`
	Func func()
}

// 所有任务
var AllTasks map[string]Task

// LdapUsersCronTasksStart 用户相关任务的注册
func LdapUsersCronTasksStart() serializer.Response {
	var CacheToLdap = user.CacheToLdap
	var HandleExpiredLdapUsers = user.HandleExpiredLdapUsers

	_, _ = crontab.Scheduler.Every(1).Day().Tag("CacheToLdap").At("10:30").Do(CacheToLdap) // 每天早10点半
	crontab.Scheduler.StartAsync()
	_, t1 := crontab.Scheduler.NextRun()
	logrus.Info("下次执行更新全量LDAP用户信息的触发时间：", t1.Format("2006-01-02 15:04:05.000"))

	_, _ = crontab.Scheduler.Every(1).Day().Tag("HandleExpiredLdapUsers").At("10:00").Do(HandleExpiredLdapUsers)
	crontab.Scheduler.StartAsync()
	_, t2 := crontab.Scheduler.NextRun()
	logrus.Info("扫描过期用户的下次触发时间：", t2.Format("2006-01-02 15:04:05.000"))

	for _, j := range crontab.Scheduler.Jobs() {
		crontab.JobsInfos = append(crontab.JobsInfos, crontab.JobsInfo{
			Tags:    j.Tags(),
			NextRun: j.NextRun(),
		})
	}

	return serializer.Response{Data: crontab.JobsInfos}
}

// TaskRegister 全局简易任务注册方法
// crontab表达式检查 https://crontab.guru/
func TaskRegister(taskName string) serializer.Response {
	// 初始化
	AllTasks = make(map[string]Task)
	// 注册任务
	AllTasks["HandleExpiredLdapUsers"] = Task{
		Tags: []string{"HandleExpiredLdapUsers"},
		Cron: "0 10 * * *",
		Func: user.TestTask,
	}

	AllTasks["CacheToLdap"] = Task{
		Tags: []string{"CacheToLdap"},
		Cron: "0 10 * * *",
		Func: user.TestTask,
	}

	// 测试
	AllTasks["Test"] = Task{
		Tags: []string{"test", "Test"},
		Cron: "30 10 * * *",
		Func: user.TestTask,
	}

	if t, ok := AllTasks[taskName]; !ok {
		logrus.Error("Task Not Found !!!")
	} else {
		_, err := crontab.Scheduler.Cron(t.Cron).Tag(taskName).Do(t.Func)
		if err != nil {
			logrus.Error(err)
		}
		crontab.Scheduler.StartAsync()
		crontab.FreshCurrentJobs()
		logrus.Info("当前所有的定时任务:", crontab.JobsInfos)
		_, t := crontab.Scheduler.NextRun()
		logrus.Info("下次执行任务[", taskName, "]的触发时间：", t.Format("2006-01-02 15:04:05.000"))
	}

	return serializer.Response{Data: 0}
}
