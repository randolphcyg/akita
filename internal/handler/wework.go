package handler

import (
	"github.com/gin-gonic/gin"
	"gitee.com/RandolphCYG/akita/internal/service/wework"
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

// CacheUsersManual 手动触发缓存企微用户
func (wuf weworkUserField) CacheUsersManual(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := wework.CacheUsersManual()
		ctx.JSON(200, res)
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
