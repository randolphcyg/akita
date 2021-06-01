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

// 企业微信工单处理总入口
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
		return serializer.Err(-1, "Occur error when handle original wework order:%v", err)
	}
	// 工单分流 将原始工单结构体转换为对应要求工单数据
	switch orderData["spName"] {
	case "UUAP账号注册":
		{
			weworkOrder := order.OriginalToUuapRegister(orderData)
			handleOrderUuapRegister(weworkOrder)
		}
	case "UUAP密码找回":
		{
			weworkOrder := order.OriginalToUuapResetPwd(orderData)
			handleOrderUuapPwdReset(weworkOrder)
		}
	default:
		log.Log().Error("无任何匹配工单,请检查tabby工单名称列表是否有此工单~")
	}

	return serializer.Response{Data: 0, Msg: "thanks,tabby! 工单处理中..."}
}

// UUAP账号注册 工单
func handleOrderUuapRegister(order order.WeworkOrderDetailsUuapRegister) (err error) {
	// 组装LDAP用户数据
	name := []rune(order.Name)
	user := &ldap.LdapAttributes{
		Dn:          "CN=" + string(name) + order.Eid + "," + ldap.DepartToDn(order.Depart),
		Num:         order.Eid,
		Sam:         order.Eid,
		AccountCtl:  "544",
		Expire:      ldap.ExpireTime(int64(-1)), // 永不过期
		Sn:          string(name[0]),
		PwdLastSet:  "0",
		DisplayName: string(name),
		GivenName:   string(name[1:]),
		Email:       order.Mail,
		Phone:       order.Mobile,
		Company:     strings.Split(order.Depart, ".")[0],
		Depart:      strings.Split(order.Depart, ".")[strings.Count(order.Depart, ".")],
		Title:       order.Title,
	}

	// 创建LDAP用户 生成初始密码
	pwd, err := ldap.AddUser(user)
	if err != nil {
		log.Log().Error("User already exist: %s", err)
		return
	}

	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	// 创建成功发送企业微信消息 将企业微信MD消息模板缓存在redis
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_register")
	if err != nil {
		log.Log().Error("读取企业微信消息模板错误:%v", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, order.SpName, order.Eid, pwd),
		},
	})
	if err != nil {
		log.Log().Error("发送企业微信通知错误：%v", err)
	}
	log.Log().Info("已经发送企业微信消息给【" + order.Userid + "】")
	return
}

// UUAP密码找回 工单
func handleOrderUuapPwdReset(order order.WeworkOrderDetailsUuapPwdReset) (err error) {
	log.Log().Info("%v", order)

	user := &ldap.LdapAttributes{
		Num:         order.Uuap,
		Name:        order.Name,
		DisplayName: order.Name,
	}

	sam, newPwd, err := user.ResetPwd()
	if err != nil {
		log.Log().Error("%s", err)
	}
	log.Log().Info(newPwd)

	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	// 创建成功发送企业微信消息 将企业微信MD消息模板缓存在redis
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_pwd_reset")
	if err != nil {
		log.Log().Error("读取企业微信消息模板错误:%v", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, order.SpName, sam, newPwd),
		},
	})
	if err != nil {
		log.Log().Error("发送企业微信通知错误：%v", err)
	}
	log.Log().Info("已经发送企业微信消息给【" + order.Userid + "】")
	return
}
