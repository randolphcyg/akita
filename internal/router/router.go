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
		// 全局设置相关
		site := v1.Group("site")
		site.GET("ping", handler.Ping) // 存活探针

		// 用户相关路由
		user := v1.Group("users")
		user.GET("login", handler.UserLogin)

		// ldap连接相关路由
		ldapConn := v1.Group("ldap/conns")
		ldapConn.GET("fetch", handler.FetchLdapConn)
		ldapConn.POST("create", handler.CreateLdapConn)
		ldapConn.POST("update", handler.UpdateLdapConn)
		ldapConn.DELETE("delete", handler.DeleteLdapConn)
		ldapConn.POST("test", handler.TestLdapConn)

		// ldap连接的字段明细配置
		ldapField := v1.Group("ldap/fields")
		ldapField.GET("fetch", handler.FetchLdapField) // /api/v1/ldap/field/fetch?conn_url=ldap://192.168.5.55:390
		ldapField.POST("create", handler.CreateLdapField)
		ldapField.POST("update", handler.UpdateLdapField)
		ldapField.DELETE("delete", handler.DeleteLdapField) // /api/v1/ldap/field/delete?conn_url=ldap://192.168.5.55:390
		ldapField.POST("test", handler.TestLdapField)

		// ldap用户
		ldapUser := v1.Group("ldap/users")
		ldapUser.GET("fetch", handler.FetchLdapUser) // 根据conn_url查询LDAP用户 /api/v1/ldap/user/fetch?conn_url=ldap://192.168.5.55:390
		ldapUser.GET("create", handler.CreateLdapUser)
		ldapUser.GET("scan/expire/manual", handler.ScanExpiredLdapUsersManual) // 扫描过期用户
		ldapUser.GET("update/cache/manual", handler.UpdateCacheUsersManual)    // 更新用户到缓存库
		ldapUser.GET("update/ldap/manual", handler.UpdateLdapUsersManual)      // 从缓存库更新用户到ldap
		ldapUser.GET("task", handler.LdapUsersCronTasksStart)                  // 注册并启动所有关于用户的定时任务

		// 通过查询hr数据接口确定是否包含某员工
		hrData := v1.Group("hr")
		hrData.GET("fetch", handler.FetchHrData)

		// 处理企业微信工单
		weworkOder := v1.Group("order")
		weworkOder.POST("handleOrders", handler.HandleOrders) // /api/v1/order/handleOrders

		// 任务
		task := v1.Group("tasks")
		// crontab定时任务
		task.POST("start", handler.TaskStart) // 启动定时任务
		task.POST("stop", handler.TaskStop)   // 停止定时任务
	}

	return r
}
