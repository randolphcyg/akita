package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/wework"
	"github.com/gin-gonic/gin"
)

// HandleOrders 处理工单
func HandleOrders(ctx *gin.Context) {
	var service wework.Order
	if err := ctx.ShouldBindJSON(&service); err == nil {
		go func() {
			service.HandleOrders(&service)
		}()
		ctx.JSON(200, "Thanks, tabby! Order processing...")
	} else {
		ctx.JSON(200, err)
	}
}

// CacheUsersManual 手动触发缓存企业微信用户
func CacheUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		go func() {
			err = wework.CacheWeworkUsersManual()
		}()
		ctx.JSON(200, "稍等片刻将更新企业微信用户缓存...")
	} else {
		ctx.JSON(200, err)
	}
}

// ScanExpiredWeworkUsersManual 手动触发扫描企业微信过期用户
func ScanExpiredWeworkUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := wework.ScanExpiredWeworkUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
