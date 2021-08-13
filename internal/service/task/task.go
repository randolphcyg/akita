package task

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/service/user"
	"gitee.com/RandolphCYG/akita/internal/service/wework"
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
var CacheHrUsers = user.CacheHrUsers
var CacheWeworkUsers = wework.CacheWeworkUsers
var SyncLdapUsers = user.SyncLdapUsers
var ScanExpiredLdapUsers = user.ScanExpiredLdapUsers
var ScanExpiredWeworkUsers = wework.ScanExpiredWeworkUsers
var ScanNewHrUsers = wework.ScanNewHrUsers

// StartAll 注册并启动所有定时任务 TODO 需要全部修正为cron并修正错误
func StartAll() serializer.Response {
	// 更新HR用户缓存【频繁】
	// _, _ = crontab.Scheduler.Every(1).Day().Tag("CacheHrUsers").At("9:00").Do(CacheHrUsers)
	// _, _ = crontab.Scheduler.Every("3m").Tag("CacheHrUsers").Do(CacheHrUsers)
	_, _ = crontab.Scheduler.Cron("0 9,13,21 * * *").Tag("CacheHrUsers").Do(CacheHrUsers)
	crontab.Scheduler.StartAsync()

	// 更新企业微信用户缓存【频繁】
	// _, _ = crontab.Scheduler.Every(1).Day().Tag("CacheWeworkUsers").At("9:00").Do(CacheWeworkUsers)
	// _, _ = crontab.Scheduler.Every("2m").Tag("CacheWeworkUsers").Do(CacheWeworkUsers)
	_, _ = crontab.Scheduler.Cron("5 9,13,21 * * *").Tag("CacheWeworkUsers").Do(CacheWeworkUsers)
	crontab.Scheduler.StartAsync()

	// 全量为内部新用户创建企业微信账号【每天 多次】
	_, _ = crontab.Scheduler.Every(1).Day().Tag("ScanNewHrUsers").At("9:30").Do(ScanNewHrUsers)
	// _, _ = crontab.Scheduler.Cron("10 9,13,21 * * *").Tag("ScanNewHrUsers").Do(ScanNewHrUsers)
	crontab.Scheduler.StartAsync()

	// 扫描过期企业微信用户并发汇总通知【每天】
	_, _ = crontab.Scheduler.Every(1).Day().Tag("ScanExpiredWeworkUsers").At("17:10").Do(ScanExpiredWeworkUsers)
	// _, _ = crontab.Scheduler.Cron("15 9 * * *").Tag("ScanExpiredWeworkUsers").Do(ScanExpiredWeworkUsers)
	crontab.Scheduler.StartAsync()

	// 扫描过期ldap用户并发通知【每天】
	_, _ = crontab.Scheduler.Every(1).Day().Tag("ScanExpiredLdapUsers").At("10:00").Do(ScanExpiredLdapUsers)
	// _, _ = crontab.Scheduler.Cron("25 9 * * *").Tag("ScanExpiredLdapUsers").Do(ScanExpiredLdapUsers)
	crontab.Scheduler.StartAsync()

	// 全量更新ldap用户信息并发汇总通知【慢 每天】
	_, _ = crontab.Scheduler.Every(1).Day().Tag("SyncLdapUsers").At("17:15").Do(SyncLdapUsers)
	// _, _ = crontab.Scheduler.Cron("0 17 * * *").Tag("SyncLdapUsers").Do(SyncLdapUsers)
	crontab.Scheduler.StartAsync()

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
func TaskRegister(taskName string) {
	// 初始化
	AllTasks = make(map[string]Task)
	// 注册任务
	AllTasks["CacheHrUsers"] = Task{
		Tags: []string{"CacheHrUsers"},
		Cron: "50 9 * * *",
		Func: CacheHrUsers,
	}

	AllTasks["CacheWeworkUsers"] = Task{
		Tags: []string{"CacheWeworkUsers"},
		Cron: "50 9 * * *",
		Func: CacheWeworkUsers,
	}

	AllTasks["ScanExpiredWeworkUsers"] = Task{
		Tags: []string{"ScanExpiredWeworkUsers"},
		Cron: "55 9 * * *",
		Func: CacheWeworkUsers,
	}

	AllTasks["ScanExpiredLdapUsers"] = Task{
		Tags: []string{"ScanExpiredLdapUsers"},
		Cron: "0 10 * * *",
		Func: ScanExpiredLdapUsers,
	}

	AllTasks["SyncLdapUsers"] = Task{
		Tags: []string{"SyncLdapUsers"},
		Cron: "0 17 * * *",
		Func: SyncLdapUsers,
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
		// TODO 任务会立即执行的bug
		fmt.Println("TODO 任务会立即执行的bug")
		_, err := crontab.Scheduler.Cron(t.Cron).Tag(taskName).Do(t.Func)
		if err != nil {
			logrus.Error(err)
		}
		_, t := crontab.Scheduler.NextRun()
		logrus.Info("下次执行任务[", taskName, "]的触发时间：", t.Format("2006-01-02 15:04:05.000"))
		crontab.Scheduler.StartAsync()
		crontab.FreshCurrentJobs()
		logrus.Info("当前所有的定时任务:", crontab.JobsInfos)
	}
}

func FetchCurrentJobs() (Jobs []crontab.JobsInfo) {
	crontab.FreshCurrentJobs()
	Jobs = crontab.JobsInfos
	return
}
