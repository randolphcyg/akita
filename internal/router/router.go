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
		// 测试
		site.GET("ping", handler.Ping)

		// 用户相关路由
		user := v1.Group("user")
		// 用户登录
		user.GET("login", handler.UserLogin)

		// ldap连接相关路由
		ldapConn := v1.Group("ldap/conn")
		// 获取所有ldap连接配置
		ldapConn.GET("fetch", handler.FetchLdapConn)
		// 增加ldap连接
		ldapConn.POST("create", handler.CreateLdapConn)
		// 更新ldap连接
		ldapConn.POST("update", handler.UpdateLdapConn)
		// 删除ldap连接
		ldapConn.DELETE("delete", handler.DeleteLdapConn)
		// 测试ldap连接
		ldapConn.POST("test", handler.TestLdapConn)

		ldapField := v1.Group("ldap/field")
		// 获取ldap连接的字段明细配置
		ldapField.GET("fetch", handler.FetchLdapField)
		// 增加ldap连接的字段明细
		ldapField.POST("create", handler.CreateLdapField)
		// 更新ldap连接的字段明细
		ldapField.POST("update", handler.UpdateLdapField)
		// 删除ldap连接的字段明细
		ldapField.DELETE("delete", handler.DeleteLdapField)
		// 测试ldap连接的字段明细
		ldapField.POST("test", handler.TestLdapField)
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
