package order

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
	SpName   string `mapstructure:"spName"`
	Partyid  string `mapstructure:"partyid"`
	Userid   string `mapstructure:"userid"`
	Remarks  string `mapstructure:"备注"`
	Name     string `mapstructure:"姓名"`
	Title    string `mapstructure:"岗位"`
	Eid      string `mapstructure:"工号"`
	Mobile   string `mapstructure:"手机"`
	Contacts []struct {
		Userid string `mapstructure:"userid"`
		Name   string `mapstructure:"name"`
	} `mapstructure:"相同角色同事"`
	Role          []string `mapstructure:"角色"`
	ActExpireDate string   `mapstructure:"账号到期时间"`
	ActExpire     string   `mapstructure:"过期时间"`
	Mail          string   `mapstructure:"邮箱"`
	Depart        string   `mapstructure:"部门"`
}
