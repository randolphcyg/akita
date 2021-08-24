package router

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/internal/handler"
	"gitee.com/RandolphCYG/akita/pkg/log"

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
		panic("gin mode unknown: " + conf.Conf.System.Mode + " (available mode: debug release test)")
	}
	return router()
}

// router 路由
func router() *gin.Engine {
	r := gin.New()
	r.Use(log.LogerMiddleware())
	v1 := r.Group("/api/v1")

	{
		// 全局设置
		site := v1.Group("site")
		site.GET("ping", handler.Ping) // 存活探针

		// ldap 连接
		ldapConns := v1.Group("ldap/conns")
		ldapConns.GET("fetch", handler.FetchLdapConn)
		ldapConns.POST("create", handler.CreateLdapConn)
		ldapConns.POST("update", handler.UpdateLdapConn)
		ldapConns.DELETE("delete", handler.DeleteLdapConn)
		ldapConns.POST("test", handler.TestLdapConn)
		// ldap 连接的字段明细配置
		ldapFields := v1.Group("ldap/fields")
		ldapFields.GET("fetch", handler.FetchLdapField)
		ldapFields.POST("create", handler.CreateLdapField)
		ldapFields.POST("update", handler.UpdateLdapField)
		ldapFields.DELETE("delete", handler.DeleteLdapField)
		ldapFields.POST("test", handler.TestLdapField)
		// ldap 用户
		ldapUsers := v1.Group("ldap/users")
		ldapUsers.GET("manual/cache/hr", handler.CacheHrUsersManual)            // 手动触发缓存HR用户
		ldapUsers.GET("manual/sync", handler.SyncLdapUsersManual)               // 手动触发更新ldap用户
		ldapUsers.GET("manual/scan/expire", handler.ScanExpiredLdapUsersManual) // 手动触发扫描过期ldap用户

		// wework 工单
		weworkOrders := v1.Group("wework/orders")
		weworkOrders.POST("handle", handler.HandleOrders)
		// wework 用户
		weworkUsers := v1.Group("wework/users")
		weworkUsers.GET("manual/cache", handler.CacheUsersManual)                   // 手动触发缓存企业微信用户
		weworkUsers.GET("manual/scan/expire", handler.ScanExpiredWeworkUsersManual) // 手动触发扫描企业微信过期用户
		weworkUsers.GET("manual/scan/new", handler.ScanNewHrUsersManual)            // 手动触发扫描HR缓存数据并为新员工创建企业微信账号

		// c7n 项目
		c7nProjects := v1.Group("c7n/projects")
		c7nProjects.GET("manual/cache", handler.CacheC7nProjectsManual) // 手动触发缓存C7N项目
		// c7n 用户
		c7nUsers := v1.Group("c7n/users")
		c7nUsers.GET("manual/sync", handler.UpdateC7nUsersManual) // 手动触发LDAP用户同步到C7N

		// tasks 定时任务
		tasks := v1.Group("tasks")
		tasks.GET("fetchall", handler.FetchAll)  // 停止定时任务
		tasks.POST("start", handler.TaskStart)   // 启动定时任务
		tasks.POST("remove", handler.TaskRemove) // 移除定时任务
		tasks.GET("stop", handler.TaskStop)      // 停止所有定时任务
	}

	// 生产模式打印路由
	if gin.Mode() == gin.ReleaseMode {
		for _, p := range r.Routes() {
			nuHandlers := len(r.Handlers)
			routePrint("%-6s %-25s --> %s (%d handlers)\n", p.Method, p.Path, p.Handler, nuHandlers)
			// "+"http://"+conf.Conf.System.Addr+"  // 不提供完整地址防止用户点击太快导致触发不该执行的功能
		}
	}

	return r
}

// 重写路由打印方法
func routePrint(format string, values ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stdout, "[GIN-release] "+format, values...)
}
