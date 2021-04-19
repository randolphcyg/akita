package handler

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	var service user.UserLoginService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Login(c)
		fmt.Println(err)
		c.JSON(200, res)
	} else {
		// 错误消息先直接写在这里
		c.JSON(200, "登录出现错误!")
	}
}
