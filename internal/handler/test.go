package handler

import (
	"time"

	"gitee.com/RandolphCYG/akita/pkg/util"
	"github.com/gin-gonic/gin"
)

// Ping 就绪探针
func Ping(ctx *gin.Context) {
	if isSilent, festival := util.IsHolidaySilentMode(time.Now().AddDate(0, 0, 1)); isSilent {
		if festival != "" {
			ctx.JSON(200, "happy "+festival+" ~")
		} else {
			ctx.JSON(200, "tomorrow is weekend ~ Have a good weekend, you too, and you ?")
		}

	} else {
		ctx.JSON(200, "tomorrow is not festival or weekend")
	}
}
