package order

import (
	"strings"

	"gitee.com/RandolphCYG/akita/internal/model"
	"github.com/goinggo/mapstructure"
	log "github.com/sirupsen/logrus"
)

// 全局变量
var (
	WeworkCfg *model.WeworkCfg
)

// 企业微信工单结构体
type WeworkOrder struct {
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
					file_id string `mapstructure:"FileId"`
				}
				// 明细控件（control参数为Table）
				Children []struct {
					Id      string `mapstructure:"id"`
					Control string `mapstructure:"control"`
					Title   []struct {
						Text string `mapstructure:"text"`
						Lang string `mapstructure:"lang"`
					} `mapstructure:"title"`

					Value []struct {
						Children []struct {
							List []struct {
								Control string `mapstructure:"control"`
								Id      string `mapstructure:"id"`
								Title   []struct {
									Text string `mapstructure:"text"`
									Lang string `mapstructure:"lang"`
								} `mapstructure:"title"`
								// 		Value   []struct {
								// 			Text string `mapstructure:"text"`
								// 		} `mapstructure:"value"`
							} `mapstructure:"list"`
						} `mapstructure:"children"`
					} `mapstructure:"value"`
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

// 解析企业微信原始工单
func ParseRawOrder(rawInfo interface{}) (orderData map[string]interface{}, err error) {
	var weworkOrder WeworkOrder
	// 反序列化工单详情
	if err = mapstructure.Decode(rawInfo, &weworkOrder); err != nil {
		log.Error("Fail to deserialize map to struct,err: ", err)
		return
	}

	// 判断工单状态 TODO 可考虑拓展操作
	if weworkOrder.SpStatus != 2 {
		log.Error("The ticket is not approved, refused to handdle!")
		return
	}

	// 清洗工单
	orderData = make(map[string]interface{})
	orderData["spName"] = weworkOrder.SpName
	orderData["partyid"] = weworkOrder.Applyer.Partyid
	orderData["userid"] = weworkOrder.Applyer.Userid
	// 抄送人
	if len(weworkOrder.Notifyer) >= 1 {
		orderData["notifyer"] = weworkOrder.Notifyer
	}
	// 处理工单数据
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
		case "Table":
			log.Info(con)
		default:
			log.Error("包含未处理工单项类型【" + con.Control + "】请及时补充后端逻辑!")
		}
	}
	return
}

// UUAP账号注册 工单详情
type WeworkOrderDetailsUuapRegister struct {
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

// 原始工单转换为UUAP注册工单结构体
func RawToUuapRegister(weworkOrder map[string]interface{}) (orderDetails WeworkOrderDetailsUuapRegister) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Error("Fail to convert raw oder,err: ", err)
	}
	return
}

// UUAP密码找回 工单详情
type WeworkOrderDetailsUuapPwdRetrieve struct {
	SpName string `mapstructure:"spName"`
	Userid string `mapstructure:"userid"`
	Name   string `mapstructure:"姓名"`
	Uuap   string `mapstructure:"工号"`
}

// 原始工单转换为UUAP密码找回结构体
func RawToUuapPwdRetrieve(weworkOrder map[string]interface{}) (orderDetails WeworkOrderDetailsUuapPwdRetrieve) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Error("Fail to convert raw oder,err: ", err)
	}
	return
}

// UUAP账号注销 工单详情
type WeworkOrderDetailsUuapDisable struct {
	SpName string `mapstructure:"spName"`
	Userid string `mapstructure:"userid"`
	Name   string `mapstructure:"姓名"`
	Uuap   string `mapstructure:"工号"`
}

// 原始工单转换为UUAP账号注销结构体
func RawToUuapPwdDisable(weworkOrder map[string]interface{}) (orderDetails WeworkOrderDetailsUuapDisable) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Error("Fail to convert raw oder,err: ", err)
	}
	return
}

// UUAP账号续期 工单详情
type WeworkOrderDetailsUuapRenewal struct {
	SpName string `mapstructure:"spName"`
	Userid string `mapstructure:"userid"`
	Name   string `mapstructure:"姓名"`
	Uuap   string `mapstructure:"工号"`
	Days   string `mapstructure:"续期天数"`
}

// 原始工单转换为UUAP账号续期结构体
func RawToUuapRenewal(weworkOrder map[string]interface{}) (orderDetails WeworkOrderDetailsUuapRenewal) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Error("Fail to convert raw oder,err: ", err)
	}
	return
}
