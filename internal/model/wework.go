package model

import (
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"time"

	"gorm.io/gorm"
)

/*
* 企微连接配置
*
 */

// WeworkCfg 企微连接配置
type WeworkCfg struct {
	gorm.Model
	// 企业ID
	CorpId    string `json:"corp_id" gorm:"type:varchar(255);unique_index;not null;comment:企业ID"`
	AppId     int    `json:"app_id" gorm:"type:int(25);unique_index;not null;comment:App ID"`
	AppName   string `json:"app_name" gorm:"type:varchar(255);unique_index;not null;comment:App名称"`
	AppSecret string `json:"app_secret" gorm:"type:varchar(255);unique_index;not null;comment:App秘钥"`
}

var (
	WeworkOrderCfg      WeworkCfg // 企微审批工单配置
	WeworkUuapCfg       WeworkCfg // UUAP公告应用
	WeworkUserManageCfg WeworkCfg // 企微通讯录管理
	CorpAPIUserManager  *api.CorpAPI
	CorpAPIMsg          *api.CorpAPI
	CorpAPIOrder        *api.CorpAPI
)

func InitWework() {
	CorpAPIUserManager = api.NewCorpAPI(WeworkUserManageCfg.CorpId, WeworkUserManageCfg.AppSecret)
	CorpAPIMsg = api.NewCorpAPI(WeworkUuapCfg.CorpId, WeworkUuapCfg.AppSecret)
	CorpAPIOrder = api.NewCorpAPI(WeworkOrderCfg.CorpId, WeworkOrderCfg.AppSecret)
}

// GetWeworkOrderCfg 查询企业微信审批应用配置
func GetWeworkOrderCfg() error {
	result := DB.Where("app_name = ?", "审批").Find(&WeworkOrderCfg)
	return result.Error
}

// GetWeworkUuapCfg 查询企微UUAP公告应用配置
func GetWeworkUuapCfg() error {
	result := DB.Where("app_name = ?", "UUAP公告应用").Find(&WeworkUuapCfg)
	return result.Error
}

// GetWeworkUserManageCfg 查询企微UUAP公告应用配置
func GetWeworkUserManageCfg() error {
	result := DB.Where("app_name = ?", "通讯录管理").Find(&WeworkUserManageCfg)
	return result.Error
}

/*
* 企微工单操作记录
*
 */

// WeworkOrder 企微工单操作记录
type WeworkOrder struct {
	gorm.Model
	SpNo          string `json:"sp_no"  gorm:"<-:create;type:varchar(255);unique_index;not null;comment:审批编号"`      // 审批编号
	ExecuteStatus bool   `json:"execute_status"  gorm:"type:bool;unique_index;not null;comment:执行状态 0 执行失败 1 执行成功"` // 执行状态
	ExecuteMsg    string `json:"execute_msg"  gorm:"type:varchar(500);unique_index;comment:执行信息"`                   // 执行信息
}

// CreateOrder 新增记录
func CreateOrder(spNo string, status bool, err string) {
	DB.Create(&WeworkOrder{SpNo: spNo, ExecuteStatus: status, ExecuteMsg: err})
}

// UpdateOrder 修改记录
func UpdateOrder(spNo string, status bool, err string) {
	DB.Model(&WeworkOrder{}).Where("sp_no = ?", spNo).Updates(WeworkOrder{ExecuteStatus: status, ExecuteMsg: err})
}

// FetchOrder 查询记录
func FetchOrder(spNo string) (result *gorm.DB, order WeworkOrder) {
	result = DB.Where("sp_no = ?", spNo).Find(&order)
	return
}

/*
* 企微操作用户记录
*
 */

// WeworkUserSyncRecord 企微用户变化记录
type WeworkUserSyncRecord struct {
	gorm.Model
	UserId   string `json:"user_id" gorm:"type:varchar(255);not null;comment:用户ID"`
	Name     string `json:"name" gorm:"type:varchar(255);not null;comment:真实姓名"`
	Eid      string `json:"eid" gorm:"type:varchar(255);not null;comment:工号"`
	SyncKind string `json:"sync_kind" gorm:"type:varchar(255);not null;comment:类别"`
}

// CreateWeworkUserSyncRecord 企微用户变化记录
func CreateWeworkUserSyncRecord(userId, name, eid, syncKind string) {
	DB.Model(&WeworkUserSyncRecord{}).Create(&WeworkUserSyncRecord{UserId: userId, Name: name, Eid: eid, SyncKind: syncKind})
}

// UpdateWeworkUserSyncRecord 更新 企微用户变化记录
func UpdateWeworkUserSyncRecord(userId, name, eid, syncKind, newSyncKind string) {
	DB.Model(&WeworkUserSyncRecord{}).Where("user_id = ? AND name = ? AND eid = ? and sync_kind = ?", userId, name, eid, syncKind).Update("sync_kind", newSyncKind)
}

// FetchWeworkUserSyncRecord 查询一段时间的企微用户变化记录
func FetchWeworkUserSyncRecord(offsetBefore, offsetAfter int) (weworkUserSyncRecord []WeworkUserSyncRecord, err error) {
	begin, _ := time.Parse("2006-01-02", time.Now().AddDate(0, 0, offsetBefore).Format("2006-01-02")) // 开始日期的零点
	end, _ := time.Parse("2006-01-02", time.Now().AddDate(0, 0, offsetAfter).Format("2006-01-02"))    // 结束日期的最后一秒
	_ = DB.Where("created_at BETWEEN ? AND ?", begin, end).Find(&weworkUserSyncRecord)
	return
}

/*
* 企微信息模板 缓存
*
 */

// WeworkMsgTemplate 企微信息模板
type WeworkMsgTemplate struct {
	Key   string `json:"key" gorm:"type:varchar(255);not null;comment:模板键值"`
	Value string `json:"value" gorm:"type:varchar(4000);not null;comment:模板内容"`
}

// CreateWeworkMsgTemplate 增 企微信息模板
func CreateWeworkMsgTemplate(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create(&WeworkMsgTemplate{Key: key, Value: value})
}

// DeleteWeworkMsgTemplate 删 企微信息模板
func DeleteWeworkMsgTemplate(key string) {
	DB.Where("key = ?", key).Delete(&WeworkMsgTemplate{})
}

// UpdateWeworkMsgTemplate 改 企微信息模板
func UpdateWeworkMsgTemplate(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create(&WeworkMsgTemplate{Key: key, Value: value})
}

// FetchWeworkMsgTemplate 查 企微信息模板
func FetchWeworkMsgTemplate(key string) (weworkUserSyncRecord WeworkUserSyncRecord, err error) {
	DB.Where("key = ?", key).First(&weworkUserSyncRecord)
	return
}

// FetchWeworkMsgTemplates 查 企微信息模板列表
func FetchWeworkMsgTemplates() (weworkMsgTemplates []WeworkMsgTemplate, err error) {
	DB.Find(&weworkMsgTemplates)
	return
}
