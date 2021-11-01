package wework

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/ldapconn"
	"gitee.com/RandolphCYG/akita/internal/service/ldapuser"
	"gitee.com/RandolphCYG/akita/pkg/c7n"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
)

var (
	ErrCompanyNotExists = errors.New("无此公司,请到LDAP服务器增加对应公司！")
	ErrDeserialize      = errors.New("反序列化错误！")
	ErrFetchDB          = errors.New("查询数据库错误！")
	// 全局使用的公司前缀映射
	companyTypes map[string]model.CompanyType
)

// Order 企业微信工单查询条件
type Order struct {
	// 用户对象类
	SpNo string `json:"sp_no"`
}

// HandleOrders 企业微信工单总入口
func (o *Order) HandleOrders() (err error) {
	// 判断工单是否存在 若存在则不处理，若不存在则保存一份 处理失败情况要记录到表中
	result, orderExecuteRecord := model.FetchOrder(o.SpNo)
	if result.RowsAffected == 1 && orderExecuteRecord.ExecuteStatus {
		err = errors.New("thanks,tabby! [" + o.SpNo + "]该工单已经处理过，忽略此次操作")
		log.Log.Warning(err)
		return
	}

	// 获取审批工单详情
	response, err := model.CorpAPIOrder.GetApprovalDetail(map[string]interface{}{
		"sp_no": o.SpNo,
	})
	if err != nil {
		log.Log.Error("Fail to get approval detail, err: ", err)
		return
	}

	// 解析企业微信原始工单
	if _, ok := response["info"]; !ok {
		log.Log.Error("Fail to parse raw order, order receipt has no field [info]!")
		return
	}

	orderData, err := order.ParseRawOrder(response["info"])
	if err != nil {
		log.Log.Error(err)
		return
	}

	// 工单分流 将原始工单结构体转换为对应要求工单数据
	switch orderData["spName"] {
	case "账号注册":
		{
			weworkOrder := order.RawToAccountsRegister(orderData)
			err = handleOrderAccountsRegister(weworkOrder)
		}
	case "UUAP密码找回":
		{
			weworkOrder := order.RawToUuapPwdRetrieve(orderData)
			err = handleOrderUuapPwdRetrieve(weworkOrder)
		}
	case "账号注销":
		{
			weworkOrder := order.RawToUuapPwdDisable(orderData)
			err = handleOrderUuapDisable(weworkOrder)
		}
	case "账号续期":
		{
			weworkOrder := order.RawToAccountsRenewal(orderData)
			err = handleOrderAccountsRenewal(weworkOrder)
		}
	case "猪齿鱼项目权限":
		{
			weworkOrder := order.RawToC7nAuthority(orderData)
			err = handleOrderC7nAuthority(weworkOrder)
		}
	default:
		log.Log.Warning("UUAP server has no handler with this kind of order, please handle it manually!")
		return
	}
	// 统一处理工单处理情况
	if result.RowsAffected == 1 && !orderExecuteRecord.ExecuteStatus { // 非首次执行 重试
		if err != nil { // 工单执行出现错误
			log.Log.Error("Fail to handle previous wework order, err: ", err)
			model.UpdateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.UpdateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	} else if result.RowsAffected == 0 { // 首次执行
		if err != nil { // 工单执行出现错误
			log.Log.Error("Fail to handle fresh wework order, err: ", err)
			model.CreateOrder(o.SpNo, false, fmt.Sprintf("%v", err))
		} else {
			model.CreateOrder(o.SpNo, true, fmt.Sprintf("%v", err))
		}
	}
	return
}

// fetchLatestCompanyType 公司前缀映射查询
func fetchLatestCompanyType() (err error) {
	model.LdapFields, err = model.GetLdapFieldByConnUrl(model.LdapCfgs.ConnUrl)
	if err != nil {
		log.Log.Error(ErrFetchDB)
		return
	}

	companyTypes, err = ldapconn.Str2CompanyTypes(model.LdapFields.CompanyType)
	if err != nil {
		log.Log.Error(ErrDeserialize)
		return
	}
	return
}

// handleOrderAccountsRegister 账号注册 工单
func handleOrderAccountsRegister(o order.AccountsRegister) (err error) {
	// 支持处理多个申请者
	for _, applicant := range o.Users {
		var expire int64
		var isOutsideComp bool
		var sam, dn, weworkExpireStr string
		var weworkDepartId, probationFlag int
		displayName := []rune(applicant.DisplayName)
		cn := string(displayName) + applicant.Eid

		// 取内外部公司前缀映射
		companyTypes, err = ldapconn.Str2CompanyTypes(model.LdapFields.CompanyType)
		if err != nil {
			log.Log.Error(ErrDeserialize)
			return err
		}
		if v, ok := companyTypes[applicant.Company]; ok {
			isOutsideComp = v.IsOuter
		} else { // 若环境变量没找到当前用户公司则刷新环境变量重试
			err = fetchLatestCompanyType()
			if err != nil {
				log.Log.Error(ErrCompanyNotExists)
				return err
			}

			if v, ok := companyTypes[applicant.Company]; ok {
				isOutsideComp = v.IsOuter
			} else {
				log.Log.Error(ErrCompanyNotExists)
				return ErrCompanyNotExists
			}
		}

		// 不同公司个性化用户名与OU
		if isOutsideComp {
			sam = companyTypes[applicant.Company].Prefix + applicant.Eid // 用户名带前缀
			dn = "CN=" + cn + ",OU=" + applicant.Company + "," + model.LdapFields.BaseDnOuter
			expire = util.ExpireTime(int64(90)) // 90天过期
			weworkExpireStr = util.ExpireStr(90)
			weworkDepartId = 79 // 外部公司企业微信部门为合作伙伴
			probationFlag = 0
		} else { // 公司内部人员默认放到待分配区 后面每天程序自动将用户架构刷新
			sam = applicant.Eid
			dn = "CN=" + cn + "," + model.LdapFields.BaseDnToBeAssigned
			expire = util.ExpireTime(int64(-1)) // 永不过期
			weworkDepartId = 69                 // 本公司企业微信部门为待分配
			probationFlag = 1
		}
		// 组装LDAP用户数据
		userInfos := &ldapuser.LdapAttributes{
			Dn:             dn,
			Num:            sam,
			Sam:            sam,
			AccountCtl:     "544",
			Expire:         expire,
			Sn:             string(displayName[0]),
			PwdLastSet:     "0",
			DisplayName:    string(displayName),
			GivenName:      string(displayName[1:]),
			Email:          applicant.Mail,
			Phone:          applicant.Mobile,
			Company:        applicant.Company,
			WeworkExpire:   weworkExpireStr,
			WeworkDepartId: weworkDepartId,
			ProbationFlag:  probationFlag,
		}

		// 将平台切片转为map 用于判断是否存在某平台
		platforms := make(map[string]int)
		for i, v := range applicant.InitPlatforms {
			platforms[v] = i
		}

		// 待进行操作判断逻辑
		if _, ok := platforms["UUAP"]; ok {
			// UUAP操作
			err = ldapuser.CreateLdapUser(o, userInfos)
			if err != nil {
				log.Log.Error(err)
			}
		}

		if _, ok := platforms["企业微信"]; ok {
			weworkUser, err := FetchUser(userInfos.Num)
			if err == nil && weworkUser.Userid != "" && weworkUser.Name == userInfos.DisplayName {
				err := handleWeworkDuplicateRegister(o, userInfos)
				if err != nil {
					log.Log.Error("Fail to handle wework duplication register, ", err)
					return err
				}
			} else {
				// 执行生成 企业微信账号 操作
				err = CreateUser(userInfos)
				if err != nil {
					log.Log.Error("Fail to create user by wework order, ", err)
					model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "自动创建失败, "+err.Error())
				}

				recordMsg := "新用户 工单公司[" + userInfos.Company + "]分配至企微部门[" + strconv.Itoa(userInfos.WeworkDepartId) + "]"
				if userInfos.ProbationFlag == 1 {
					recordMsg += " Tag:[试用期员工]"
				}
				model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, recordMsg)
				log.Log.Info(recordMsg)
			}
		}

		if _, ok := platforms["猪齿鱼"]; ok {
			if _, ok := platforms["UUAP"]; !ok {
				// 确保需要猪齿鱼的有UUAP 若无则创建
				err = ldapuser.CreateLdapUser(o, userInfos)
				if err != nil {
					log.Log.Error(err)
				}
			}

			// 执行初始化 猪齿鱼 操作
			c7n.SyncUsers()                                                     // 更新ldap用户
			c7nUser, _ := c7n.FetchUser(applicant.DisplayName, applicant.Eid)   // 将新ldap用户添加到默认空项目
			role, _ := c7n.FetchRole("项目成员")                                    // 获取项目成员角色的ID
			err = c7n.AssignUserProjectRole("4", c7nUser.Id, []string{role.Id}) // 分配角色
			if err != nil {
				log.Log.Error("Fail to assign new user c7n default project!", err)
			}
		}

		if _, ok := platforms["UVPN"]; ok {
			if _, ok := platforms["UUAP"]; !ok {
				// 确保需要UVPN的有UUAP 若无则创建
				err = ldapuser.CreateLdapUser(o, userInfos)
				if err != nil {
					log.Log.Error(err)
				}
			}

			// TODO 执行初始化 UVPN 操作
		}
	}
	return
}

// handleWeworkDuplicateRegister 处理企业微信用户重复注册
func handleWeworkDuplicateRegister(o order.AccountsRegister, user *ldapuser.LdapAttributes) (err error) {
	duplicateRegisterWeworkUserWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_wework_user_duplicate_register")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}

	// 获取连接
	LdapConn, err := model.LdapPool.Get()
	if err != nil {
		log.Log.Error("Fail to get ldap connection, err: ", err)
		return
	}
	defer LdapConn.Close()

	entry, err := ldapuser.FetchUser(user)
	if err != nil {
		return
	}
	sam := entry.GetAttributeValue("sAMAccountName")

	_, err = model.CorpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(duplicateRegisterWeworkUserWeworkMsgTemplate, o.SpName, user.DisplayName, sam),
		},
	})
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单【" + o.SpName + "】用户【" + o.Userid + "】姓名【" + user.DisplayName + "】工号【" + user.Num + "】状态【已注册过的企业微信用户】")
	return
}

// handleOrderUuapPwdRetrieve UUAP密码找回 工单
func handleOrderUuapPwdRetrieve(o order.UuapPwdRetrieve) (err error) {
	user := &ldapuser.LdapAttributes{
		Num:         o.Eid,
		DisplayName: o.DisplayName,
	}

	sam, newPwd, err := user.RetrievePwd()
	if err != nil {
		return
	}

	// 创建成功发送企业微信消息
	retrieveUuapPwdWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_pwd_retrieve")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = model.CorpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(retrieveUuapPwdWeworkMsgTemplate, o.SpName, user.DisplayName, sam, newPwd),
		},
	})
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + o.DisplayName + "]工号[" + o.Eid + "]状态[密码找回]")
	return
}

// handleOrderUuapDisable 账号注销 工单
func handleOrderUuapDisable(o order.UuapDisable) (err error) {
	user := &ldapuser.LdapAttributes{
		Num:         o.Eid,
		DisplayName: o.DisplayName,
	}

	err = user.Disable()
	if err != nil {
		log.Log.Error(err)
		return
	}

	// 续期成功发送企业微信消息
	renewalUuapWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_uuap_disable")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = model.CorpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, o.SpName, user.DisplayName),
		},
	})
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + o.DisplayName + "]工号[" + o.Eid + "]状态[注销]")
	return
}

// handleOrderAccountsRenewal 账号续期 工单
func handleOrderAccountsRenewal(o order.AccountsRenewal) (err error) {
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
			err := RenewalUuap(o, applicant)
			if err != nil {
				return err
			}
		}

		if _, ok := platforms["企业微信"]; ok {
			// 企业微信续期操作
			weworkUser, _ := FetchUser(applicant.Eid)
			expireDays, _ := strconv.Atoi(applicant.Days)
			err := RenewalUser(weworkUser.Userid, applicant, expireDays)
			if err != nil {
				return err
			}
		}
	}

	return
}

// RenewalUuap 续期
func RenewalUuap(o order.AccountsRenewal, applicant order.RenewalApplicant) (err error) {
	days, _ := strconv.ParseInt(applicant.Days, 10, 64)
	user := &ldapuser.LdapAttributes{
		Num:         applicant.Eid,
		DisplayName: applicant.DisplayName,
		Expire:      util.ExpireTime(days),
	}

	err = user.Renewal()
	if err != nil {
		handleWeworkOrderFindUserErr(o, applicant.DisplayName, applicant.Eid)
		return
	}

	// 续期成功发送企业微信消息
	renewalUuapWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_uuap_renewal")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}
	_, err = model.CorpAPIMsg.MessageSend(map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalUuapWeworkMsgTemplate, o.SpName, user.DisplayName, applicant.Days),
		},
	})
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + applicant.DisplayName + "]工号[" + applicant.Eid + "]状态[续期" + applicant.Days + "天]")
	return
}

// handleWeworkOrderFindUserErr 处理未找到企微用户错误
func handleWeworkOrderFindUserErr(o order.AccountsRenewal, name, eid string) {
	c7nFindProjectErrMsgTemplate, _ := cache.HGet("wework_msg_templates", "wework_template_wework_find_user_err")
	msg := map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(c7nFindProjectErrMsgTemplate, o.SpName, name, name, eid),
		},
	}
	_, err := model.CorpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + name + "]工号[" + eid + "]状态[c7n项目权限|未找到企微用户,姓名: " + name + " 工号: " + eid + "]")
}

// handleC7nOrderFindUserErr 处理未找到c7n用户错误
func handleC7nOrderFindUserErr(o order.C7nAuthority, name, eid string) {
	c7nFindProjectErrMsgTemplate, _ := cache.HGet("wework_msg_templates", "wework_template_c7n_find_user_err")
	msg := map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(c7nFindProjectErrMsgTemplate, o.SpName, o.DisplayName, name, eid),
		},
	}
	_, err := model.CorpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + o.DisplayName + "]工号[" + o.Eid + "]状态[c7n项目权限|未找到c7n用户,姓名: " + name + " 工号: " + eid + "]")
}

// handleC7nOrderFindProjectErr 处理未找到c7n项目错误
func handleC7nOrderFindProjectErr(o order.C7nAuthority, p string) {
	c7nFindProjectErrMsgTemplate, _ := cache.HGet("wework_msg_templates", "wework_template_c7n_find_project_err")
	msg := map[string]interface{}{
		"touser":  o.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(c7nFindProjectErrMsgTemplate, o.SpName, o.DisplayName, p),
		},
	}
	_, err := model.CorpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
	}
	log.Log.Info("企业微信回执消息:工单[" + o.SpName + "]用户[" + o.Userid + "]姓名[" + o.DisplayName + "]工号[" + o.Eid + "]状态[c7n项目权限|未找到c7n项目: " + p + " ]")
}

// handleOrderC7nAuthority c7n权限处理
func handleOrderC7nAuthority(order order.C7nAuthority) (err error) {
	// c7n 用户处理流程
	c7nUser, err := c7n.FetchUser(order.DisplayName, order.Eid)
	if err != nil || c7nUser.Id == "" { // 有报错或者未查询到用户则回执执行错误消息
		handleC7nOrderFindUserErr(order, order.DisplayName, order.Eid)
		return
	}

	// c7n项目及角色处理流程
	for _, p := range order.C7nProjects {
		project, err := c7n.FetchProject(p.Project)
		if err != nil {
			handleC7nOrderFindProjectErr(order, p.Project)
			continue // 若遇到用户填写错误 继续执行下一条
		}

		// 角色的处理流程
		var c7nRoleIds []string
		for _, r := range p.Roles {
			role, _ := c7n.FetchRole(r)
			c7nRoleIds = append(c7nRoleIds, role.Id)
		}

		// 将用户添加到对应项目对应角色
		err = c7n.AssignUserProjectRole(strconv.Itoa(project.Id), c7nUser.Id, c7nRoleIds)
		s, _ := json.Marshal(p.Roles)
		if err != nil {
			log.Log.Info("为用户[" + c7nUser.RealName + "]分配项目[" + project.Name + "]的[" + string(s) + "]失败, " + err.Error())
		}
		log.Log.Info("成功为用户[" + c7nUser.RealName + "]分配项目[" + project.Name + "]的[" + string(s) + "]角色")
	}

	return
}
