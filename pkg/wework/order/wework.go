package order

import (
	"strings"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"github.com/goinggo/mapstructure"
)

// 全局变量
var (
	WeworkCfg *model.WeworkCfg
)

// 企业微信工单结构体
type WeworkOrder struct {
	SpNo       string `mapstructure:"sp_no"`       // 审批编号
	SpName     string `mapstructure:"sp_name"`     // 审批申请类型名称（审批模板名称）
	SpStatus   int    `mapstructure:"sp_status"`   // 申请单状态：1-审批中；2-已通过；3-已驳回；4-已撤销；6-通过后撤销；7-已删除；10-已支付
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
					file_id string `mapstructure:"FileId"`
				}
				// 明细控件（control参数为Table）
				Children []struct {
					List []struct {
						Control string `mapstructure:"control"`
						Id      string `mapstructure:"id"`
						Title   []struct {
							Text string `mapstructure:"text"`
							Lang string `mapstructure:"lang"`
						} `mapstructure:"title"`
						Value []struct {
							Text string `mapstructure:"text"`
						} `mapstructure:"value"`
					} `mapstructure:"list"`
				} `mapstructure:"children"`
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

// 处理后的工单详情 这里维护与企业微信工单对应关系
type WeworkOrderDetails struct {
	SpName  string `mapstructure:"spName"`
	Partyid string `mapstructure:"partyid"`
	Userid  string `mapstructure:"userid"`
	Remarks string `mapstructure:"备注"`
	Name    string `mapstructure:"姓名"`
	Title   string `mapstructure:"岗位"`
	Eid     string `mapstructure:"工号"`
	Mobile  string `mapstructure:"手机"`
	Mail    string `mapstructure:"邮箱"`
	Depart  string `mapstructure:"部门"`
}

// 找回密码工单
type WeworkOrderResetUuapPwd struct {
	Name string `mapstructure:"姓名"`
	Uuap string `mapstructure:"UUAP账号"`
}

// 将企业微信原始工单转换为对应工单
func RawOrderToObj(rawInfo interface{}) (orderData map[string]interface{}, err error) {
	var weworkOrder WeworkOrder
	// 反序列化工单详情
	if err = mapstructure.Decode(rawInfo, &weworkOrder); err != nil {
		log.Log().Error("Occur error when deserialize map to struct:%v", err)
		return
	}

	// 判断工单状态 可以拓展操作
	if weworkOrder.SpStatus != 2 {
		log.Log().Error("工单不是已审批状态!")
		return
	}

	// 清洗工单
	orderData = make(map[string]interface{})
	orderData["spName"] = weworkOrder.SpName
	orderData["partyid"] = weworkOrder.Applyer.Partyid
	orderData["userid"] = weworkOrder.Applyer.Userid
	for _, con := range weworkOrder.ApplyData.Contents {
		switch con.Control {
		case "Number":
			orderData[con.Title[0].Text] = con.Value.NewNumber
		case "Text":
			orderData[con.Title[0].Text] = strings.ToLower(strings.TrimSpace(con.Value.Text)) // 字符串去除空格并转为小写
		case "Textarea":
			orderData[con.Title[0].Text] = strings.ToLower(strings.TrimSpace(con.Value.Text)) // 字符串去除空格并转为小写
		case "Date":
			orderData[con.Title[0].Text] = con.Value.Date.STimestamp
		case "Selector":
			if con.Value.Selector.Type == "multi" { // 多选
				tempSelectors := make([]string, len(con.Value.Selector.Options)-2)
				for _, value := range con.Value.Selector.Options {
					tempSelectors = append(tempSelectors, value.Value[0].Text)
				}
				orderData[con.Title[0].Text] = tempSelectors
			} else { // 单选
				orderData[con.Title[0].Text] = con.Value.Selector.Options[0].Value[0].Text
			}
		case "Tips": // 忽略说明类型
			continue
		case "Contact":
			orderData[con.Title[0].Text] = con.Value.Members
		default:
			log.Log().Error("包含未处理工单项类型：%v，请及时补充后端逻辑!", con.Control)
		}
	}
	return
}

// 将原始工单转换为UUAP创建工单详情结构体 以后这种方法通用与原始工单的解析
func OriginalOrderToUuapCreateOrder(weworkOrder map[string]interface{}) (weworkOrderDetails WeworkOrderDetails) {
	if err := mapstructure.Decode(weworkOrder, &weworkOrderDetails); err != nil {
		log.Log().Error("原始工单转换错误:%v", err)
	}
	return
}

// 将原始工单转换为UUAP密码找回结构体
func OriginalOrderToUuapResetPwdOrder(weworkOrder map[string]interface{}) (weworkOrderDetails WeworkOrderResetUuapPwd) {
	if err := mapstructure.Decode(weworkOrder, &weworkOrderDetails); err != nil {
		log.Log().Error("原始工单转换错误:%v", err)
	}
	return
}
