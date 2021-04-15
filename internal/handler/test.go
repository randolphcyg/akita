package handler

import (
	"github.com/gin-gonic/gin"
)

// Ping 测试
func Ping(c *gin.Context) {
	// var service admin.SlavePingService
	// if err := c.ShouldBindJSON(&service); err == nil {
	// 	res := service.Test()
	// 	c.JSON(200, res)
	// } else {
	// 	c.JSON(200, ErrorResponse(err))
	// }
	res := "success!"
	c.JSON(200, res)
}
