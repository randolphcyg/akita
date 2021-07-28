package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录
func UserLogin(ctx *gin.Context) {
	var service user.UserLoginService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Login(ctx)
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
		res := service.CreateUser(service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// FetchHrData 查询用户
func FetchHrData(ctx *gin.Context) {
	var service user.HrDataService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.FetchHrData(service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// ScanExpiredLdapUsersTask 查询并处理过期的LDAP用户
func ScanExpiredLdapUsersTask(ctx *gin.Context) {
	var service user.LdapUserService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.ScanExpiredLdapUsersTask()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// ScanExpiredLdapUsersManual 查询并处理过期的LDAP用户
func ScanExpiredLdapUsersManual(ctx *gin.Context) {
	var service user.LdapUserService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.ScanExpiredLdapUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// UpdateLdapUsersTask 更新ldap用户信息定时任务
func UpdateLdapUsersTask(ctx *gin.Context) {
	var service user.HrDataService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.UpdateLdapUsersTask()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// UpdateLdapUsersManual 手动更新ldap用户信息
func UpdateLdapUsersManual(ctx *gin.Context) {
	var service user.HrDataService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.UpdateLdapUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
