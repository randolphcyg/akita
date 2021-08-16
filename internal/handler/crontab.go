package handler

import (
	"net/http"

	"gitee.com/RandolphCYG/akita/internal/service/task"
	"github.com/gin-gonic/gin"
)

// 定时任务字段
type TaskField struct {
	Name string `form:"name" json:"name" xml:"name" binding:"required"`
}

// StartAll 注册所有用户定时任务
func StartAll(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err == nil {
		res := task.StartAll()
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
	res := task.TaskSart(taskField.Name)
	ctx.JSON(200, res)
}

// TaskRemove 移除定时任务
func TaskRemove(ctx *gin.Context) {
	var taskField TaskField
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.TaskRemove(taskField.Name)
	ctx.JSON(200, res)
}

// TaskStop 停止所有定时任务
func TaskStop(ctx *gin.Context) {
	var taskField TaskField
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.TaskStop()
	ctx.JSON(200, res)
}

// FetchAll 查询所有定时任务
func FetchAll(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.FetchTasks()
	ctx.JSON(200, res)
}
