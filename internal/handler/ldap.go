package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/conn"
	"github.com/gin-gonic/gin"
)

// FetchLdapConn 查询所有ldap连接
func FetchLdapConn(ctx *gin.Context) {
	var service conn.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.Fetch()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// CreateLdapConn 增加ldap连接
func CreateLdapConn(ctx *gin.Context) {
	var service conn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Add(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// UpdateLdapConn 更新ldap连接
func UpdateLdapConn(ctx *gin.Context) {
	var service conn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Update(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// DeleteLdapConn 删除ldap连接
func DeleteLdapConn(ctx *gin.Context) {
	var service conn.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Delete(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// TestLdapConn 测试ldap连接
func TestLdapConn(ctx *gin.Context) {
	var service conn.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		// 用 map[string]uint 接收json字符串
		json := make(map[string]uint)
		ctx.BindJSON(&json)
		var id uint = json["id"]
		res := service.Test(id)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// FetchLdapField 查询对应ldap连接的ldap字段明细
func FetchLdapField(ctx *gin.Context) {
	url := ctx.Query("conn_url")
	var service conn.LdapFieldService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.FetchField(url)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// CreateLdapField 增加ldap字段记录
func CreateLdapField(ctx *gin.Context) {
	var service conn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.AddField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// UpdateLdapField 更新ldap字段记录
func UpdateLdapField(ctx *gin.Context) {
	var service conn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.UpdateField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// DeleteLdapField 删除ldap字段记录
func DeleteLdapField(ctx *gin.Context) {
	url := ctx.Query("conn_url")
	var service conn.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.DeleteField(url)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, err)
	}
}

// TestLdapField 测试ldap字段记录
func TestLdapField(ctx *gin.Context) {
	// var service conn.LdapConnService
	// if err := ctx.ShouldBindUri(&service); err == nil {
	// 	// 用 map[string]uint 接收json字符串
	// 	json := make(map[string]uint)
	// 	ctx.BindJSON(&json)
	// 	var id uint = json["id"]
	// 	res := service.Test(id)
	// 	ctx.JSON(200, res)
	// } else {
	// 	ctx.JSON(200, err)
	// }
	// TODO
}
