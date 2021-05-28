package order

import (
	"fmt"
	"strings"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
)

// Order 企业微信工单查询条件
type Order struct {
	// 用户对象类
	SpNo string `json:"sp_no"`
}

// 处理工单
func (service *Order) HandleOrders(o *Order) serializer.Response {
	// 实例化 API 类 企业微信配置保存到数据库，系统初始化时加载到环境变量中
	corpAPI := api.NewCorpAPI(model.WeworkOrderCfg.CorpId, model.WeworkOrderCfg.AppSecret)
	// 获取审批工单详情
	response, err := corpAPI.GetApprovalDetail(map[string]interface{}{
		"sp_no": o.SpNo,
	})
	if err != nil {
		log.Log().Error("Occur error when get approval detail:%v", err)
	}
	// 将企业微信原始工单转换为对应工单
	orderData, err := order.RawOrderToObj(response["info"])
	if err != nil {
		log.Log().Error("%v", err)
		return serializer.Err(-1, "处理原始企业微信工单报错!", err)
	}

	// 将原始工单结构体转换为对应要求工单数据 工单分流
	if orderData["spName"] == "UUAP账号申请工单" {
		weworkOrder := order.OriginalOrderToUuapCreateOrder(orderData)
		orderUuapCreate(weworkOrder)
	} else if orderData["spName"] == "UUAP密码找回" {
		weworkOrder := order.OriginalOrderToUuapResetPwdOrder(orderData)
		orderUuapPwdReset(weworkOrder)
	}

	return serializer.Response{Data: 0, Msg: "thanks,tabby! 工单处理中..."}
}

// 工单-UUAP重置密码 TODO：具体实现贯通测试
func orderUuapPwdReset(o order.WeworkOrderResetUuapPwd) (err error) {
	log.Log().Info(o.Uuap)
	return
}

// 工单-UUAP创建用户
func orderUuapCreate(weworkOrderDetails order.WeworkOrderDetails) (err error) {
	// 组装LDAP用户数据
	name := []rune(weworkOrderDetails.Name)
	user := &ldap.LdapAttributes{
		Dn:          "CN=" + string(name) + weworkOrderDetails.Eid + "," + ldap.DepartToDn(weworkOrderDetails.Depart),
		Num:         weworkOrderDetails.Eid,
		Sam:         weworkOrderDetails.Eid,
		AccountCtl:  "544",
		Expire:      ldap.ExpireTime(int64(-1)), // 永不过期
		Sn:          string(name[0]),
		PwdLastSet:  "0",
		DisplayName: string(name),
		GivenName:   string(name[1:]),
		Email:       weworkOrderDetails.Mail,
		Phone:       weworkOrderDetails.Mobile,
		Company:     strings.Split(weworkOrderDetails.Depart, ".")[0],
		Depart:      strings.Split(weworkOrderDetails.Depart, ".")[strings.Count(weworkOrderDetails.Depart, ".")],
		Title:       weworkOrderDetails.Title,
	}

	// 创建LDAP用户 生成初始密码
	pwd, err := ldap.AddUser(user)
	if err != nil {
		log.Log().Error("User already exist: %s", err)
		return
	}

	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	// 创建成功发送企业微信消息 将企业微信MD消息模板缓存在redis
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "uuap_create_ww_template")
	if err != nil {
		log.Log().Error("读取企业微信消息模板错误:%v", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  weworkOrderDetails.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, weworkOrderDetails.Eid, pwd),
		},
	})
	if err != nil {
		log.Log().Error("发送企业微信通知错误：%v", err)
	}
	return
}
