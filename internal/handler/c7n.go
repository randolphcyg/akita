package handler

import (
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"github.com/gin-gonic/gin"
)

type C7nHandler interface {
	CacheProjectsManual(ctx *gin.Context)
	SyncUsersManual(ctx *gin.Context)
}

// c7nField 无用结构体 用于interface
type c7nField struct {
	Name string
}

func NewC7nHandler() C7nHandler {
	return &c7nField{}
}

// CacheProjectsManual 手动触发缓存C7N项目
func (c c7nField) CacheProjectsManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := c7n.CacheProjectsManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// SyncUsersManual 手动触发LDAP用户的同步
func (c c7nField) SyncUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := c7n.SyncUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
