package router

import (
	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/internal/handler"
	"gitee.com/RandolphCYG/akita/pkg/log"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	if conf.Conf.System.Mode == "debug" {
		log.Log().Info("#######当前运行模式：debug")
		return InitDebugRouter()
	} else {
		log.Log().Info("#######当前运行模式：test")
		return InitTestRouter()
	}
}

// InitDebugRouter 初始化测试模式路由
func InitDebugRouter() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")

	/*
		路由
	*/
	{
		// 全局设置相关
		site := v1.Group("site")
		site.GET("ping", handler.Ping)

		// 用户相关路由
		user := v1.Group("user")
		// 用户登录
		user.GET("login", handler.UserLogin)

		// ldap连接相关路由
		ldap := v1.Group("ldap")
		// 获取所有ldap连接配置
		ldap.GET("fetch", handler.FetchLdapConn)
		// 增加ldap连接
		ldap.POST("create", handler.CreateLdapConn)
		// 更新ldap连接
		ldap.POST("update", handler.UpdateLdapConn)
		// 删除ldap连接
		ldap.DELETE("delete", handler.DeleteLdapConn)
	}

	return r
}

// InitTestRouter 初始化测试模式路由
func InitTestRouter() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")

	/*
		路由
	*/
	{
		// 全局设置相关
		site := v1.Group("site")
		site.GET("ping", handler.Ping)

		// 用户相关路由
		user := v1.Group("user")
		// 用户登录
		user.POST("login", handler.UserLogin)
	}
	return r
}
