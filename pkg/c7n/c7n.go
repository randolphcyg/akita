package c7n

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/kirinlabs/HttpRequest"

	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
)

// C7nFetchTokenRes 获取 c7n token 接口返回数据结构体
type C7nFetchTokenRes struct {
	// 正确时候
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	// 错误时候
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// C7nProjectFields 项目字段
type C7nProjectFields struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Id             int    `json:"id"`
	Enabled        bool   `json:"enabled"`
	CreationDate   string `json:"creationDate"`
	CreatedBy      int    `json:"createdBy"`
	LastUpdateDate string `json:"lastUpdateDate"`
	LastUpdatedBy  int    `json:"lastUpdatedBy"`
}

// C7nFetchUserRes 查询用户接口返回结构体
type C7nFetchUserRes struct {
	TotalPages       int             `json:"totalPages"`
	TotalElements    int             `json:"totalElements"`
	NumberOfElements int             `json:"numberOfElements"`
	Size             int             `json:"size"`
	Number           int             `json:"number"`
	Content          []C7nUserFields `json:"content"`
	Empty            bool            `json:"empty"`
}

// C7nUserFields 用户字段
type C7nUserFields struct {
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

// C7nFetchRoleRes 查询用户接口返回结构体
type C7nFetchRoleRes struct {
	TotalPages       int             `json:"totalPages"`
	TotalElements    int             `json:"totalElements"`
	NumberOfElements int             `json:"numberOfElements"`
	Size             int             `json:"size"`
	Number           int             `json:"number"`
	Content          []C7nRoleFields `json:"content"`
	Empty            bool            `json:"empty"`
}

// C7nRoleFields 角色字段
type C7nRoleFields struct {
	Id          string `json:"id"`
	Code        string `json:"code"`
	Assignable  bool   `json:"assignable"`
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	TplRoleName string `json:"tplRoleName"`
	RoleLevel   string `json:"roleLevel"`
}

// C7nFetchLdapRes 查询ldap连接接口返回结构体
type C7nFetchLdapRes struct {
	Id     string `json:"id"`
	BaseDn string `json:"baseDn"`
}

// FetchToken 获取token
func FetchToken() (header map[string]string, err error) {
	// 从缓存取url
	c7nFetchToken, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_token")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	var c7nFetchTokenRes C7nFetchTokenRes
	req := HttpRequest.NewRequest()
	respC7nFetchToken, err := req.Post(c7nFetchToken)
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}
	defer respC7nFetchToken.Close() // 关闭

	// 反序列化
	err = respC7nFetchToken.Json(&c7nFetchTokenRes)
	if err != nil {
		log.Log.Error("Fail to convert response to json, err: ", err)
		return
	}
	if c7nFetchTokenRes.Error != "" {
		log.Log.Error(c7nFetchTokenRes.ErrorDescription)
		return
	}
	header = map[string]string{
		"Authorization": c7nFetchTokenRes.TokenType + " " + c7nFetchTokenRes.AccessToken,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	return header, nil
}

// FetchC7nProject 查询 c7n 项目 返回唯一值，忽略其他结果
func FetchC7nProject(projectName string) (c7nProject C7nProjectFields, err error) {
	project, err := cache.HGet("c7n_projects", strings.ToUpper(projectName))
	err = json.Unmarshal([]byte(project), &c7nProject)
	if err != nil {
		return
	}
	return
}

// FtechC7nUser 根据登录名查询用户
func FtechC7nUser(realName, loginName string) (c7nUser C7nUserFields, err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}

	// 从缓存取url
	c7nFetchUser, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_user")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nFetchUser, err := req.Get(fmt.Sprintf(c7nFetchUser, url.QueryEscape(realName), url.QueryEscape(loginName)))
	if err != nil {
		log.Log.Error("Fail to fetch c7n user, err: ", err)
		return
	}
	defer respC7nFetchUser.Close() // 关闭

	// 反序列化
	var c7nFetchUserRes C7nFetchUserRes
	err = respC7nFetchUser.Json(&c7nFetchUserRes)
	if err != nil {
		log.Log.Error("Fail to convert response to json, err: ", err)
		return
	}

	// 数据筛选
	if len(c7nFetchUserRes.Content) >= 1 {
		c7nUser = c7nFetchUserRes.Content[0]
	}

	return
}

// FetchC7nRoles 查询 c7n 角色列表
func FetchC7nRoles(roleName string) (c7nRole C7nRoleFields, err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}

	// 从缓存取url
	c7nFetchRoles, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_roles")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nFetchRoles, err := req.Get(c7nFetchRoles)
	if err != nil {
		log.Log.Error("Fail to fetch c7n user, err: ", err)
		return
	}
	defer respC7nFetchRoles.Close() // 关闭

	// 反序列化
	var c7nFetchRoleRes C7nFetchRoleRes
	err = respC7nFetchRoles.Json(&c7nFetchRoleRes)
	if err != nil {
		log.Log.Error("Fail to convert response to json, err: ", err)
		return
	}

	// 数据筛选
	for _, role := range c7nFetchRoleRes.Content {
		if role.Enabled {
			if strings.Contains(role.Name, roleName) {
				c7nRole = role
			}
		}
	}
	return
}

// AssignC7nUserProjectRole 为c7n用户分配某项目的某角色
func AssignC7nUserProjectRole(c7nProjectId string, c7nUserId string, c7nRoleIds []string) (err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}

	// 从缓存取url
	c7nAssignUserProjectRole, err := cache.HGet("third_party_sys_cfg", "c7n_assign_user_project_role")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	_, err = req.Put(fmt.Sprintf(c7nAssignUserProjectRole, c7nProjectId, c7nUserId), c7nRoleIds)
	return
}

// UpdateC7nUsersManual 手动触发LDAP用户同步到C7N
func UpdateC7nUsersManual() serializer.Response {
	go func() {
		UpdateC7nUsers()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发LDAP用户同步到C7N成功!"}
}

// UpdateC7nUsers 同步c7n ldap用户
func UpdateC7nUsers() {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
	}

	// 从缓存取url
	c7nFetchLdapConn, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_ldap_conn")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nFetchLdapConn, err := req.Get(c7nFetchLdapConn)
	if err != nil {
		log.Log.Error("Fail to fetch c7n user, err: ", err)
	}
	defer respC7nFetchLdapConn.Close() // 关闭

	// 反序列化
	var c7nFetchLdapRes C7nFetchLdapRes
	err = respC7nFetchLdapConn.Json(&c7nFetchLdapRes)
	if err != nil {
		log.Log.Error("Fail to convert response to json, err: ", err)
	}

	// 数据筛选
	LdapId := c7nFetchLdapRes.Id

	// 同步用户
	// 从缓存取url
	c7nSyncLdapUsers, err := cache.HGet("third_party_sys_cfg", "c7n_sync_ldap_users")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
	}
	respC7nSyncLdapUsers, err := req.Post(fmt.Sprintf(c7nSyncLdapUsers, LdapId))
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
	}
	defer respC7nSyncLdapUsers.Close() // 关闭
	return
}

// CacheC7nProjectsManual 手动触发缓存所有c7n项目
func CacheC7nProjectsManual() serializer.Response {
	go func() {
		CacheC7nProjects()
	}()
	return serializer.Response{Data: 0, Msg: "手动触发缓存c7n项目成功!"}
}

// CacheC7nProjects 缓存所有c7n项目
func CacheC7nProjects() {
	log.Log.Info("开始更新c7n项目缓存...")
	// 从缓存取url
	c7nFetchAllProjects, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_all_projects")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respC7nFetchAllProjects, err := req.Get(c7nFetchAllProjects)
	if err != nil {
		return
	}

	defer respC7nFetchAllProjects.Close() // 关闭

	var c7nPros []C7nProjectFields
	err = respC7nFetchAllProjects.Json(&c7nPros)
	if err != nil {
		return
	}

	// 清空缓存
	_, err = cache.HDel("c7n_projects")

	for _, project := range c7nPros {
		if project.Enabled {
			data, _ := json.Marshal(project)
			_, err = cache.HSet("c7n_projects", project.Name, data)
		}
	}
	log.Log.Info("更新c7n项目缓存完成")
}
