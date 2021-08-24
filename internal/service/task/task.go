package task

import (
	"encoding/json"

	"gitee.com/RandolphCYG/akita/internal/service/user"
	"gitee.com/RandolphCYG/akita/internal/service/wework"
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var C *cron.Cron

// JobWapper 包装塞入方法名
type JobWapper struct {
	// Id   cron.EntryID `json:"id" comment:"任务ID"`
	Name string `json:"name" comment:"任务名称"`
	Cron string `form:"cron" json:"cron" comment:"cron表达式"`
	Func func()
}

// Job 当前任务
type Job struct {
	Id cron.EntryID `json:"id" comment:"任务ID"`
}

// 重写任务执行方法
func (j JobWapper) Run() {
	j.Func()
}

// 所有任务
var AllTasks map[string]JobWapper
var CurrentTasks map[string]Job // 当前所有任务
var CacheHrUsers = user.CacheHrUsers
var CacheWeworkUsers = wework.CacheWeworkUsers
var SyncLdapUsers = user.SyncLdapUsers
var ScanExpiredLdapUsers = user.ScanExpiredLdapUsers
var ScanExpiredWeworkUsers = wework.ScanExpiredWeworkUsers
var ScanNewHrUsers = wework.ScanNewHrUsers
var CacheC7nProjects = c7n.CacheC7nProjects
var UpdateC7nUsers = c7n.UpdateC7nUsers

// crontab表达式检查 https://crontab.guru/
func init() {
	// 初始化全局定时任务对象
	C = cron.New()
	C.Start()
	AllTasks = make(map[string]JobWapper) // 注册任务
	CurrentTasks = make(map[string]Job)   // 初始化当前所有任务

	// 同步ldap用户到C7N【频繁】
	AllTasks["UpdateC7nUsers"] = JobWapper{
		Cron: "*/3 * * * *",
		Func: UpdateC7nUsers,
	}

	// 更新HR用户缓存【频繁】
	AllTasks["CacheHrUsers"] = JobWapper{
		Cron: "20 2,8,14,20 * * *",
		Func: CacheHrUsers,
	}
	// 更新企业微信用户缓存【频繁】
	AllTasks["CacheWeworkUsers"] = JobWapper{
		Cron: "10 7-22 * * *",
		Func: CacheWeworkUsers,
	}
	// 更新c7n项目缓存【频繁】
	AllTasks["CacheC7nProjects"] = JobWapper{
		Cron: "30 * * * *",
		Func: CacheC7nProjects,
	}
	// 全量为内部新用户创建企业微信账号【每天 工作时间】 依赖HR缓存和企业微信缓存
	AllTasks["ScanNewHrUsers"] = JobWapper{
		Cron: "25 9-17 * * *",
		Func: ScanNewHrUsers,
	}
	// 扫描过期ldap用户并发通知【每天一次】
	AllTasks["ScanExpiredLdapUsers"] = JobWapper{
		Cron: "10 9 * * *",
		Func: ScanExpiredLdapUsers,
	}
	// 扫描过期企业微信用户并发汇总通知【每天一次】
	AllTasks["ScanExpiredWeworkUsers"] = JobWapper{
		Cron: "00 17 * * *",
		Func: ScanExpiredWeworkUsers,
	}
	// 全量更新ldap用户信息并发汇总通知【慢 每天一次】
	AllTasks["SyncLdapUsers"] = JobWapper{
		Cron: "5 17 * * *",
		Func: SyncLdapUsers,
	}
}

// StartAll 注册并启动所有定时任务
func StartAll() {
	// 添加所有任务
	for name, t := range AllTasks {
		enterId, _ := C.AddJob(t.Cron, JobWapper{name, t.Cron, t.Func}) // cron.EntryID(len(C.Entries()) - 1),
		CurrentTasks[name] = Job{enterId}
	}
	log.Info("启动全部定时任务成功, :", CurrentTasks)
}

// TaskSart 启动定时任务 TODO bug 无法将定时任务序列化返回
func TaskSart(taskName string) serializer.Response {
	if t, ok := AllTasks[taskName]; !ok {
		return serializer.Response{Msg: "启动定时任务[" + taskName + "]失败! 未注册此任务!", Data: -1}
	} else {
		if _, exist := CurrentTasks[taskName]; exist {
			return serializer.Response{Msg: "启动定时任务[" + taskName + "]失败! 此任务已经在计划列表中，无需重复添加!", Data: -1}
		}
		enterId, _ := C.AddJob(t.Cron, JobWapper{taskName, t.Cron, t.Func}) // cron.EntryID(len(C.Entries()) - 1),
		CurrentTasks[taskName] = Job{enterId}
	}
	log.Info("动态添加定时任务，当前所有任务:", C.Entries(), "所有任务CurrentTasks:", CurrentTasks)
	res, _ := json.Marshal(CurrentTasks)
	return serializer.Response{Msg: "启动定时任务[" + taskName + "]成功!", Data: string(res)}
}

// TaskRemove 移除定时任务  TODO bug修复
func TaskRemove(taskName string) serializer.Response {
	if _, ok := CurrentTasks[taskName]; !ok {
		return serializer.Response{Msg: "移除定时任务[" + taskName + "]失败! 无此任务!", Data: -1}
	} else {
		C.Remove(CurrentTasks[taskName].Id)
		delete(CurrentTasks, taskName)
	}
	log.Info("动态移除定时任务，当前剩余任务:", CurrentTasks)
	return serializer.Response{Msg: "移除定时任务[" + taskName + "]成功!", Data: 0}
}

//  FetchTasks 查询所有任务 TODO bug 无法将定时任务序列化返回
func FetchTasks() serializer.Response {
	res, _ := json.Marshal(CurrentTasks)
	log.Info("当前所有任务CurrentTasks:", CurrentTasks)
	return serializer.Response{Data: string(res)}
}

// TaskStop 停止所有定时任务
func TaskStop() serializer.Response {
	log.Info("停止之前当前所有任务:", C.Entries())
	C.Stop()
	log.Info("停止后当前所有任务:", C.Entries())
	return serializer.Response{Msg: "停止所有定时任务成功，在执行中的不会被打断!"}
}
