package handler

import (
	"gitee.com/RandolphCYG/akita/internal/service/ldap"
	"github.com/gin-gonic/gin"
)

// FetchLdapConn 查询所有ldap连接
func FetchLdapConn(c *gin.Context) {
	var service ldap.LdapConnService
	if err := c.ShouldBindUri(&service); err == nil {
		res := service.Fetch()
		c.JSON(200, res)
	} else {
		c.JSON(200, "查询错误")
	}
}

// CreateLdapConn 增加ldap连接
func CreateLdapConn(c *gin.Context) {
	var service ldap.LdapConnService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Add(&service)
		c.JSON(200, res)
	} else {
		c.JSON(200, "增加错误")
	}
}

// UpdateLdapConn 更新ldap连接
func UpdateLdapConn(c *gin.Context) {
	var service ldap.LdapConnService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Update(&service)
		c.JSON(200, res)
	} else {
		c.JSON(200, "修改错误")
	}
}

// DeleteLdapConn 删除ldap连接
func DeleteLdapConn(c *gin.Context) {
	var service ldap.LdapConnService
	if err := c.ShouldBindJSON(&service); err == nil {
		res := service.Delete(&service)
		c.JSON(200, res)
	} else {
		c.JSON(200, "删除错误")
	}
}
