package handler

import (
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"github.com/gin-gonic/gin"
)

// CacheC7nProjectsManual 手动触发缓存C7N项目
func CacheC7nProjectsManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := c7n.CacheC7nProjectsManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// UpdateC7nUsersManual 手动触发LDAP用户同步到C7N
func UpdateC7nUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := c7n.UpdateC7nUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
