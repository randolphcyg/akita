package hruser

import (
	"encoding/json"
	"gitee.com/RandolphCYG/akita/internal/middleware/log"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
)

/*
* 这里是外部接口(HR数据)的模型
 */

// HrDataService HR数据查询条件
type HrDataService struct {
	// 获取 token 的 URL
	UrlGetToken string `json:"url_get_token" gorm:"type:varchar(255);not null;comment:获取token的地址"`
	// 获取 数据 的URL
	UrlGetData string `json:"url_get_data" gorm:"type:varchar(255);not null;comment:获取数据的地址"`
}

// CacheUsersManual 手动触发缓存HR用户
func CacheUsersManual() serializer.Response {
	go func() {
		CacheUsers()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发缓存HR用户成功!"}
}

// CacheUsers 缓存HR用户
func CacheUsers() {
	log.Log.Info("开始缓存HR用户...")
	var hrDataConn hr.HrDataConn
	if result := model.DB.First(&hrDataConn); result.Error != nil {
		log.Log.Error("Fail to get HR data connection cfg!")
	}
	hrUsers, err := hr.FetchData(&hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	})
	if err != nil {
		log.Log.Error(err)
		return
	}

	// 先清空缓存
	_, err = cache.HDel("hr_users")
	if err != nil {
		log.Log.Error("Fail to clean ldap users cache,:", err)
	}

	// 将HR接口元数据写入缓存
	for _, user := range hrUsers {
		userData, _ := json.Marshal(user)
		_, err := cache.HSet("hr_users", user.Name+user.Eid, userData)
		if err != nil {
			log.Log.Error("Fail to update ldap users to cache,:", err)
		}
	}
	log.Log.Info("缓存HR用户成功!")
}
