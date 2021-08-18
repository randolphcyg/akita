package router

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/internal/handler"
	"gitee.com/RandolphCYG/akita/pkg/logger"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	switch conf.Conf.System.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	return router()
}

// InitDebugRouter debug模式路由
func router() *gin.Engine {
	fmt.Println(gin.Mode())
	r := gin.New()
	r.Use(logger.LogerMiddleware())
	v1 := r.Group("/api/v1")

	{
		// 全局设置
		site := v1.Group("site")
		site.GET("ping", handler.Ping) // 存活探针

		// ldap连接相关路由
		ldapConns := v1.Group("ldap/conns")
		ldapConns.GET("fetch", handler.FetchLdapConn)
		ldapConns.POST("create", handler.CreateLdapConn)
		ldapConns.POST("update", handler.UpdateLdapConn)
		ldapConns.DELETE("delete", handler.DeleteLdapConn)
		ldapConns.POST("test", handler.TestLdapConn)

		// ldap连接的字段明细配置
		ldapFields := v1.Group("ldap/fields")
		ldapFields.GET("fetch", handler.FetchLdapField)
		ldapFields.POST("create", handler.CreateLdapField)
		ldapFields.POST("update", handler.UpdateLdapField)
		ldapFields.DELETE("delete", handler.DeleteLdapField)
		ldapFields.POST("test", handler.TestLdapField)

		// ldap用户
		ldapUsers := v1.Group("ldap/users")
		ldapUsers.GET("manual/cache/hr", handler.CacheHrUsersManual)            // 手动触发缓存HR用户
		ldapUsers.GET("manual/sync", handler.SyncLdapUsersManual)               // 手动触发更新ldap用户
		ldapUsers.GET("manual/scan/expire", handler.ScanExpiredLdapUsersManual) // 手动触发扫描过期ldap用户

		// 企业微信工单
		weworkOrders := v1.Group("wework/orders")
		weworkOrders.POST("handle", handler.HandleOrders)

		// 企业微信用户
		weworkUsers := v1.Group("wework/users")
		weworkUsers.GET("manual/cache", handler.CacheUsersManual)                   // 手动触发缓存企业微信用户
		weworkUsers.GET("manual/scan/expire", handler.ScanExpiredWeworkUsersManual) // 手动触发扫描企业微信过期用户
		weworkUsers.GET("manual/scan/new", handler.ScanNewHrUsersManual)            // 手动触发扫描HR缓存数据并为新员工创建企业微信账号

		// c7n
		c7nProjects := v1.Group("c7n/projects")
		c7nProjects.GET("manual/cache", handler.CacheC7nProjectsManual)

		// 任务
		tasks := v1.Group("tasks")
		tasks.GET("startall", handler.StartAll)  // 注册并启动所有定时任务
		tasks.GET("fetchall", handler.FetchAll)  // 停止定时任务
		tasks.POST("start", handler.TaskStart)   // 启动定时任务
		tasks.POST("remove", handler.TaskRemove) // 移除定时任务
		tasks.GET("stop", handler.TaskStop)      // 停止所有定时任务
	}

	return r
}
