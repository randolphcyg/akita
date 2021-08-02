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

// WeworkOrder 企业微信工单
type WeworkOrder struct {
	gorm.Model
	SpNo          string `json:"sp_no"  gorm:"<-:create;type:varchar(255);unique_index;not null;comment:审批编号"`      // 审批编号
	ExecuteStatus bool   `json:"execute_status"  gorm:"type:bool;unique_index;not null;comment:执行状态 0 执行失败 1 执行成功"` // 执行状态
	ExecuteMsg    string `json:"execute_msg"  gorm:"type:varchar(500);unique_index;comment:执行信息"`                   // 执行信息
}

// NewWeworkCfg 构造方法
func NewWeworkCfg() WeworkCfg {
	return WeworkCfg{}
}

var WeworkOrderCfg WeworkCfg
var WeworkUuapCfg WeworkCfg
var WeworkOrderObj WeworkOrder
var WeworkUserManageCfg WeworkCfg

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

// GetWeworkUserManageCfg 查询企业微信UUAP公告应用配置
func GetWeworkUserManageCfg() error {
	result := DB.Where("app_name = ?", "通讯录管理").Find(&WeworkUserManageCfg)
	return result.Error
}

/*
* 企业微信工单处理
*
 */

// 新增记录
func CreateOrder(spNo string, status bool, err string) {
	DB.Create(&WeworkOrder{SpNo: spNo, ExecuteStatus: status, ExecuteMsg: err})
}

// 修改记录
func UpdateOrder(spNo string, status bool, err string) {
	DB.Model(&WeworkOrder{}).Where("sp_no = ?", spNo).Updates(WeworkOrder{ExecuteStatus: status, ExecuteMsg: err})
}

// 查询记录
func FetchOrder(spNo string) (result *gorm.DB, order WeworkOrder) {
	result = DB.Where("sp_no = ?", spNo).Find(&order)
	return
}
