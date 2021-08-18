package handler

import (
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"github.com/gin-gonic/gin"
)

// CacheC7nProjectsManual 手动触发缓存c7n项目
func CacheC7nProjectsManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := c7n.CacheC7nProjectsManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
