package wework

import (
	"encoding/json"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/util"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
	"github.com/sirupsen/logrus"
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
	logrus.Info("开始更新企业微信用户缓存...")
	var usersMsg UsersMsg
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.UserSimpleList(map[string]interface{}{
		"department_id": "1",
		"fetch_child":   "1",
	})
	if err != nil {
		logrus.Error(err)
	}

	temp, err := json.Marshal(res)
	json.Unmarshal(temp, &usersMsg)

	if usersMsg.Errcode != 0 {
		logrus.Error("Fail to fetch wework user list, err:", usersMsg.Errmsg)
	}

	// 先清空缓存
	_, err = cache.HDel("wework_users")
	if err != nil {
		logrus.Error("Fail to clean wework users cache,:", err)
	}

	done := make(chan int, 20) // 带 20 个缓存
	for i, userInfo := range usersMsg.Userlist {
		go func(i int, userInfo User) {
			var userDetails UserDetails
			getUserDetailRes, err := corpAPIUserManager.UserGet(map[string]interface{}{
				"userid": userInfo.Userid,
			})
			if err != nil {
				logrus.Error(err)
			}

			temp, err := json.Marshal(getUserDetailRes)
			json.Unmarshal(temp, &userDetails)

			if len(userDetails.Extattr.Attrs) >= 1 { // 不符合规范的用户忽略
				// fmt.Println(i+1, userDetails.Extattr.Attrs[0].Value, userDetails)
				// 缓存用户
				_, err = cache.HSet("wework_users", userDetails.Extattr.Attrs[0].Value, temp)
				if err != nil {
					logrus.Error("Fail to cache wework user,:", err)
				}
			}
			<-done
		}(i, userInfo)
		done <- 1
	}
	logrus.Info("更新企业微信用户缓存完成!")
}

// FetchUser 根据工号查找用户
func FetchUser(eid string) (userDetails UserDetails, err error) {
	user, err := cache.HGet("wework_users", eid)
	if err != nil {
		logrus.Error("无此用户: ", err)
		return
	}
	json.Unmarshal([]byte(user), &userDetails)
	return
}

// FetchDeparts 获取部门列表
func FetchDeparts() (departsMsg DepartsMsg, err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.DepartmentList(map[string]interface{}{})
	if err != nil {
		logrus.Error(err)
		return
	}
	b, err := json.Marshal(res)
	json.Unmarshal(b, &departsMsg)

	if departsMsg.Errcode != 0 {
		logrus.Error("Fail to fetch wework user list, err:", departsMsg.Errmsg)
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
		"department":        69,
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
		"to_invite": false,
	}
	// 创建用户
	var msg WeworkMsg
	res, err := corpAPIUserManager.UserCreate(weworkUserInfos)
	b, err := json.Marshal(res)
	json.Unmarshal(b, &msg)
	if msg.Errcode != 0 {
		logrus.Error("企业微信用户创建出错: ", user.DisplayName, user.Sam, msg.Errmsg)
	}
	logrus.Info("Success to create wework user!")
	return nil
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
		logrus.Error(err)
		return
	}
	logrus.Info("Success to renewal wework user!")
	return nil
}

// ScanExpireUsers TODO 全量扫描过期账户
func ScanExpireUsers() serializer.Response {
	var usersMsg UsersMsg
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.UserSimpleList(map[string]interface{}{
		"department_id": "1",
		"fetch_child":   "1",
	})
	if err != nil {
		logrus.Error(err)
		return serializer.Response{Data: -1, Msg: "Fail to update wework user cache!"}
	}

	temp, err := json.Marshal(res)
	json.Unmarshal(temp, &usersMsg)

	if usersMsg.Errcode != 0 {
		logrus.Error("Fail to fetch wework user list, err:", usersMsg.Errmsg)
		return serializer.Response{Data: -1, Msg: "Fail to fetch wework users!"}
	}

	var users []UserDetails
	done := make(chan int, 20) // 带 20 个缓存

	// 遍历企业微信用户
	for i, userInfo := range usersMsg.Userlist {
		go func(i int, userInfo User) {
			var userDetails UserDetails
			// 查询
			getUserDetailRes, err := corpAPIUserManager.UserGet(map[string]interface{}{
				"userid": userInfo.Userid,
			})
			if err != nil {
				logrus.Error(err)
			}
			// fmt.Println(i+1, getUserDetailRes)
			temp, err := json.Marshal(getUserDetailRes)
			json.Unmarshal(temp, &userDetails)

			// 过期用户禁用
			if userDetails.Status == 1 {
				// DisableUser(userDetails)
			}
			<-done
		}(i, userInfo)
		done <- 1
	}

	return serializer.Response{Data: users, Msg: "Success to scan expire wework users!"}
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
		logrus.Error(err)
		return
	}
	logrus.Info("Success to disable wework user!")
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
		logrus.Error(err)
		return
	}
	logrus.Info("Success to format wework user!")
	return nil
}
