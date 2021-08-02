package wework

import (
	"encoding/json"
	"fmt"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"github.com/sirupsen/logrus"
)

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
	Extattr        []struct {
		Type int `json:"type"`
		Name int `json:"name"`
		Text []struct {
			Value string `json:"value"`
		} `json:"text"`
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

// FetchUsers 获取用户列表
func FetchUsers() (userDetails UserDetails, err error) {
	var usersMsg UsersMsg
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	res, err := corpAPIUserManager.UserSimpleList(map[string]interface{}{
		"department_id": "1",
		"fetch_child":   "1",
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	temp, err := json.Marshal(res)
	json.Unmarshal(temp, &usersMsg)

	if usersMsg.Errcode != 0 {
		logrus.Error("Fail to fetch wework user list, err:", usersMsg.Errmsg)
		return
	}

	// 先清空缓存
	_, err = cache.HDel("wework_users")
	if err != nil {
		logrus.Error("Fail to clean wework users cache,:", err)
	}

	for i, userInfo := range usersMsg.Userlist {
		getUserDetailRes, err := corpAPIUserManager.UserGet(map[string]interface{}{
			"userid": userInfo.Userid,
		})
		if err != nil {
			logrus.Error(err)
		}

		temp, err := json.Marshal(getUserDetailRes)
		json.Unmarshal(temp, &userDetails)
		_, err = cache.HSet("wework_users", userDetails.Userid, temp)
		if err != nil {
			logrus.Error("Fail to cache wework user,:", err)
		}
		fmt.Println(i+1, userDetails)
		// break
	}
	return
}

func FetchUserByEid() {

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

func CreateWeworkUser(user *ldap.LdapAttributes) (err error) {
	corpAPIUserManager := api.NewCorpAPI(model.WeworkUserManageCfg.CorpId, model.WeworkUserManageCfg.AppSecret)
	weworkUserInfos := map[string]interface{}{
		"userid":            user.Num,
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
			},
		},
		"to_invite": false,
	}
	// 创建用户
	_, err = corpAPIUserManager.UserCreate(weworkUserInfos)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Info("Success to create wework user!")
	return
}
