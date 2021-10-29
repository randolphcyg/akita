package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/ldapuser"
	"github.com/gin-gonic/gin"
)

type LdapUserHandler interface {
	ScanExpiredLdapUsersManual(ctx *gin.Context)
	SyncLdapUsersManual(ctx *gin.Context)
}

// ldapUserField 定时任务字段
type ldapUserField struct {
	Name string
}

func NewLdapUserHandler() LdapUserHandler {
	return &ldapUserField{}
}

// ScanExpiredLdapUsersManual 手动触发扫描过期ldap用户
func (lu ldapUserField) ScanExpiredLdapUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := ldapuser.ScanExpiredUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// SyncLdapUsersManual 手动触发更新ldap用户
func (lu ldapUserField) SyncLdapUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := ldapuser.SyncUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
