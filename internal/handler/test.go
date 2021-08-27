package handler

import (
	"time"

	"gitee.com/RandolphCYG/akita/pkg/util"
	"github.com/gin-gonic/gin"
)

// Ping 就绪探针
func Ping(ctx *gin.Context) {
	if isE := util.IsWeekend(time.Now().AddDate(0, 0, 1)); isE {
		ctx.JSON(200, "tomorrow is weekwnd~")
	} else {
		ctx.JSON(200, "tomorrow is not weekwnd !!")
	}
}
