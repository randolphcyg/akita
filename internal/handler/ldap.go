package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/ldap"
	"github.com/gin-gonic/gin"
)

// FetchLdapConn 查询所有ldap连接
func FetchLdapConn(ctx *gin.Context) {
	var service ldap.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		res := service.Fetch()
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "查询错误")
	}
}

// CreateLdapConn 增加ldap连接
func CreateLdapConn(ctx *gin.Context) {
	var service ldap.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Add(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "增加错误")
	}
}

// UpdateLdapConn 更新ldap连接
func UpdateLdapConn(ctx *gin.Context) {
	var service ldap.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Update(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "修改错误")
	}
}

// DeleteLdapConn 删除ldap连接
func DeleteLdapConn(ctx *gin.Context) {
	var service ldap.LdapConnService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.Delete(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "删除错误")
	}
}

// TestLdapConn 测试ldap连接
func TestLdapConn(ctx *gin.Context) {
	var service ldap.LdapConnService
	if err := ctx.ShouldBindUri(&service); err == nil {
		// 用 map[string]uint 接收json字符串
		json := make(map[string]uint)
		ctx.BindJSON(&json)
		var id uint = json["id"]
		res := service.Test(id)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "测试失败")
	}
}

// FetchLdapField 查询所有ldap字段
func FetchLdapField(ctx *gin.Context) {
	// var service ldap.LdapConnService
	// if err := ctx.ShouldBindUri(&service); err == nil {
	// 	res := service.Fetch()
	// 	ctx.JSON(200, res)
	// } else {
	// 	ctx.JSON(200, "查询错误")
	// }
	// TODO
}

// CreateLdapField 增加ldap字段记录
func CreateLdapField(ctx *gin.Context) {
	var service ldap.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.AddField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "增加错误")
	}
}

// UpdateLdapField 更新ldap字段记录
func UpdateLdapField(ctx *gin.Context) {
	var service ldap.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.UpdateField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "修改错误")
	}
}

// DeleteLdapField 删除ldap字段记录
func DeleteLdapField(ctx *gin.Context) {
	var service ldap.LdapFieldService
	if err := ctx.ShouldBindJSON(&service); err == nil {
		res := service.DeleteField(&service)
		ctx.JSON(200, res)
	} else {
		ctx.JSON(200, "删除错误")
	}
}

// TestLdapField 测试ldap字段记录
func TestLdapField(ctx *gin.Context) {
	// var service ldap.LdapConnService
	// if err := ctx.ShouldBindUri(&service); err == nil {
	// 	// 用 map[string]uint 接收json字符串
	// 	json := make(map[string]uint)
	// 	ctx.BindJSON(&json)
	// 	var id uint = json["id"]
	// 	res := service.Test(id)
	// 	ctx.JSON(200, res)
	// } else {
	// 	ctx.JSON(200, "测试失败")
	// }
	// TODO
}
