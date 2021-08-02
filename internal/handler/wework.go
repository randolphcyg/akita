package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/wework"
	"github.com/gin-gonic/gin"
)

// HandleOrders 处理工单
func HandleOrders(ctx *gin.Context) {
	var service wework.Order
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.HandleOrders(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
