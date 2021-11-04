package c7n

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"strings"

	"github.com/kirinlabs/HttpRequest"

	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
)

// TokenResp 获取token接口返回数据结构体
type TokenResp struct {
	// 正确时候
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	// 错误时候
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ProjectFields 项目字段
type ProjectFields struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Id             int    `json:"id"`
	Enabled        bool   `json:"enabled"`
	CreationDate   string `json:"creationDate"`
	CreatedBy      int    `json:"createdBy"`
	LastUpdateDate string `json:"lastUpdateDate"`
	LastUpdatedBy  int    `json:"lastUpdatedBy"`
}

// UserResp 查询用户接口返回结构体
type UserResp struct {
	TotalPages       int          `json:"totalPages"`
	TotalElements    int          `json:"totalElements"`
	NumberOfElements int          `json:"numberOfElements"`
	Size             int          `json:"size"`
	Number           int          `json:"number"`
	Content          []UserFields `json:"content"`
	Empty            bool         `json:"empty"`
}

// UserFields 用户字段
type UserFields struct {
	Id             string `json:"id"`
	LoginName      string `json:"loginName"`
	RealName       string `json:"realName"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	OrganizationId int    `json:"organizationId"`
	Enabled        bool   `json:"enabled"`
	EnabledFlag    bool   `json:"enabledFlag"`
	Locked         bool   `json:"locked"`
	LockedFlag     bool   `json:"lockedFlag"`
	Ldap           bool   `json:"ldap"`
	LdapFlag       bool   `json:"ldapFlag"`
	Admin          bool   `json:"admin"`
	AdminFlag      bool   `json:"adminFlag"`
}

// RoleResp 查询用户接口返回结构体
type RoleResp struct {
	TotalPages       int          `json:"totalPages"`
	TotalElements    int          `json:"totalElements"`
	NumberOfElements int          `json:"numberOfElements"`
	Size             int          `json:"size"`
	Number           int          `json:"number"`
	Content          []RoleFields `json:"content"`
	Empty            bool         `json:"empty"`
}

// RoleFields 角色字段
type RoleFields struct {
	Id          string `json:"id"`
	Code        string `json:"code"`
	Assignable  bool   `json:"assignable"`
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	TplRoleName string `json:"tplRoleName"`
	RoleLevel   string `json:"roleLevel"`
}

// FetchLdapRes 查询ldap连接接口返回结构体
type FetchLdapRes struct {
	Id     string `json:"id"`
	BaseDn string `json:"baseDn"`
}

// GetToken 获取token
func GetToken() (header map[string]string, err error) {
	// 从缓存取url
	c7nFetchToken, err := cache.HGet("third_party_cfgs", "c7n_fetch_token")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
		return
	}

	var tokenResp TokenResp
	req := HttpRequest.NewRequest()
	respC7nFetchToken, err := req.Post(c7nFetchToken)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}
	defer respC7nFetchToken.Close() // 关闭

	// 反序列化
	err = respC7nFetchToken.Json(&tokenResp)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrConvertRespToJson)
		return
	}
	if tokenResp.Error != "" {
		err = errors.New(tokenResp.ErrorDescription)
		return
	}
	header = map[string]string{
		"Authorization": tokenResp.TokenType + " " + tokenResp.AccessToken,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	return header, nil
}

// FetchProject 查询 c7n 项目 返回唯一值，忽略其他结果
func FetchProject(projectName string) (projectFields ProjectFields, err error) {
	project, err := cache.HGet("c7n_projects", strings.ToUpper(projectName))
	err = json.Unmarshal([]byte(project), &projectFields)
	if err != nil {
		return
	}
	return
}

// FetchUser 根据真实姓名、登录名查询用户
func FetchUser(realName, loginName string) (user UserFields, err error) {
	// 取token
	header, err := GetToken()
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}

	// 从缓存取url
	c7nFetchUser, err := cache.HGet("third_party_cfgs", "c7n_fetch_user")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	fetchUserUrl := fmt.Sprintf(c7nFetchUser, url.QueryEscape(realName), url.QueryEscape(loginName))
	respFetchUser, err := req.Get(fetchUserUrl)
	if err != nil {
		err = errors.New("Fail to fetch c7n user, err: " + err.Error())
		return
	}
	defer respFetchUser.Close() // 关闭

	// 反序列化
	var userResp UserResp
	err = respFetchUser.Json(&userResp)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrConvertRespToJson)
		return
	}

	// 数据筛选
	if len(userResp.Content) >= 1 {
		user = userResp.Content[0]
	}

	return
}

// FetchRole 查询角色
func FetchRole(roleName string) (role RoleFields, err error) {
	// 取token
	header, err := GetToken()
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}

	// 从缓存取url
	c7nFetchRoles, err := cache.HGet("third_party_cfgs", "c7n_fetch_roles")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nFetchRoles, err := req.Get(c7nFetchRoles)
	if err != nil {
		err = errors.New("Fail to fetch c7n user, err: " + err.Error())
		return
	}
	defer respC7nFetchRoles.Close() // 关闭

	// 反序列化
	var roleResp RoleResp
	err = respC7nFetchRoles.Json(&roleResp)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrConvertRespToJson)
		return
	}

	// 数据筛选
	for _, r := range roleResp.Content {
		if r.Enabled {
			if r.Name == roleName {
				role = r
			}
		}
	}
	return
}

// AssignUserProjectRole 为c7n用户分配某项目的某角色
func AssignUserProjectRole(c7nProjectId string, c7nUserId string, c7nRoleIds []string) (err error) {
	// 取token
	header, err := GetToken()
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}

	// 从缓存取url
	c7nAssignUserProjectRole, err := cache.HGet("third_party_cfgs", "c7n_assign_user_project_role")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	_, err = req.Put(fmt.Sprintf(c7nAssignUserProjectRole, c7nProjectId, c7nUserId), c7nRoleIds)
	return
}

// SyncUsersManual 手动触发LDAP用户同步到C7N
func SyncUsersManual() serializer.Response {
	go func() {
		SyncUsers()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发LDAP用户同步到C7N成功!"}
}

// SyncUsers 同步用户
func SyncUsers() {
	// 取token
	header, err := GetToken()
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
	}

	// 从缓存取url
	getLdapConn, err := cache.HGet("third_party_cfgs", "c7n_fetch_ldap_conn")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respGetLdapConn, err := req.Get(getLdapConn)
	if err != nil {
		err = errors.New("Fail to fetch c7n user, err: " + err.Error())
	}
	defer respGetLdapConn.Close() // 关闭

	// 反序列化
	var fetchLdapRes FetchLdapRes
	err = respGetLdapConn.Json(&fetchLdapRes)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrConvertRespToJson)
	}

	// 数据筛选
	LdapId := fetchLdapRes.Id

	// 同步用户
	// 从缓存取url
	syncLdapUsers, err := cache.HGet("third_party_cfgs", "c7n_sync_ldap_users")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
	}
	respSyncLdapUsers, err := req.Post(fmt.Sprintf(syncLdapUsers, LdapId))
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
	}
	defer respSyncLdapUsers.Close() // 关闭
	return
}

// CacheProjectsManual 手动触发缓存所有c7n项目
func CacheProjectsManual() serializer.Response {
	go func() {
		CacheProjects()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发缓存c7n项目成功!"}
}

// CacheProjects 缓存所有c7n项目
func CacheProjects() {
	fmt.Println("开始更新c7n项目缓存...")
	// 从缓存取url
	c7nGetAllProjects, err := cache.HGet("third_party_cfgs", "c7n_fetch_all_projects")
	if err != nil {
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
		return
	}

	// 取token
	header, err := GetToken()
	if err != nil {
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nGetAllProjects, err := req.Get(c7nGetAllProjects)
	if err != nil {
		return
	}

	defer respC7nGetAllProjects.Close() // 关闭

	var c7nProjects []ProjectFields
	err = respC7nGetAllProjects.Json(&c7nProjects)
	if err != nil {
		return
	}

	// 清空缓存
	_, err = cache.HDel("c7n_projects")

	for _, project := range c7nProjects {
		if project.Enabled {
			data, _ := json.Marshal(project)
			_, err = cache.HSet("c7n_projects", project.Name, data)
		}
	}
	fmt.Println("更新c7n项目缓存完成")
}
