package handler

import (
	"net/http"

	"gitee.com/RandolphCYG/akita/internal/service/task"
	"gitee.com/RandolphCYG/akita/pkg/crontab"
	"github.com/gin-gonic/gin"
)

// 定时任务字段
type TaskField struct {
	Name string `form:"name" json:"name" xml:"name" binding:"required"`
}

// LdapUsersCronTasksStart 注册所有用户定时任务
func LdapUsersCronTasksStart(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := task.LdapUsersCronTasksStart()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// TaskStart 启动定时任务
func TaskStart(ctx *gin.Context) {
	var taskField TaskField
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.TaskRegister(taskField.Name)
	ctx.JSON(200, res)
}

// TaskStop 停止定时任务
func TaskStop(ctx *gin.Context) {
	var taskField TaskField
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := crontab.TaskStop(taskField.Name)
	ctx.JSON(200, res)
}