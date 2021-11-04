package task

import (
	"encoding/json"
	"github.com/robfig/cron/v3"

	"gitee.com/RandolphCYG/akita/internal/middleware/log"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/hruser"
	"gitee.com/RandolphCYG/akita/internal/service/ldapuser"
	"gitee.com/RandolphCYG/akita/internal/service/wework"
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
)

var (
	HrCacheUsers           = hruser.CacheUsers
	WeworkCacheUsers       = wework.CacheUsers
	LdapSyncUsers          = ldapuser.SyncUsers
	LdapScanExpiredUsers   = ldapuser.ScanExpiredUsers
	WeworkScanExpiredUsers = wework.ScanExpiredUsers
	WeworkScanNewHrUsers   = wework.ScanNewHrUsers
	C7nCacheProjects       = c7n.CacheProjects
	C7nUpdateUsers         = c7n.SyncUsers
)

// crontab表达式检查 https://crontab.guru/
func init() {
	// 初始化全局定时任务对象
	model.Cron = cron.New()
	model.Cron.Start()
	model.AllTasks = make(map[string]model.JobWrapper) // 注册任务
	model.CurrentTasks = make(map[string]model.Job)    // 初始化当前所有任务

	// 同步ldap用户到C7N【频繁】
	model.AllTasks["C7nUpdateUsers"] = model.JobWrapper{
		Cron: "*/3 * * * *",
		Func: C7nUpdateUsers,
	}

	// 更新HR用户缓存【频繁】
	model.AllTasks["HrCacheUsers"] = model.JobWrapper{
		Cron: "20 2,8,14,20 * * *",
		Func: HrCacheUsers,
	}
	// 更新企业微信用户缓存【频繁】
	model.AllTasks["WeworkCacheUsers"] = model.JobWrapper{
		Cron: "10 7-22 * * *",
		Func: WeworkCacheUsers,
	}
	// 更新c7n项目缓存【频繁】
	model.AllTasks["C7nCacheProjects"] = model.JobWrapper{
		Cron: "30 * * * *",
		Func: C7nCacheProjects,
	}
	// 全量为内部新用户创建企业微信账号【每天 工作时间】 依赖HR缓存和企业微信缓存
	model.AllTasks["WeworkScanNewHrUsers"] = model.JobWrapper{
		Cron: "25 9-17 * * *",
		Func: WeworkScanNewHrUsers,
	}
	// 扫描过期ldap用户并发通知【每天一次】
	model.AllTasks["LdapScanExpiredUsers"] = model.JobWrapper{
		Cron: "10 9 * * *",
		Func: LdapScanExpiredUsers,
	}
	// 扫描过期企业微信用户并发汇总通知【每天一次】
	model.AllTasks["WeworkScanExpiredUsers"] = model.JobWrapper{
		Cron: "00 17 * * *",
		Func: WeworkScanExpiredUsers,
	}
	// 全量更新ldap用户信息并发汇总通知【慢 每天一次】
	model.AllTasks["LdapSyncUsers"] = model.JobWrapper{
		Cron: "5 17 * * *",
		Func: LdapSyncUsers,
	}
}

// Start 启动定时任务 TODO bug 无法将定时任务序列化返回
func Start(taskName string) serializer.Response {
	if t, ok := model.AllTasks[taskName]; !ok {
		return serializer.Response{Msg: "启动定时任务[" + taskName + "]失败! 未注册此任务!", Data: -1}
	} else {
		if _, exist := model.CurrentTasks[taskName]; exist {
			return serializer.Response{Msg: "启动定时任务[" + taskName + "]失败! 此任务已经在计划列表中，无需重复添加!", Data: -1}
		}
		enterId, _ := model.Cron.AddJob(t.Cron, model.JobWrapper{Name: taskName, Cron: t.Cron, Func: t.Func})
		model.CurrentTasks[taskName] = model.Job{Id: enterId}
	}
	log.Log.Info("动态添加定时任务，当前所有任务:", model.Cron.Entries(), "所有任务CurrentTasks:", model.CurrentTasks)
	res, _ := json.Marshal(model.CurrentTasks)
	return serializer.Response{Msg: "启动定时任务[" + taskName + "]成功!", Data: string(res)}
}

// Remove 移除定时任务  TODO bug修复
func Remove(taskName string) serializer.Response {
	if _, ok := model.CurrentTasks[taskName]; !ok {
		return serializer.Response{Msg: "移除定时任务[" + taskName + "]失败! 无此任务!", Data: -1}
	} else {
		model.Cron.Remove(model.CurrentTasks[taskName].Id)
		delete(model.CurrentTasks, taskName)
	}
	log.Log.Info("动态移除定时任务，当前剩余任务:", model.CurrentTasks)
	return serializer.Response{Msg: "移除定时任务[" + taskName + "]成功!", Data: 0}
}

// GetTasks 查询所有任务 TODO bug 无法将定时任务序列化返回
func GetTasks() serializer.Response {
	res, _ := json.Marshal(model.CurrentTasks)
	log.Log.Info("当前所有任务CurrentTasks:", model.CurrentTasks)
	return serializer.Response{Data: string(res)}
}

// StopAll 停止所有定时任务
func StopAll() serializer.Response {
	log.Log.Info("停止之前当前所有任务:", model.Cron.Entries())
	model.Cron.Stop()
	log.Log.Info("停止后当前所有任务:", model.Cron.Entries())
	return serializer.Response{Msg: "停止所有定时任务成功，在执行中的不会被打断!"}
}
