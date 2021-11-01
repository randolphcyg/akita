package model

import (
	"time"

	"gorm.io/gorm"

	"gitee.com/RandolphCYG/akita/pkg/wework/api"
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

// RawWeworkOrder 企业微信工单结构体
type RawWeworkOrder struct {
	SpNo       string `mapstructure:"sp_no"`       // 审批编号
	SpName     string `mapstructure:"sp_name"`     // 审批申请类型名称（审批模板名称）
	SpStatus   int    `mapstructure:"sp_status"`   // 申请单状态 1-审批中；2-已通过；3-已驳回；4-已撤销；6-通过后撤销；7-已删除；10-已支付
	TemplateId string `mapstructure:"template_id"` // 审批模板id
	ApplyTime  int    `mapstructure:"apply_time"`  // 审批申请提交时间,Unix时间戳
	// 申请人信息
	Applyer struct {
		Userid  string `mapstructure:"userid"`
		Partyid string `mapstructure:"partyid"`
	} `mapstructure:"applyer"`
	SpRecord []struct { // 审批流程信息，可能有多个审批节点。
		SpStatus     int `mapstructure:"sp_status"`
		Approverattr int `mapstructure:"approverattr"`
		Details      []struct {
			Approver struct {
				Userid string `mapstructure:"userid"`
			} `mapstructure:"approver"`
			Speech   string   `mapstructure:"speech"`
			SpStatus int      `mapstructure:"sp_status"`
			Sptime   int      `mapstructure:"sptime"`
			MediaId  []string `mapstructure:"media_id"`
		} `mapstructure:"details"`
	} `mapstructure:"sp_record"`
	Notifyer []struct {
		Userid string `mapstructure:"userid"`
	} `mapstructure:"notifyer"`
	// 审批申请数据
	ApplyData struct {
		Contents []struct {
			Control string `mapstructure:"control"` // 控件类型
			Id      string `mapstructure:"id"`
			Title   []struct {
				Text string `mapstructure:"text"`
				Lang string `mapstructure:"lang"`
			} `mapstructure:"title"`
			// 待解析的所有数据
			Value struct {
				// 文本/多行文本控件
				Text string `mapstructure:"text"`
				// 数字控件
				NewNumber string `mapstructure:"new_number"`
				// 金额控件（control参数为Money）
				NewMoney string `mapstructure:"new_money"`
				// 日期/日期+时间控件（control参数为Date）
				Date struct {
					Type       string `mapstructure:"type"`
					STimestamp string `mapstructure:"s_timestamp"`
				} `mapstructure:"date"`
				// 单选/多选控件（control参数为Selector）
				Selector struct {
					Type    string `mapstructure:"type"`
					Options []struct {
						Key   string `mapstructure:"key"`
						Value []struct {
							Text string `mapstructure:"text"`
							Lang string `mapstructure:"lang"`
						} `mapstructure:"value"`
					} `mapstructure:"options"`
					ExpType string `mapstructure:"exp_type"`
				} `mapstructure:"selector"`
				// 成员控件（control参数为Contact，且value参数为members）
				Members []struct {
					Userid string `mapstructure:"userid"`
					Name   string `mapstructure:"name"`
				} `mapstructure:"members"`
				// 部门控件（control参数为Contact，且value参数为departments）
				Departments []struct {
					OpenapiId string `mapstructure:"openapi_id"`
					Name      string `mapstructure:"name"`
				} `mapstructure:"departments"`
				// 附件控件（control参数为File）
				Files []struct {
					FileId string `mapstructure:"file_id"`
				} `mapstructure:"files"`
				// 时长组件（control参数为DateRange）
				DateRange struct {
					Type        string `mapstructure:"type"`
					NewBegin    int    `mapstructure:"new_begin"`
					NewEnd      int    `mapstructure:"new_end"`
					nNwDuration int    `mapstructure:"new_duration"`
				} `mapstructure:"date_range"`
				// 位置控件（control参数为Location）
				Location struct {
					Latitude  string `mapstructure:"type"`
					Longitude string `mapstructure:"new_begin"`
					Title     string `mapstructure:"new_end"`
					Address   string `mapstructure:"new_duration"`
					Time      int    `mapstructure:"time"`
				} `mapstructure:"location"`
				// 关联审批单控件（control参数为RelatedApproval）
				RelatedApproval []struct {
					TemplateNames []struct {
						Text string `mapstructure:"text"`
						Lang string `mapstructure:"lang"`
					} `mapstructure:"template_names"`
					SpStatus   int    `mapstructure:"sp_status"`
					Name       string `mapstructure:"name"`
					CreateTime int    `mapstructure:"create_time"`
					SpNo       string `mapstructure:"sp_no"`
				} `mapstructure:"related_approval"`
				// 公式控件（control参数为Formula）
				Formula struct {
					Value string `mapstructure:"value"`
				} `mapstructure:"formula"`
				// 说明文字控件（control参数为Tips）
				Tips []interface{} `mapstructure:"tips"`

				// 明细控件（control参数为Table）
				Children []struct {
					Id      string `mapstructure:"id"`
					Control string `mapstructure:"control"`
					Title   []struct {
						Text string `mapstructure:"text"`
						Lang string `mapstructure:"lang"`
					} `mapstructure:"title"`
					List []struct {
						Id      string `mapstructure:"id"`
						Control string `mapstructure:"control"`
						Title   []struct {
							Text string `mapstructure:"text"`
							Lang string `mapstructure:"lang"`
						} `mapstructure:"title"`
						Value struct { // 明细各个对象可能有的控件类型
							// 文本/多行文本控件
							Text string `mapstructure:"text"`
							// 数字控件
							NewNumber string `mapstructure:"new_number"`
							// 金额控件（control参数为Money）
							NewMoney string `mapstructure:"new_money"`
							// 日期/日期+时间控件（control参数为Date）
							Date struct {
								Type       string `mapstructure:"type"`
								STimestamp string `mapstructure:"s_timestamp"`
							} `mapstructure:"date"`
							// 单选/多选控件（control参数为Selector）
							Selector struct {
								Type    string `mapstructure:"type"`
								Options []struct {
									Key   string `mapstructure:"key"`
									Value []struct {
										Text string `mapstructure:"text"`
										Lang string `mapstructure:"lang"`
									} `mapstructure:"value"`
								} `mapstructure:"options"`
								ExpType string `mapstructure:"exp_type"`
							} `mapstructure:"selector"`
							// 成员控件（control参数为Contact，且value参数为members）
							Members []struct {
								Userid string `mapstructure:"userid"`
								Name   string `mapstructure:"name"`
							} `mapstructure:"members"`
							// 部门控件（control参数为Contact，且value参数为departments）
							Departments []struct {
								OpenapiId string `mapstructure:"openapi_id"`
								Name      string `mapstructure:"name"`
							} `mapstructure:"departments"`
							// 附件控件（control参数为File）
							Files []struct {
								FileId string `mapstructure:"file_id"`
							} `mapstructure:"files"`
							// 时长组件（control参数为DateRange）
							DateRange struct {
								Type        string `mapstructure:"type"`
								NewBegin    int    `mapstructure:"new_begin"`
								NewEnd      int    `mapstructure:"new_end"`
								nNwDuration int    `mapstructure:"new_duration"`
							} `mapstructure:"date_range"`
							// 位置控件（control参数为Location）
							Location struct {
								Latitude  string `mapstructure:"type"`
								Longitude string `mapstructure:"new_begin"`
								Title     string `mapstructure:"new_end"`
								Address   string `mapstructure:"new_duration"`
								Time      int    `mapstructure:"time"`
							} `mapstructure:"location"`
							// 关联审批单控件（control参数为RelatedApproval）
							RelatedApproval []struct {
								TemplateNames []struct {
									Text string `mapstructure:"text"`
									Lang string `mapstructure:"lang"`
								} `mapstructure:"template_names"`
								SpStatus   int    `mapstructure:"sp_status"`
								Name       string `mapstructure:"name"`
								CreateTime int    `mapstructure:"create_time"`
								SpNo       string `mapstructure:"sp_no"`
							} `mapstructure:"related_approval"`
							// 公式控件（control参数为Formula）
							Formula struct {
								Value string `mapstructure:"value"`
							} `mapstructure:"formula"`
							// 说明文字控件（control参数为Tips）
							Tips []interface{} `mapstructure:"tips"`
						} `mapstructure:"value"`
					} `mapstructure:"list"`
				} `mapstructure:"children"`
			} `mapstructure:"value"`
		} `mapstructure:"contents"`
	} `mapstructure:"apply_data"`
	// 审批申请备注信息，可能有多个备注节点
	Comments []struct {
		CommentUserInfo struct {
			Userid string `mapstructure:"userid"`
		} `mapstructure:"comment_user_info"`
		Commenttime    int      `mapstructure:"commenttime"`
		Commentcontent string   `mapstructure:"commentcontent"`
		Commentid      string   `mapstructure:"commentid"`
		MediaId        []string `mapstructure:"media_id"`
	} `mapstructure:"comments"`
}

// Applicant 申请人
type Applicant struct {
	DisplayName   string   `mapstructure:"姓名"`
	Eid           string   `mapstructure:"工号"`
	Mobile        string   `mapstructure:"手机"`
	Mail          string   `mapstructure:"邮箱"`
	Company       string   `mapstructure:"公司"`
	InitPlatforms []string `mapstructure:"所需平台"`
}

// AccountsRegister 各平台账号注册 工单详情 多个
type AccountsRegister struct {
	SpName  string      `mapstructure:"spName"`
	Partyid string      `mapstructure:"partyid"`
	Userid  string      `mapstructure:"userid"`
	Users   []Applicant `mapstructure:"待申请人员"`
}

// AccountsRegisterSingle 各平台账号注册 工单详情 单个
type AccountsRegisterSingle struct {
	SpName        string   `mapstructure:"spName"`
	Partyid       string   `mapstructure:"partyid"`
	Userid        string   `mapstructure:"userid"`
	DisplayName   string   `mapstructure:"姓名"`
	Eid           string   `mapstructure:"工号"`
	Mobile        string   `mapstructure:"手机"`
	Mail          string   `mapstructure:"邮箱"`
	Company       string   `mapstructure:"公司"`
	InitPlatforms []string `mapstructure:"所需平台"`
}

// UuapPwdRetrieve UUAP密码找回 工单详情
type UuapPwdRetrieve struct {
	SpName      string `mapstructure:"spName"`
	Userid      string `mapstructure:"userid"`
	DisplayName string `mapstructure:"姓名"`
	Eid         string `mapstructure:"工号"`
}

// UuapDisable 账号注销 工单详情
type UuapDisable struct {
	SpName      string `mapstructure:"spName"`
	Userid      string `mapstructure:"userid"`
	DisplayName string `mapstructure:"姓名"`
	Eid         string `mapstructure:"工号"`
}

// RenewalApplicant 申请人
type RenewalApplicant struct {
	DisplayName string   `mapstructure:"姓名"`
	Eid         string   `mapstructure:"工号"`
	Platforms   []string `mapstructure:"平台"`
	Days        string   `mapstructure:"续期天数"`
}

// AccountsRenewal 账号续期 工单详情
type AccountsRenewal struct {
	SpName string             `mapstructure:"spName"`
	Userid string             `mapstructure:"userid"`
	Users  []RenewalApplicant `mapstructure:"待申请人员"`
}

// C7nProject c7n项目
type C7nProject struct {
	Project string   `mapstructure:"项目"`
	Roles   []string `mapstructure:"角色"`
}

// C7nAuthority c7n项目权限 工单详情
type C7nAuthority struct {
	SpName      string       `mapstructure:"spName"`
	Userid      string       `mapstructure:"userid"`
	Eid         string       `mapstructure:"工号"`
	DisplayName string       `mapstructure:"姓名"`
	C7nProjects []C7nProject `mapstructure:"猪齿鱼项目"`
}
