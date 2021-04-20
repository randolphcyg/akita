package handler

import (
	"github.com/gin-gonic/gin"
)

// Ping 测试
func Ping(ctx *gin.Context) {
	res := "success!"
	ctx.JSON(200, res)
}
