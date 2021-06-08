package order

import (
	"fmt"
	"strconv"
	"strings"

	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/conn"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
	log "github.com/sirupsen/logrus"
)

// Order 企业微信工单查询条件
type Order struct {
	// 用户对象类
	SpNo string `json:"sp_no"`
}

// 企业微信工单处理总入口
func (service *Order) HandleOrders(o *Order) serializer.Response {
	// 判断工单是否存在 若存在则不处理，若不存在则保存一份 处理失败情况要记录到表中
	result, orderExecuteRecord := model.FetchOrder(o.SpNo)
	if result.RowsAffected == 1 && orderExecuteRecord.ExecuteStatus {
		log.Warning("【" + o.SpNo + "】该工单已经处理过，忽略此次操作~")
		return serializer.Response{Data: 0, Msg: "thanks,tabby! 【" + o.SpNo + "】该工单已经处理过，忽略此次操作~"}
	}

	// 实例化 API 类 企业微信配置保存到数据库，系统初始化时加载到环境变量中
	corpAPI := api.NewCorpAPI(model.WeworkOrderCfg.CorpId, model.WeworkOrderCfg.AppSecret)
	// 获取审批工单详情
	response, err := corpAPI.GetApprovalDetail(map[string]interface{}{
		"sp_no": o.SpNo,
	})
	if err != nil {
		log.Error("Fail to get approval detail,err: ", err)
	}
	// 解析企业微信原始工单
	orderData, err := order.ParseRawOrder(response["info"])
	if err != nil {
		log.Error(err)
		return serializer.Err(-1, "Fail to handle raw wework order,err: ", err)
	}
	// 工单分流 将原始工单结构体转换为对应要求工单数据
	switch orderData["spName"] {
	case "UUAP账号注册":
		{
			weworkOrder := order.RawToUuapRegister(orderData)
			err = handleOrderUuapRegister(weworkOrder)
		}
	case "UUAP密码找回":
		{
			weworkOrder := order.RawToUuapPwdRetrieve(orderData)
			err = handleOrderUuapPwdRetrieve(weworkOrder)
		}
	case "UUAP账号注销":
		{
			weworkOrder := order.RawToUuapPwdDisable(orderData)
			err = handleOrderUuapDisable(weworkOrder)
		}
	case "UUAP账号续期":
		{
			weworkOrder := order.RawToUuapRenewal(orderData)
			err = handleOrderUuapRenewal(weworkOrder)
		}
	default:
		log.Error("Akita server has no handdler with this kind of order, please handle it manually!")
	}
	// 统一处理工单处理情况
	if result.RowsAffected == 1 && !orderExecuteRecord.ExecuteStatus { // 非首次执行 重试
		if err != nil { // 工单执行出现错误
			log.Error("Fail to handle previous wework order,err: ", err)
			model.UpdateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.UpdateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	} else if result.RowsAffected == 0 { // 首次执行
		if err != nil { // 工单执行出现错误
			log.Error("Fail to handle fresh wework order,err: ", err)
			model.CreateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.CreateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	}

	return serializer.Response{Data: 0, Msg: "thanks,tabby! 工单处理中..."}
}

// UUAP账号注册 工单
func handleOrderUuapRegister(order order.WeworkOrderDetailsUuapRegister) (err error) {
	// 组装LDAP用户数据
	companyTypes, err := conn.Str2CompanyTypes(bootstrap.LdapField.CompanyType)
	if err != nil {
		log.Error("Fail to deserialize str,err: ", err)
	}

	companyName := strings.Split(order.Depart, ".")[0]
	isCompanyOutside := companyTypes[companyName].IsOuter
	prefix := companyTypes[companyName].Prefix
	dn := ""
	name := []rune(order.Name)
	sam := order.Eid
	expire := ldap.ExpireTime(int64(-1)) // 永不过期
	// 外部公司个性化用户名与OU位置
	if isCompanyOutside == 1 {
		sam = prefix + sam
		dn = "CN=" + string(name) + order.Eid + "," + "OU=" + companyName + "," + bootstrap.LdapField.BaseDnOuter
		expire = ldap.ExpireTime(int64(90)) // 90天过期
	} else { // 本公司默认逻辑
		dn = "CN=" + string(name) + order.Eid + "," + ldap.DepartToDn(order.Depart)
	}

	user := &ldap.LdapAttributes{
		Dn:          dn,
		Num:         sam,
		Sam:         sam,
		AccountCtl:  "544",
		Expire:      expire,
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
		log.Error("User already exist: ", err)
		return
	}

	// 创建成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_register")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
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
		log.Error("Fail to send wework msg,err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Info("企业微信消息发送成功！工单【" + order.SpName + "】用户【" + order.Userid + "】")
	return
}

// UUAP密码找回 工单
func handleOrderUuapPwdRetrieve(order order.WeworkOrderDetailsUuapPwdRetrieve) (err error) {
	user := &ldap.LdapAttributes{
		Num:         order.Uuap,
		DisplayName: order.Name,
	}

	sam, newPwd, err := user.RetrievePwd()
	if err != nil {
		log.Error("Fail to retrieve pwd,err: ", err)
	}

	// 创建成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	createUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_pwd_retrieve")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(createUuapWeworkMsgTemplate, order.SpName, user.DisplayName, sam, newPwd),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg,err: ", err)
	}
	log.Info("企业微信消息发送成功！工单【" + order.SpName + "】用户【" + order.Userid + "】")
	return
}

// UUAP账号注销 工单
func handleOrderUuapDisable(order order.WeworkOrderDetailsUuapDisable) (err error) {
	user := &ldap.LdapAttributes{
		Num:         order.Uuap,
		DisplayName: order.Name,
	}

	err = user.Disable()
	if err != nil {
		log.Error(err)
		return
	}

	// 续期成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	renewalUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_disable")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, order.SpName, user.DisplayName),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg,err: ", err)
	}
	log.Info("企业微信消息发送成功！工单【" + order.SpName + "】用户【" + order.Userid + "】")
	return
}

// UUAP账号续期 工单
func handleOrderUuapRenewal(order order.WeworkOrderDetailsUuapRenewal) (err error) {
	days, _ := strconv.ParseInt(order.Days, 10, 64)
	user := &ldap.LdapAttributes{
		Num:         order.Uuap,
		DisplayName: order.Name,
		Expire:      ldap.ExpireTime(days),
	}

	err = user.Renewal()
	if err != nil {
		log.Error(err)
		return
	}

	// 续期成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	renewalUuapWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_uuap_renewal")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  order.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, order.SpName, user.DisplayName, order.Days),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg,err: ", err)
	}
	log.Info("企业微信消息发送成功！工单【" + order.SpName + "】用户【" + order.Userid + "】")
	return
}
