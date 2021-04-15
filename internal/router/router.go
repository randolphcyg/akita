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
		v1.GET("ping", handler.Ping)
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
		v1.GET("ping", handler.Ping)
	}
	return r
}
