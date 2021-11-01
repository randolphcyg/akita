package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitee.com/RandolphCYG/akita/internal/service/task"
)

type TaskHandler interface {
	Start(ctx *gin.Context)
	Remove(ctx *gin.Context)
	StopAll(ctx *gin.Context)
	FetchAll(ctx *gin.Context)
}

// taskField 定时任务字段
type taskField struct {
	Name string
}

func NewTaskHandler() TaskHandler {
	return &taskField{}
}

// Start 启动定时任务
func (taskField taskField) Start(ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.Start(taskField.Name)
	ctx.JSON(200, res)
}

// Remove 移除定时任务
func (taskField taskField) Remove(ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.Remove(taskField.Name)
	ctx.JSON(200, res)
}

// StopAll 停止所有定时任务
func (taskField taskField) StopAll(ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(&taskField); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.StopAll()
	ctx.JSON(200, res)
}

// FetchAll 查询所有定时任务
func (taskField taskField) FetchAll(ctx *gin.Context) {
	if err := ctx.ShouldBind(0); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res := task.GetTasks()
	ctx.JSON(200, res)
}
