package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/user"
	"github.com/gin-gonic/gin"
)

// ScanExpiredLdapUsersManual 手动触发扫描过期ldap用户
func ScanExpiredLdapUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := user.ScanExpiredLdapUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// CacheHrUsersManual 手动触发缓存HR用户
func CacheHrUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := user.CacheHrUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// SyncLdapUsersManual 手动触发更新ldap用户
func SyncLdapUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := user.SyncLdapUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
