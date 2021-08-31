package wework

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/ldap"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
)

type WeworkMsg struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// 用户
type UserDetails struct {
	Errcode        int    `json:"errcode"`
	Errmsg         string `json:"errmsg"`
	Userid         string `json:"userid"`
	Name           string `json:"name"`
	Mobile         string `json:"mobile"`
	Position       string `json:"position"`
	Gender         string `json:"gender"`
	Email          string `json:"email"`
	Status         int    `json:"status"`
	IsLeaderInDept []int  `json:"is_leader_in_dept"`
	MainDepartment int    `json:"main_department"`
	Department     []int  `json:"department"`
	Order          []int  `json:"order"`
	Extattr        struct {
		Attrs []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text"`
			Value string `json:"value"`
		} `json:"attrs"`
	} `json:"extattr"`
}

type User struct {
	Userid     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"department"`
}

// 获取用户列表消息
type UsersMsg struct {
	Userlist []User `json:"userlist"`
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
}

// 部门
type Depart struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Order    int    `json:"order"`
	Parentid int    `json:"parentid"`
}

// 获取部门列表消息
type DepartsMsg struct {
	Department []Depart `json:"department"`
	Errcode    int      `json:"errcode"`
	Errmsg     string   `json:"errmsg"`
}

// CacheWeworkUsersManual 手动触发缓存企业微信用户
func CacheWeworkUsersManual() (err error) {
	CacheWeworkUsers()
	return
}

// CacheWeworkUsers 缓存企业微信用户
func CacheWeworkUsers() {
	log.Log.Info("开始更新企业微信用户缓存...")
	var usersMsg UsersMsg
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.UserSimpleList(map[string]interface{}{
		"department_id": "1",
		"fetch_child":   "1",
	})
	if err != nil {
		log.Log.Error(err)
	}

	temp, err := json.Marshal(res)
	json.Unmarshal(temp, &usersMsg)

	if usersMsg.Errcode != 0 {
		log.Log.Error("Fail to fetch wework user list, err:", usersMsg.Errmsg)
	}

	// 先清空缓存
	_, err = cache.HDel("wework_users")
	if err != nil {
		log.Log.Error("Fail to clean wework users cache,:", err)
	}

	done := make(chan int, 20) // 带 20 个缓存
	for i, userInfo := range usersMsg.Userlist {
		go func(i int, userInfo User) {
			var userDetails UserDetails
			getUserDetailRes, err := corpAPIUserManager.UserGet(map[string]interface{}{
				"userid": userInfo.Userid,
			})
			if err != nil {
				log.Log.Error(err)
			}

			temp, err := json.Marshal(getUserDetailRes)
			json.Unmarshal(temp, &userDetails)

			if len(userDetails.Extattr.Attrs) >= 1 && userDetails.Extattr.Attrs[0].Name == "工号" { // 忽略不符合规范的用户
				_, err = cache.HSet("wework_users", userDetails.Extattr.Attrs[0].Value, temp) // 缓存用户
				if err != nil {
					log.Log.Error("Fail to cache wework user,:", err)
				}
			}
			<-done
		}(i, userInfo)
		done <- 1
	}
	log.Log.Info("更新企业微信用户缓存完成!")
}

// FetchUser 根据工号查找用户
func FetchUser(eid string) (userDetails UserDetails, err error) {
	user, err := cache.HGet("wework_users", eid)
	if err != nil {
		err = errors.New("无此用户: " + err.Error())
		return
	}
	json.Unmarshal([]byte(user), &userDetails)
	return
}

// FetchDeparts 通过部门名称获取唯一部门ID
func FetchDepart(departName string) (depart Depart, err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		log.Log.Error(err)
		return
	}
	b, err := json.Marshal(res)
	var departsMsg DepartsMsg
	json.Unmarshal(b, &departsMsg)

	if departsMsg.Errcode != 0 {
		log.Log.Error("Fail to fetch wework user list, err:", departsMsg.Errmsg)
		return
	}
	for _, d := range departsMsg.Department {
		if departName == d.Name { // 只返回查到的第一个
			return d, nil
		}
	}

	return
}

// FetchDeparts 获取部门列表
func FetchDeparts() (departsMsg DepartsMsg, err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		log.Log.Error(err)
		return
	}
	b, err := json.Marshal(res)
	json.Unmarshal(b, &departsMsg)

	if departsMsg.Errcode != 0 {
		log.Log.Error("Fail to fetch wework user list, err:", departsMsg.Errmsg)
		return
	}

	return
}

// CreateUser 创建企业微信用户
func CreateUser(user *ldap.LdapAttributes) (err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	weworkUserInfos := map[string]interface{}{
		"userid":            user.Sam,
		"name":              user.DisplayName,
		"mobile":            user.Phone,
		"department":        user.WeworkDepartId,
		"email":             user.Email,
		"is_leader_in_dept": 0,
		"enable":            1,
		// 自定义属性
		"extattr": map[string]interface{}{
			"attrs": []interface{}{
				map[string]interface{}{
					"type": 0,
					"name": "工号",
					"text": map[string]string{
						"value": user.Num,
					},
				},
				map[string]interface{}{
					"type": 0,
					"name": "过期日期",
					"text": map[string]string{
						"value": user.WeworkExpire,
					},
				},
			},
		},
		"to_invite": true, // 邀请用户
	}

	var msg WeworkMsg
	res, err := corpAPIUserManager.UserCreate(weworkUserInfos)
	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if msg.Errcode != 0 {
		err = errors.New(msg.Errmsg)
	}
	return
}

// RenewalUser 企业微信用户续期
func RenewalUser(weworkUserId string, applicant order.RenewalApplicant, expireDays int) (err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	weworkUserInfos := map[string]interface{}{
		"userid": weworkUserId,
		"enable": 1,
		"extattr": map[string]interface{}{
			"attrs": []interface{}{
				map[string]interface{}{
					"type": 0,
					"name": "工号",
					"text": map[string]string{
						"value": applicant.Eid,
					},
				},
				map[string]interface{}{
					"type": 0,
					"name": "过期日期",
					"text": map[string]string{
						"value": util.ExpireStr(expireDays),
					},
				},
			},
		},
	}
	// 更新用户
	var msg WeworkMsg
	res, err := corpAPIUserManager.UserUpdate(weworkUserInfos)
	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	log.Log.Info("Success to renewal wework user!")
	return nil
}

// ScanNewHrUsersManual 手动触发扫描HR数据并为新员工创建企业微信账号
func ScanNewHrUsersManual() serializer.Response {
	go func() {
		ScanNewHrUsers()
	}()
	return serializer.Response{Data: 0, Msg: "Success to scan new hr users to wework!"}
}

// ScanNewHrUsers 扫描HR数据并为新员工创建企业微信账号
func ScanNewHrUsers() {
	go func() {
		hrUsers, err := cache.HGetAll("hr_users") // 从缓存取HR元数据
		if err != nil {
			log.Log.Error("Fail to fetch ldap users cache,:", err)
		}
		for _, u := range hrUsers {
			var hrUser hr.HrUser
			json.Unmarshal([]byte(u), &hrUser) // 反序列化
			// 判断如果企业微信没这个本公司用户，则进行创建，并记录到数据库这个操作
			if hrUser.Stat != "离职" && hrUser.CompanyCode == "2600" {
				u, _ := FetchUser(hrUser.Eid)
				if u.Name == "" {
					dp, _ := FetchDepart(strings.Split(hrUser.Department, ".")[len(strings.Split(hrUser.Department, "."))-1])
					var userInfos *ldap.LdapAttributes
					if dp.Name != "" {
						// 组装LDAP用户数据
						userInfos = &ldap.LdapAttributes{
							Sam:            hrUser.Eid,
							Num:            hrUser.Eid,
							DisplayName:    hrUser.Name,
							Email:          strings.ToLower(hrUser.Mail),
							Phone:          hrUser.Mobile,
							WeworkExpire:   "",
							WeworkDepartId: dp.Id,
						}
						err = CreateUser(userInfos)
						if err != nil {
							log.Log.Error("Fail to create wework automatically, ", err)
							model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "自动创建失败, "+err.Error())
						}
						model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "新用户 分配至企微["+dp.Name+"]")
					} else {
						// 组装LDAP用户数据
						userInfos = &ldap.LdapAttributes{
							Sam:            hrUser.Eid,
							Num:            hrUser.Eid,
							DisplayName:    hrUser.Name,
							Email:          strings.ToLower(hrUser.Mail),
							Phone:          hrUser.Mobile,
							WeworkExpire:   "",
							WeworkDepartId: 69,
						}
						log.Log.Warning("未找到与该新用户[" + userInfos.DisplayName + "]的HR数据部门对应的同名企业微信部门！将用户暂存在待分配~")
						model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "HR数据部门["+hrUser.Department+"]自动分配至企业微信部门[69新加入待分配]")
						err = CreateUser(userInfos)
						if err != nil {
							log.Log.Error("Fail to create wework automatically, ", err)
							model.UpdateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "HR数据部门["+hrUser.Department+"]自动分配至企微部门[69新加入待分配]", "自动创建失败, "+err.Error())
						}
						model.UpdateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "HR数据部门["+hrUser.Department+"]自动分配至企微部门[69新加入待分配]", "新用户 HR数据部门["+hrUser.Department+"]自动分配至企微部门[69新加入待分配]")
					}
				}
			}
		}
	}()
}

// ScanExpiredWeworkUsersManual 手动触发扫描企业微信过期用户
func ScanExpiredWeworkUsersManual() serializer.Response {
	go func() {
		ScanExpiredWeworkUsers()
	}()
	return serializer.Response{Data: 0, Msg: "Success to scan expired wework users!"}
}

// ScanExpiredWeworkUsers 扫描企业微信过期用户 内部人员根据HR接口；外部人员根据过期标识，过期标识临近则发送提醒
func ScanExpiredWeworkUsers() {
	done := make(chan bool, 2)

	// 检查HR数据中过期的内部用户
	go func() {
		hrUsers, err := cache.HGetAll("hr_users") // 从缓存取HR元数据
		if err != nil {
			log.Log.Error("Fail to fetch ldap users cache,:", err)
		}
		for _, u := range hrUsers {
			var hrUser hr.HrUser
			json.Unmarshal([]byte(u), &hrUser) // 反序列化
			if hrUser.Stat == "离职" {
				weworkUser, _ := FetchUser(strings.TrimSpace(hrUser.Eid))
				if weworkUser.Userid != "" && weworkUser.Status == 1 { // 若发现HR数据中离职的用户 企业微信状态还是1则禁用
					model.CreateWeworkUserSyncRecord(weworkUser.Userid, weworkUser.Name, weworkUser.Extattr.Attrs[0].Value, "禁用")
					DisableUser(weworkUser) // 禁用
				}
			}
		}
		done <- true
	}()

	// 检查企业微信中过期的外部用户
	go func() {
		weworkUsers, _ := cache.HGetAll("wework_users")
		for _, u := range weworkUsers {
			var weworkUser UserDetails
			json.Unmarshal([]byte(u), &weworkUser)
			if len(weworkUser.Extattr.Attrs) >= 2 && weworkUser.Extattr.Attrs[1].Name == "过期日期" && weworkUser.Extattr.Attrs[1].Value != "" {
				if util.IsExpire(weworkUser.Extattr.Attrs[1].Value) { // 若已经过期
					model.CreateWeworkUserSyncRecord(weworkUser.Userid, weworkUser.Name, weworkUser.Extattr.Attrs[0].Value, "禁用")
					DisableUser(weworkUser) // 禁用
				} else { // 若即将过期，符合条件则发送即将过期通知
					remainingDays := util.SubDays(util.ExpireStrToTime(weworkUser.Extattr.Attrs[1].Value), time.Now())
					if remainingDays == 1 || remainingDays == 2 || remainingDays == 3 || remainingDays == 7 || remainingDays == 14 { // 倒数三天以及倒数1/2周都发通知
						SendWeworkOuterUserExpiredMsg(weworkUser, remainingDays)
					}
				}
			}
		}
		done <- true
	}()

	_, _ = <-done, <-done
	log.Log.Info("扫描内外部公司过期企业微信用户完成!")

	// 汇总通知
	todayWeworkUserSyncRecords, _ := model.FetchTodayWeworkUserSyncRecord()
	// 发消息
	now := time.Now()
	today := now.Format("2006年01月02日")
	tempTitle := `<font color="warning"> ` + today + ` </font>企业微信用户变化：`
	temp := `>%s. <font color="warning"> %s </font>账号<font color="comment"> %s </font>变化类别<font color="info"> %s </font>`
	var msgs string
	for i, u := range todayWeworkUserSyncRecords {
		if i != len(todayWeworkUserSyncRecords) {
			msgs += "\n\n"
		}
		msgs += fmt.Sprintf(temp, strconv.Itoa(i+1), u.Name, u.UserId, u.SyncKind)
	}

	// 根据是否为节假日决定是否发消息
	if isSilent, _ := util.IsHolidaySilentMode(now); isSilent {
		// 消息静默
	} else {
		// 工作日正常发送通知
		if len(todayWeworkUserSyncRecords) == 0 {
			util.SendRobotMsg(`<font color="warning"> ` + today + ` </font>企业微信用户无变化`)
		} else {
			// 消息过长的作剪裁处理
			msgs := util.TruncateMsg(tempTitle+msgs, "\n\n")
			for _, m := range msgs {
				util.SendRobotMsg(m)
			}
		}
	}

	log.Log.Info("汇总通知发送成功!")
}

// SendWeworkOuterUserExpiredMsg 给企业微信即将过期用户发送续期通知
func SendWeworkOuterUserExpiredMsg(user UserDetails, remainingDays int) {
	corpAPIMsg := api.NewCorpAPI(model.WeworkUuapCfg.CorpId, model.WeworkUuapCfg.AppSecret)
	renewalNotifyWeworkMsgTemplate, err := cache.HGet("wework_templates", "wework_template_wework_renewal_notify")
	if err != nil {
		log.Log.Error("读取企业微信消息模板错误: ", err)
	}

	msg := map[string]interface{}{
		"touser":  user.Userid,
		"msgtype": "markdown",
		"agentid": model.WeworkUuapCfg.AppId,
		"markdown": map[string]interface{}{
			"content": fmt.Sprintf(renewalNotifyWeworkMsgTemplate, user.Name, strconv.Itoa(remainingDays)),
		},
	}
	_, err = corpAPIMsg.MessageSend(msg)
	if err != nil {
		log.Log.Error("Fail to send wework msg, err: ", err)
		// TODO 发送企业微信消息错误，应当考虑重发逻辑
	}
	log.Log.Info("企业微信回执消息:企业微信用户【" + user.Userid + "】姓名【" + user.Name + "】状态【即将过期】")
}

// DisableUser 禁用企业微信用户
func DisableUser(u UserDetails) (err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	weworkUserInfos := map[string]interface{}{
		"userid": u.Userid,
		"enable": 0,
	}
	// 更新用户
	var msg WeworkMsg
	res, err := corpAPIUserManager.UserUpdate(weworkUserInfos)
	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	// 此处将禁用企业微信用户的记录保存下
	if len(u.Extattr.Attrs) >= 1 && u.Extattr.Attrs[0].Name == "工号" {
		model.CreateWeworkUserSyncRecord(u.Userid, u.Name, u.Extattr.Attrs[0].Value, "禁用")
	}

	log.Log.Info("Success to disable wework user!")
	return
}

// FormatHistoryUser 企业微信历史用户规整
func FormatHistoryUser(user UserDetails, formatEid string, formatMail string) (err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	weworkUserInfos := map[string]interface{}{
		"userid": user.Userid,
		"enable": 1,
		"email":  formatMail,
		"extattr": map[string]interface{}{
			"attrs": []interface{}{
				map[string]interface{}{
					"type": 0,
					"name": "工号",
					"text": map[string]string{
						"value": formatEid,
					},
				},
			},
		},
	}
	// 更新用户
	var msg WeworkMsg
	res, err := corpAPIUserManager.UserUpdate(weworkUserInfos)
	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	log.Log.Info("Success to format wework user!")
	return nil
}
