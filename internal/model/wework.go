package model

import (
	"gorm.io/gorm"
)

// WeworkCfg 企业微信连接配置
type WeworkCfg struct {
	gorm.Model
	// 企业ID
	CorpId    string `json:"corp_id" gorm:"type:varchar(255);unique_index;not null;comment:企业ID"`
	AppId     int    `json:"app_id" gorm:"type:int(25);unique_index;not null;comment:App ID"`
	AppName   string `json:"app_name" gorm:"type:varchar(255);unique_index;not null;comment:App名称"`
	AppSecret string `json:"app_secret" gorm:"type:varchar(255);unique_index;not null;comment:App秘钥"`
}

// NewWeworkCfg 构造方法
func NewWeworkCfg() WeworkCfg {
	return WeworkCfg{}
}

var WeworkOrderCfg WeworkCfg
var WeworkUuapCfg WeworkCfg

// GetWeworkOrderCfg 查询企业微信审批应用配置
func GetWeworkOrderCfg() error {
	result := DB.Where("app_name = ?", "审批").Find(&WeworkOrderCfg)
	return result.Error
}

// GetWeworkUuapCfg 查询企业微信UUAP公告应用配置
func GetWeworkUuapCfg() error {
	result := DB.Where("app_name = ?", "UUAP公告应用").Find(&WeworkUuapCfg)
	return result.Error
}
