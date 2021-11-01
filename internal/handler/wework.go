package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/wework"
	"github.com/gin-gonic/gin"
)

type WeworkOrdersHandler interface {
	HandleOrders(ctx *gin.Context)
}

// weworkOrdersField 定时任务字段
type weworkOrdersField struct {
	Name string
}

func NewWeworkOrdersHandler() WeworkOrdersHandler {
	return &weworkOrdersField{}
}

// HandleOrders 处理工单
func (wof weworkOrdersField) HandleOrders(ctx *gin.Context) {
	var service wework.Order
	if err := ctx.ShouldBindJSON(&service); err == nil {
		go func() {
			err := service.HandleOrders()
			if err != nil {
				return
			}
		}()
		ctx.JSON(200, "Thanks, tabby! Order processing...")
	} else {
		ctx.JSON(200, err)
	}
}

type WeworkUserHandler interface {
	CacheUsersManual(ctx *gin.Context)
	ScanExpiredUsersManual(ctx *gin.Context)
	ScanNewHrUsersManual(ctx *gin.Context)
}

// weworkUserField 定时任务字段
type weworkUserField struct {
	Name string
}

func NewWeworkUserHandler() WeworkUserHandler {
	return &weworkUserField{}
}

// CacheUsersManual 手动触发缓存企业微信用户
func (wuf weworkUserField) CacheUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		go func() {
			err = wework.CacheUsersManual()
		}()
		ctx.JSON(200, "稍等片刻将更新企业微信用户缓存...")
	} else {
		ctx.JSON(200, err)
	}
}

// ScanExpiredUsersManual 手动触发扫描企业微信过期用户
func (wuf weworkUserField) ScanExpiredUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := wework.ScanExpiredUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// ScanNewHrUsersManual 手动触发扫描HR数据并未新员工创建企业微信账号
func (wuf weworkUserField) ScanNewHrUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := wework.ScanNewHrUsersManual()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
