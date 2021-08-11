package wework

import (
	"errors"
	"fmt"
	"strconv"

	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/conn"
	"gitee.com/RandolphCYG/akita/internal/service/user"
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Order 企业微信工单查询条件
type Order struct {
	// 用户对象类
	SpNo string `json:"sp_no"`
}

// 工单总入口
func (service *Order) HandleOrders(o *Order) (err error) {
	// 判断工单是否存在 若存在则不处理，若不存在则保存一份 处理失败情况要记录到表中
	result, orderExecuteRecord := model.FetchOrder(o.SpNo)
	if result.RowsAffected == 1 && orderExecuteRecord.ExecuteStatus {
		err = errors.New("thanks,tabby! 【" + o.SpNo + "】该工单已经处理过，忽略此次操作~")
		log.Warning(err)
		return
	}

	// 实例化 API 类 企业微信配置保存到数据库，系统初始化时加载到环境变量中
	corpAPI := api.NewCorpAPI(model.WeworkOrderCfg.CorpId, model.WeworkOrderCfg.AppSecret)
	// 获取审批工单详情
	response, err := corpAPI.GetApprovalDetail(map[string]interface{}{
		"sp_no": o.SpNo,
	})
	if err != nil {
		log.Error("Fail to get approval detail, err: ", err)
		return
	}

	// 解析企业微信原始工单
	if _, ok := response["info"]; !ok {
		log.Error("Fail to parse raw order, order receipt has no field [info]!")
		return
	}

	orderData, err := order.ParseRawOrder(response["info"])
	if err != nil {
		log.Error(err)
		return
	}

	// 工单分流 将原始工单结构体转换为对应要求工单数据
	switch orderData["spName"] {
	case "统一账号注册":
		{
			weworkOrder := order.RawToAccountsRegister(orderData)
			err = handleOrderAccountsRegister(weworkOrder)
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
	case "测试工单":
		{
			weworkOrder := order.RawToUuapRenewal(orderData)
			err = handleOrderUuapRenewal(weworkOrder)
		}
	case "猪齿鱼项目权限":
		{
			weworkOrder := order.RawToC7nAuthority(orderData)
			err = handleOrderC7nAuthority(weworkOrder)
		}
	default:
		err = errors.New("Akita server has no handdler with this kind of order, please handle it manually!")
	}
	// 统一处理工单的处理情况
	if result.RowsAffected == 1 && !orderExecuteRecord.ExecuteStatus { // 非首次执行 重试
		if err != nil { // 工单执行出现错误
			log.Error("Fail to handle previous wework order, err: ", err)
			model.UpdateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.UpdateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	} else if result.RowsAffected == 0 { // 首次执行
		if err != nil { // 工单执行出现错误
			log.Error("Fail to handle fresh wework order, err: ", err)
			model.CreateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.CreateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	}
	return
}

// handleOrderAccountsRegister 统一账号注册 工单
func handleOrderAccountsRegister(o order.WeworkOrderDetailsAccountsRegister) (err error) {
	// 支持处理多个申请者
	for _, applicant := range o.Users {
		var expire int64
		var isOutsideComp bool
		var sam, dn, weworkExpireStr string
		var companyTypes map[string]model.CompanyType
		displayName := []rune(applicant.DisplayName)
		cn := string(displayName) + applicant.Eid
		companyTypes, err = conn.Str2CompanyTypes(bootstrap.LdapField.CompanyType)
		if err != nil {
			log.Error("Fail to deserialize str, err: ", err)
			return
		}
		// 取内外部公司前缀映射
		if v, ok := companyTypes[applicant.Company]; ok {
			isOutsideComp = v.IsOuter
		} else { // 若没有这个公司返回报错
			log.Error(errors.New("无此公司,请到LDAP服务器增加对应公司！"))
			return
		}

		// 不同公司个性化用户名与OU
		if isOutsideComp {
			sam = companyTypes[applicant.Company].Prefix + applicant.Eid // 用户名带前缀
			dn = "CN=" + cn + ",OU=" + applicant.Company + "," + bootstrap.LdapField.BaseDnOuter
			expire = ldap.ExpireTime(int64(90)) // 90天过期
			weworkExpireStr = util.ExpireStr(90)
		} else { // 公司内部人员默认放到待分配区 后面每天程序自动将用户架构刷新
			sam = applicant.Eid
			dn = "CN=" + cn + "," + bootstrap.LdapField.BaseDnToBeAssigned
			expire = ldap.ExpireTime(int64(-1)) // 永不过期
		}
		// 组装LDAP用户数据
		userInfos := &ldap.LdapAttributes{
			Dn:           dn,
			Num:          sam,
			Sam:          sam,
			AccountCtl:   "544",
			Expire:       expire,
			Sn:           string(displayName[0]),
			PwdLastSet:   "0",
			DisplayName:  string(displayName),
			GivenName:    string(displayName[1:]),
			Email:        applicant.Mail,
			Phone:        applicant.Mobile,
			Company:      applicant.Company,
			WeworkExpire: weworkExpireStr,
		}
		fmt.Println(userInfos)

		// 将平台切片转为map 用于判断是否存在某平台
		platforms := make(map[string]int)
		for i, v := range applicant.InitPlatforms {
			platforms[v] = i
		}

		// 待进行操作判断逻辑
		if _, ok := platforms["UUAP"]; ok {
			// UUAP操作
			err = user.CreateLdapUser(o, userInfos)
			if err != nil {
				log.Error(err)
			}
		}

		if _, ok := platforms["企业微信"]; ok {
			// 执行生成 企业微信账号 操作
			err = CreateUser(userInfos)
			if err != nil {
				log.Error(err)
			}
		}

		if _, ok := platforms["猪齿鱼"]; ok {
			if _, ok := platforms["UUAP"]; !ok {
				// 确保需要猪齿鱼的有UUAP 若无则创建
				err = user.CreateLdapUser(o, userInfos)
				if err != nil {
					log.Error(err)
				}
			}

			// 执行初始化 猪齿鱼 操作
			err = c7n.UpdateC7nUsers()                                             // 更新ldap用户
			c7nUser, _ := c7n.FtechC7nUser(applicant.DisplayName)                  // 将新ldap用户添加到默认空项目
			role, _ := c7n.FetchC7nRoles("项目成员")                                   // 获取项目成员角色的ID
			err = c7n.AssignC7nUserProjectRole("4", c7nUser.Id, []string{role.Id}) // 分配角色
			if err != nil {
				logrus.Error("Fail to assign new user c7n default project!", err)
			}
		}

		if _, ok := platforms["UVPN"]; ok {
			if _, ok := platforms["UUAP"]; !ok {
				// 确保需要UVPN的有UUAP 若无则创建
				err = user.CreateLdapUser(o, userInfos)
				if err != nil {
					log.Error(err)
				}
			}

			// TODO 执行初始化 UVPN 操作
		}
	}
	return
}

// UUAP密码找回 工单
func handleOrderUuapPwdRetrieve(o order.WeworkOrderDetailsUuapPwdRetrieve) (err error) {
	user := &ldap.LdapAttributes{
		Num:         o.Eid,
		DisplayName: o.DisplayName,
	}

	sam, newPwd, err := user.RetrievePwd()
	if err != nil {
		log.Error("Fail to retrieve pwd, err: ", err)
	}

	// 创建成功发送企业微信消息
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	retrieveUuapPwdWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_pwd_retrieve")
	if err != nil {
		log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = corpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(retrieveUuapPwdWeworkMsgTemplate, o.SpName, user.DisplayName, sam, newPwd),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg, err: ", err)
	}
	log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + o.DisplayName + "】工号【" + o.Eid + "】状态【密码找回】")
	return
}

// UUAP账号注销 工单
func handleOrderUuapDisable(o order.WeworkOrderDetailsUuapDisable) (err error) {
	user := &ldap.LdapAttributes{
		Num:         o.Eid,
		DisplayName: o.DisplayName,
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
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, o.SpName, user.DisplayName),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg, err: ", err)
	}
	log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + o.DisplayName + "】工号【" + o.Eid + "】状态【注销】")
	return
}

// UUAP账号续期 工单
func handleOrderUuapRenewal(o order.WeworkOrderDetailsUuapRenewal) (err error) {
	// 支持处理多个申请者
	for _, applicant := range o.Users {
		fmt.Println(applicant)

		// 将平台切片转为map 用于判断是否存在某平台
		platforms := make(map[string]int)
		for i, v := range applicant.Platforms {
			platforms[v] = i
		}

		// 待进行操作判断逻辑
		if _, ok := platforms["UUAP"]; ok {
			// UUAP续期操作
			RenewalUuap(o, applicant)
		}

		if _, ok := platforms["企业微信"]; ok {
			// 企业微信续期操作 TODO 批量的话 根据工号锁定用户
			weworkUser, _ := FetchUser(applicant.Eid)
			expireDays, _ := strconv.Atoi(applicant.Days)
			RenewalUser(weworkUser.Userid, applicant, expireDays)
		}
	}

	return
}

// RenewalUuap uuap续期
func RenewalUuap(o order.WeworkOrderDetailsUuapRenewal, applicant order.RenewalApplicant) (err error) {
	days, _ := strconv.ParseInt(applicant.Days, 10, 64)
	user := &ldap.LdapAttributes{
		Num:         applicant.Eid,
		DisplayName: applicant.DisplayName,
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
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, o.SpName, user.DisplayName, applicant.Days),
		},
	})
	if err != nil {
		log.Error("Fail to send wework msg, err: ", err)
	}
	log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + applicant.DisplayName + "】工号【" + applicant.Eid + "】状态【续期" + applicant.Days + "天】")
	return
}

// c7n权限处理 TODO 逻辑未完成
func handleOrderC7nAuthority(order order.WeworkOrderDetailsC7nAuthority) (err error) {
	// project, _ := c7n.FetchC7nProject("XXX")
	// fmt.Println(project)
	// user, _ := c7n.FtechC7nUser("XXX")
	// fmt.Println(user)

	// var c7nRoleIds []string
	// // TODO 需要该项目该用户的旧角色作增操作
	// role, _ := c7n.FetchC7nRoles("项目成员")
	// c7nRoleIds = append(c7nRoleIds, role.Id)
	// fmt.Println(c7nRoleIds)

	// c7n.AssignC7nUserProjectRole(strconv.Itoa(project.Id), user.Id, c7nRoleIds)

	return
}
