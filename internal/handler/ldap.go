package handler

import (
	"github.com/gin-gonic/gin"
	"gitee.com/RandolphCYG/akita/internal/service/ldapconn"
)

type LdapConnHandler interface {
	Fetch(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Test(ctx *gin.Context)
}

// ldapConnField 定时任务字段
type ldapConnField struct {
	Name string
}

func NewLdapConnHandler() LdapConnHandler {
	return &ldapConnField{}
}

// Fetch 查询所有ldap连接
func (lcf ldapConnField) Fetch(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.Fetch()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Create 增加ldap连接
func (lcf ldapConnField) Create(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Add(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Update 更新ldap连接
func (lcf ldapConnField) Update(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Update(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Delete 删除ldap连接
func (lcf ldapConnField) Delete(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Delete(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Test 测试ldap连接
func (lcf ldapConnField) Test(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		// 用 map[string]uint 接收json字符串
		json := make(map[string]uint)
		ctx.BindJSON(&json)
		var id = json["id"]
		res := service.Test(id)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

type LdapFieldHandler interface {
	Fetch(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Test(ctx *gin.Context)
}

// ldapField 定时任务字段
type ldapField struct {
	Name string
}

func NewLdapFieldHandler() LdapFieldHandler {
	return &ldapField{}
}

// Fetch 查询对应ldap连接的ldap字段明细
func (lf ldapField) Fetch(ctx *gin.Context) {
	url := ctx.Query("conn_url")
	var service ldapconn.LdapFieldService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.FetchField(url)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Create 增加ldap字段记录
func (lf ldapField) Create(ctx *gin.Context) {
	var service ldapconn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.AddField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Update 更新ldap字段记录
func (lf ldapField) Update(ctx *gin.Context) {
	var service ldapconn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.UpdateField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Delete 删除ldap字段记录
func (lf ldapField) Delete(ctx *gin.Context) {
	url := ctx.Query("conn_url")
	var service ldapconn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.DeleteField(url)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// Test 测试ldap字段记录
func (lf ldapField) Test(ctx *gin.Context) {
	var service ldapconn.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		// 用 map[string]uint 接收json字符串
		json := make(map[string]uint)
		err := ctx.BindJSON(&json)
		if err != nil {
			return
		}
		res := service.Test(json["id"])
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}
