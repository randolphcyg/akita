package order

import (
	"encoding/json"
	"strings"

	"github.com/goinggo/mapstructure"
	"github.com/pkg/errors"

	"gitee.com/RandolphCYG/akita/pkg/log"
)

// WeworkOrder 企业微信工单结构体
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

// ParseRawOrder 解析企业微信原始工单
func ParseRawOrder(rawInfo interface{}) (orderData map[string]interface{}, err error) {
	var weworkOrder WeworkOrder
	// 反序列化工单详情
	if err = mapstructure.Decode(rawInfo, &weworkOrder); err != nil {
		log.Log.Error("Fail to deserialize map to struct, err: ", err)
		return
	}

	// 判断工单状态
	if weworkOrder.SpStatus != 2 { // 工单不是通过状态
		log.Log.Error("The ticket is not approved!")
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
				tempSelectors := make([]string, len(con.Value.Selector.Options))
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
		case "File":
			t, _ := json.Marshal(con.Value.Files)
			var tem []struct{ FileId string }
			json.Unmarshal(t, &tem)
			var temp []string
			for _, file := range tem {
				temp = append(temp, file.FileId)
			}
			orderData[con.Title[0].Text] = temp
		// 明细的处理
		case "Table":
			temps := make([]map[string]interface{}, 0)
			for _, u := range con.Value.Children {
				temp := make(map[string]interface{})
				for _, c := range u.List {
					switch c.Control {
					case "Number":
						temp[c.Title[0].Text] = c.Value.NewNumber
					case "Text": // 明细文本
						temp[c.Title[0].Text] = strings.ToLower(strings.TrimSpace(c.Value.Text)) // 字符串去除空格并转为小写
					case "Textarea": // 明细多行文本
						temp[c.Title[0].Text] = strings.ToLower(strings.TrimSpace(c.Value.Text)) // 字符串去除空格并转为小写
					case "Date":
						temp[con.Title[0].Text] = c.Value.Date.STimestamp
					case "Selector": // 明细选择
						if c.Value.Selector.Type == "multi" { // 多选
							tempSel := make([]string, 0)
							for _, v := range c.Value.Selector.Options {
								tempSel = append(tempSel, v.Value[0].Text)
							}
							temp[c.Title[0].Text] = tempSel
						} else { // 单选
							temp[c.Title[0].Text] = c.Value.Selector.Options[0].Value[0].Text
						}
					case "Contact":
						temp[c.Title[0].Text] = c.Value.Members
					case "File": // 明细文件
						t, _ := json.Marshal(c.Value.Files)
						var tem []struct{ FileId string }
						json.Unmarshal(t, &tem)
						var te []string
						for _, file := range tem {
							te = append(te, file.FileId)
						}
						temp[c.Title[0].Text] = te
					}
				}
				temps = append(temps, temp)
			}
			orderData[con.Title[0].Text] = temps
		default:
			err = errors.New("包含未处理工单项类型【" + con.Control + "】请及时补充后端逻辑")
			return
		}
	}
	return
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

// RawToAccountsRegister 原始工单转换为账号注册工单结构体
func RawToAccountsRegister(weworkOrder map[string]interface{}) (orderDetails AccountsRegister) {
	if _, ok := weworkOrder["姓名"]; ok {
		var temp AccountsRegisterSingle
		if err := mapstructure.Decode(weworkOrder, &temp); err != nil {
			log.Log.Error("Fail to convert raw order, err: ", err)
		}
		// 将单转多
		orderDetails.Partyid = temp.Partyid
		orderDetails.SpName = temp.SpName
		orderDetails.Userid = temp.Userid
		orderDetails.Users = append(orderDetails.Users, Applicant{
			DisplayName:   temp.DisplayName,
			Eid:           temp.Eid,
			Mobile:        temp.Mobile,
			Mail:          temp.Mail,
			Company:       temp.Company,
			InitPlatforms: temp.InitPlatforms,
		})
	} else {
		if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
			log.Log.Error("Fail to convert raw order, err: ", err)
		}
	}
	return
}

// UuapPwdRetrieve UUAP密码找回 工单详情
type UuapPwdRetrieve struct {
	SpName      string `mapstructure:"spName"`
	Userid      string `mapstructure:"userid"`
	DisplayName string `mapstructure:"姓名"`
	Eid         string `mapstructure:"工号"`
}

// RawToUuapPwdRetrieve 原始工单转换为UUAP密码找回结构体
func RawToUuapPwdRetrieve(weworkOrder map[string]interface{}) (orderDetails UuapPwdRetrieve) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Log.Error("Fail to convert raw order, err: ", err)
	}
	return
}

// UuapDisable 账号注销 工单详情
type UuapDisable struct {
	SpName      string `mapstructure:"spName"`
	Userid      string `mapstructure:"userid"`
	DisplayName string `mapstructure:"姓名"`
	Eid         string `mapstructure:"工号"`
}

// RawToUuapPwdDisable 原始工单转换为账号注销结构体
func RawToUuapPwdDisable(weworkOrder map[string]interface{}) (orderDetails UuapDisable) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Log.Error("Fail to convert raw order, err: ", err)
	}
	return
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

// RawToAccountsRenewal 原始工单转换为账号续期结构体
func RawToAccountsRenewal(weworkOrder map[string]interface{}) (orderDetails AccountsRenewal) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Log.Error("Fail to convert raw order, err: ", err)
	}
	return
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

// RawToC7nAuthority 原始工单转换为 c7n项目权限 结构体
func RawToC7nAuthority(weworkOrder map[string]interface{}) (orderDetails C7nAuthority) {
	if err := mapstructure.Decode(weworkOrder, &orderDetails); err != nil {
		log.Log.Error("Fail to convert raw order, err: ", err)
	}
	return
}
