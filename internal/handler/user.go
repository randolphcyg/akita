package handler

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录
func UserLogin(ctx *gin.Context) {
	var service user.UserLoginService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Login(ctx)
		fmt.Println(err)
		ctx.JSON(200, res)
	} else {
		// 错误消息先直接写在这里
		ctx.JSON(200, "登录出现错误!")
	}
}

// FetchLdapUser 查询LDAP用户
func FetchLdapUser(ctx *gin.Context) {
	conn_url := ctx.Query("conn_url")
	if conn_url == "" {
		ctx.JSON(200, "没有传ldap连接")
	}
	var service user.LdapUserService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.FetchUser(conn_url)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// CreateLdapUser 创建ldap用户
func CreateLdapUser(ctx *gin.Context) {
	var service user.LdapUserService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.AddUser(service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
