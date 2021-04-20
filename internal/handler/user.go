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
