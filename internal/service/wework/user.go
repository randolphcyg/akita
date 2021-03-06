package wework

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gitee.com/RandolphCYG/akita/internal/middleware/log"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/ldapuser"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"
)

type Msg struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// UserDetails 用户
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
	Order          []int  `json:"weOrder"`
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

// UsersMsg 获取用户列表消息
type UsersMsg struct {
	Userlist []User `json:"userlist"`
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
}

// Depart 部门
type Depart struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Order    int    `json:"weOrder"`
	Parentid int    `json:"parentid"`
}

// DepartsMsg 获取部门列表消息
type DepartsMsg struct {
	Department []Depart `json:"department"`
	Errcode    int      `json:"errcode"`
	Errmsg     string   `json:"errmsg"`
}

// CacheUsersManual 手动触发缓存企微用户
func CacheUsersManual() serializer.Response {
	go func() {
		CacheUsers()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发缓存企微用户成功!"}
}

// CacheUsers 缓存企业微信用户
func CacheUsers() {
	log.Log.Info("开始更新企微用户缓存...")
	var usersMsg UsersMsg
	res, err := model.CorpAPIUserManager.UserSimpleList(map[string]interface{}{
		"department_id": "1",
		"fetch_child":   "1",
	})
	if err != nil {
		log.Log.Error("Fail to get wework user list:", err)
	}

	temp, _ := json.Marshal(res)
	err = json.Unmarshal(temp, &usersMsg)
	if err != nil {
		return
	}

	if usersMsg.Errcode != 0 {
		err = errors.Wrap(err, serializer.ErrFetchWeUserList+usersMsg.Errmsg)
		return
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
			getUserDetailRes, err := model.CorpAPIUserManager.UserGet(map[string]interface{}{
				"userid": userInfo.Userid,
			})
			if err != nil {
				log.Log.Error(err)
			}

			temp, _ := json.Marshal(getUserDetailRes)
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
	log.Log.Info("更新企微用户缓存完成!")
}

// FetchUser 根据工号查找用户
func FetchUser(eid string) (userDetails UserDetails, err error) {
	user, err := cache.HGet("wework_users", eid)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(user), &userDetails)
	if err != nil {
		return UserDetails{}, err
	}
	return
}

// FetchDepart 通过HR部门名称及父部门名称获取唯一部门
func FetchDepart(hrDepartName string) (depart Depart, err error) {
	var departName, parentDepartName string
	if len(strings.Split(hrDepartName, ".")) >= 1 {
		departName = strings.Split(hrDepartName, ".")[len(strings.Split(hrDepartName, "."))-1]
	}
	if len(strings.Split(hrDepartName, ".")) >= 2 {
		parentDepartName = strings.Split(hrDepartName, ".")[len(strings.Split(hrDepartName, "."))-2]
	}

	res, err := model.CorpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		log.Log.Error(err)
		return
	}
	b, err := json.Marshal(res)
	var departsMsg DepartsMsg
	json.Unmarshal(b, &departsMsg)

	if departsMsg.Errcode != 0 {
		err = errors.Wrap(err, serializer.ErrFetchWeUserList+departsMsg.Errmsg)
		return
	}

	for _, d := range departsMsg.Department {
		if departName == d.Name { // 如果发现有多个同名部门，则查询其父部门进行综合判断
			parentDepart, err := FetchDepartById(d.Parentid)
			if err != nil {
				log.Log.Error(err)
			}
			if parentDepart.Name == parentDepartName {
				return d, nil
			}
		}
	}

	return
}

// FetchDepartById 根据ID获取部门
func FetchDepartById(id int) (department Depart, err error) {
	res, err := model.CorpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		log.Log.Error(err)
		return
	}
	var departsMsg DepartsMsg
	b, err := json.Marshal(res)
	json.Unmarshal(b, &departsMsg)
	for _, depart := range departsMsg.Department {
		if depart.Id == id {
			return depart, nil
		}
	}

	if departsMsg.Errcode != 0 {
		err = errors.Wrap(err, serializer.ErrFetchWeUserList+departsMsg.Errmsg)
		return
	}

	return
}

// FetchDeparts 获取部门列表
func FetchDeparts() (departsMsg DepartsMsg, err error) {
	res, err := model.CorpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		log.Log.Error(err)
		return
	}
	b, err := json.Marshal(res)

	err = json.Unmarshal(b, &departsMsg)
	if err != nil {
		return DepartsMsg{}, err
	}

	if departsMsg.Errcode != 0 {
		err = errors.Wrap(err, serializer.ErrFetchWeUserList+departsMsg.Errmsg)
		return
	}

	return
}

// CreateUser 创建企业微信用户
func CreateUser(user *ldapuser.LdapAttributes) (err error) {
	weworkUserInfos := map[string]interface{}{
		"userid":            user.Sam,
		"name":              user.DisplayName,
		"mobile":            user.Phone,
		"department":        []int{user.WeworkDepartId},
		"main_department":   user.WeworkDepartId,
		"email":             user.Email,
		"is_leader_in_dept": []int{0},
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

	var msg Msg
	res, err := model.CorpAPIUserManager.UserCreate(weworkUserInfos)
	if err != nil {
		return
	}

	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if msg.Errcode != 0 {
		err = errors.New(msg.Errmsg)
	}

	// 打标签
	if user.ProbationFlag == 1 {
		weworkUserTagInfos := map[string]interface{}{
			"tagid":    36,
			"userlist": []string{user.Sam},
		}
		var WeworkMsgTag Msg
		tagRes, _ := model.CorpAPIUserManager.TagAddUser(weworkUserTagInfos)

		bTagRes, _ := json.Marshal(tagRes)
		json.Unmarshal(bTagRes, &WeworkMsgTag)
		if WeworkMsgTag.Errcode == 0 && WeworkMsgTag.Errmsg == "ok" {
			log.Log.Info("已为[", user.DisplayName, "]打上[试用期员工]标签!")
		}
	}

	return
}

// ScanNewHrUsersManual 手动触发扫描HR数据并为新员工创建企业微信账号
func ScanNewHrUsersManual() serializer.Response {
	go func() {
		CacheUsers() // 更新企业微信缓存
		ScanNewHrUsers()
		CacheUsers() // 更新企业微信缓存
	}()
	return serializer.Response{Data: 0, Msg: "Success to scan new hr users to wework!"}
}

// ScanNewHrUsers 扫描HR数据并为新员工创建企业微信账号
func ScanNewHrUsers() {
	hrUsers, err := cache.HGetAll("hr_users") // 从缓存取HR元数据
	if err != nil {
		err = errors.Wrap(err, serializer.ErrFetchLDAPUserCache)
		return
	}

	for _, hu := range hrUsers {
		var hrUser hr.User
		json.Unmarshal([]byte(hu), &hrUser) // 反序列化
		// 判断如果企业微信没这个本公司用户，则进行创建，并记录到数据库这个操作
		if hrUser.CompanyCode == "2600" && hrUser.Stat != "离职" {
			u, _ := FetchUser(strings.TrimSpace(hrUser.Eid)) // HR数据中有少数工号错误地加了空格 这里进行去除
			if u.Name == "" {                                // 本公司新人
				dp, _ := FetchDepart(hrUser.Department)
				var userInfos *ldapuser.LdapAttributes
				if dp.Name != "" {
					// 组装LDAP用户数据
					userInfos = &ldapuser.LdapAttributes{
						Sam:            hrUser.Eid,
						Num:            hrUser.Eid,
						DisplayName:    hrUser.Name,
						Email:          strings.ToLower(hrUser.Mail),
						Phone:          hrUser.Mobile,
						WeworkExpire:   "",
						WeworkDepartId: dp.Id,
						ProbationFlag:  1,
					}
					err = CreateUser(userInfos)
					if err != nil {
						log.Log.Error("Fail to create wework automatically, ", err)
						model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "自动创建失败, "+err.Error())
						break
					}
					recordMsg := "新用户 分配至企微[" + dp.Name + "]"
					model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, recordMsg)
				} else {
					// 组装LDAP用户数据
					userInfos = &ldapuser.LdapAttributes{
						Sam:            hrUser.Eid,
						Num:            hrUser.Eid,
						DisplayName:    hrUser.Name,
						Email:          strings.ToLower(hrUser.Mail),
						Phone:          hrUser.Mobile,
						WeworkExpire:   "",
						WeworkDepartId: 69,
						ProbationFlag:  1,
					}
					log.Log.Warning("未找到与该新用户[" + userInfos.DisplayName + "]的HR数据部门对应的同名企业微信部门！将用户暂存在待分配~")
					recordMsg := "HR数据部门[" + hrUser.Department + "]分配至企业微信部门[69新加入待分配]"
					model.CreateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, recordMsg)
					err = CreateUser(userInfos)
					if err != nil {
						log.Log.Error("Fail to create wework automatically, ", err)
						model.UpdateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, "HR数据部门["+hrUser.Department+"]分配至企微部门[69新加入待分配]", "自动创建失败, "+err.Error())
						break
					}
					// 消息自定义
					if userInfos.ProbationFlag == 1 {
						model.UpdateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, recordMsg, "新用户 "+recordMsg+" Tag:[试用期员工]")
						log.Log.Info("新用户 " + recordMsg + " Tag:[试用期员工]")
					} else {
						model.UpdateWeworkUserSyncRecord(userInfos.Sam, userInfos.DisplayName, userInfos.Num, recordMsg, "新用户 "+recordMsg)
						log.Log.Info("新用户 " + recordMsg)
					}

				}
			} else { // 已经存在的公司老人
				if len(u.Extattr.Attrs) >= 2 && u.Extattr.Attrs[1].Name == "过期日期" && u.Extattr.Attrs[1].Value != "" { // 外部员工转为本公司员工，去除过期字段
					err := u.ClearUserExpiredFlag()
					if err != nil {
						log.Log.Error(err)
					}
				}
			}
		}

	}
}

// ScanExpiredUsersManual 手动触发扫描企业微信过期用户
func ScanExpiredUsersManual() serializer.Response {
	go func() {
		ScanExpiredUsers()
	}()
	return serializer.Response{Data: 0, Msg: "Success to scan expired wework users!"}
}

// ScanExpiredUsers 扫描企业微信过期用户 内部人员根据HR接口；外部人员根据过期标识，过期标识临近则发送提醒
func ScanExpiredUsers() {
	done := make(chan bool, 2)

	// 检查HR数据中过期的内部用户
	go func() {
		hrUsers, err := cache.HGetAll("hr_users") // 从缓存取HR元数据
		if err != nil {
			err = errors.Wrap(err, serializer.ErrFetchLDAPUserCache)
			return
		}

		for _, u := range hrUsers {
			var hrUser hr.User
			json.Unmarshal([]byte(u), &hrUser) // 反序列化
			if hrUser.Stat == "离职" {
				weworkUser, _ := FetchUser(strings.TrimSpace(hrUser.Eid))
				if weworkUser.Userid != "" { // 若发现HR数据中离职的用户 企业微信账号则干掉
					DeleteUser(weworkUser) // 删除操作
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
	now := time.Now()
	// 若是周一 则将周末的处理结果一并发出
	var weworkUserSyncRecords []model.WeworkUserSyncRecord
	if util.IsMonday(now) {
		weworkUserSyncRecords, _ = model.FetchWeworkUserSyncRecord(-2, 1)
	} else {
		weworkUserSyncRecords, _ = model.FetchWeworkUserSyncRecord(0, 1)
	}

	// 发消息
	today := now.Format("2006年01月02日")
	tempTitle := `<font color="warning"> ` + today + ` </font>企业微信用户变化：`
	temp := `>%s. <font color="warning"> %s </font>账号<font color="comment"> %s </font>变化类别<font color="info"> %s </font>`
	var msgs string
	for i, u := range weworkUserSyncRecords {
		if i != len(weworkUserSyncRecords) {
			msgs += "\n\n"
		}
		msgs += fmt.Sprintf(temp, strconv.Itoa(i+1), u.Name, u.UserId, u.SyncKind)
	}

	// 根据是否为节假日决定是否发消息
	if isSilent, _ := util.IsHolidaySilentMode(now); isSilent {
		// 消息静默
	} else {
		// 工作日正常发送通知
		if len(weworkUserSyncRecords) == 0 {
			util.SendRobotMsg(`<font color="warning"> ` + today + ` </font>企业微信用户无变化`)
		} else {
			// 消息过长 作剪裁处理
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
	renewalNotifyWeworkMsgTemplate, err := cache.HGet("wework_msg_templates", "wework_template_wework_renewal_notify")
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
	_, err = model.CorpAPIMsg.MessageSend(msg)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrSendWeMsg)
		return
	}
	log.Log.Info("企业微信回执消息:企业微信用户【" + user.Userid + "】姓名【" + user.Name + "】状态【即将过期】")
}

// DisableUser 禁用企业微信用户
func DisableUser(u UserDetails) (err error) {
	weworkUserInfos := map[string]interface{}{
		"userid": u.Userid,
		"enable": 0,
	}
	// 更新用户
	var msg Msg
	res, err := model.CorpAPIUserManager.UserUpdate(weworkUserInfos)
	if err != nil {
		return
	}

	b, err := json.Marshal(res)
	err = json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	// 此处将禁用企业微信用户的记录保存下
	if len(u.Extattr.Attrs) >= 1 && u.Extattr.Attrs[0].Name == "工号" {
		model.CreateWeworkUserSyncRecord(u.Userid, u.Name, u.Extattr.Attrs[0].Value, "禁用")
	}

	log.Log.Info("Success to disable wework user: " + u.Name + " " + u.Userid)
	return
}

// DeleteUser 删除企业微信用户
func DeleteUser(u UserDetails) (err error) {
	weworkUserInfos := map[string]interface{}{
		"userid": u.Userid,
	}
	// 更新用户
	var msg Msg
	res, err := model.CorpAPIUserManager.UserDelete(weworkUserInfos)
	if err != nil {
		return
	}

	b, err := json.Marshal(res)
	err = json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	// 此处将删除企业微信用户的记录保存下
	if len(u.Extattr.Attrs) >= 1 && u.Extattr.Attrs[0].Name == "工号" {
		model.CreateWeworkUserSyncRecord(u.Userid, u.Name, u.Extattr.Attrs[0].Value, "删除")
	}

	log.Log.Info("Success to delete wework user: " + u.Name + " " + u.Userid)
	return
}

// ClearUserExpiredFlag 清空用户过期字段
func (u *UserDetails) ClearUserExpiredFlag() (err error) {
	weworkUserInfos := map[string]interface{}{
		"userid": u.Userid,
		"extattr": map[string]interface{}{
			"attrs": []interface{}{
				map[string]interface{}{
					"type": 0,
					"name": "过期日期",
					"text": map[string]string{
						"value": "",
					},
				},
			},
		},
	}

	// 更新用户
	var msg Msg
	res, err := model.CorpAPIUserManager.UserUpdate(weworkUserInfos)
	if err != nil {
		return
	}

	b, err := json.Marshal(res)
	err = json.Unmarshal(b, &msg)
	if err != nil {
		log.Log.Error(err)
		return
	}
	// 此处将修改企业微信用户的记录保存下
	if len(u.Extattr.Attrs) >= 1 && u.Extattr.Attrs[0].Name == "工号" {
		model.CreateWeworkUserSyncRecord(u.Userid, u.Name, u.Extattr.Attrs[0].Value, "外部员工转为本公司员工，去除过期字段")
	}

	log.Log.Info("Success to clear wework user's expired flag!")
	return
}

// Renewal 账号续期
func (u *UserDetails) Renewal(eid string, expireDays int) (err error) {
	weworkUserInfos := map[string]interface{}{
		"userid": u.Userid,
		"enable": 1,
		"extattr": map[string]interface{}{
			"attrs": []interface{}{
				map[string]interface{}{
					"type": 0,
					"name": "工号",
					"text": map[string]string{
						"value": eid,
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
	var msg Msg
	res, err := model.CorpAPIUserManager.UserUpdate(weworkUserInfos)
	if err != nil {
		return
	}

	b, err := json.Marshal(res)
	err = json.Unmarshal(b, &msg)
	if err != nil {
		return err
	}
	return nil
}
