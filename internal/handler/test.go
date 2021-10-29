package handler

import (
	"gitee.com/RandolphCYG/akita/internal/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitee.com/RandolphCYG/akita/pkg/util"
)

// Ping test calendar module | ldapconn connection pool
func Ping(ctx *gin.Context) {
	if isSilent, festival := util.IsHolidaySilentMode(time.Now().AddDate(0, 0, 1)); isSilent {
		if festival != "" {
			ctx.JSON(200, "happy "+festival+" ~")
		} else {
			ctx.JSON(200, "tomorrow is weekend ~ Have a good weekend, you too, and you ?; Now the length of ldapconn connection pool is "+strconv.Itoa(model.LdapPool.Len()))
		}
	} else {
		ctx.JSON(200, "tomorrow is not festival or weekend; Now the length of ldapconn connection pool is "+strconv.Itoa(model.LdapPool.Len()))
	}
}
