package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/hruser"
	"github.com/gin-gonic/gin"
)

type HrUserHandler interface {
	CacheHrUsersManual(ctx *gin.Context)
}

// hrUserField 定时任务字段
type hrUserField struct {
	Name string
}

func NewHrUserHandler() HrUserHandler {
	return &hrUserField{}
}

// CacheHrUsersManual 手动触发缓存HR用户
func (hu hrUserField) CacheHrUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := hruser.CacheUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
