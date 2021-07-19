package c7n

import (
	"fmt"
	"net/url"
	"strings"

	"gitee.com/RandolphCYG/akita/pkg/cache"
	"github.com/kirinlabs/HttpRequest"
	log "github.com/sirupsen/logrus"
)

// C7nToken 获取 c7n token 接口返回数据结构体
type C7nToken struct {
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

// 为c7n用户分配角色
func FetchToken() (header map[string]string, err error) {
	// 从缓存取url
	c7nFetchTokenUrl, err := cache.HGet("third_party_sys_cfg", "c7n_fetch_token_url")
	if err != nil {
		log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	var resp C7nToken
	req := HttpRequest.NewRequest()
	respFetchToken, err := req.Post(c7nFetchTokenUrl)
	if err != nil {
		log.Error("Fail to fetch token,err: ", err)
		return
	}
	defer respFetchToken.Close() // 关闭
	// 反序列化
	err = respFetchToken.Json(&resp)
	if err != nil {
		log.Error("Fail to convert response to json,err: ", err)
		return
	}
	if resp.Error != "" {
		log.Error(resp.ErrorDescription)
		return
	}
	header = map[string]string{
		"Authorization": resp.TokenType + " " + resp.AccessToken,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	return header, nil
}

// 查询 c7n 项目 返回唯一值，忽略其他结果
func FetchC7nProject(projectName string) (c7nProject C7nProjectFields, err error) {
	// 从缓存取url
	fetchAllProjectsUrl, err := cache.HGet("third_party_sys_cfg", "fetch_all_projects_url")
	if err != nil {
		log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Error("Fail to fetch token,err: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respFetchData, err := req.Get(fetchAllProjectsUrl)
	if err != nil {
		log.Error("Fail to fetch c7n project,err: ", err)
		return
	}
	defer respFetchData.Close() // 关闭

	// 反序列化
	var c7nPros []C7nProjectFields
	err = respFetchData.Json(&c7nPros)
	if err != nil {
		log.Error("Fail to convert response to json,err: ", err)
		return
	}
	// 数据筛选
	projectName = strings.ToUpper(projectName) // 将待查询项目大写化
	for _, project := range c7nPros {
		if project.Enabled {
			// 先判断有没有相等的
			if project.Name == projectName {
				c7nProject = project
				return
			} else if strings.Contains(project.Name, projectName) {
				c7nProject = project
				return
			}
		}
	}
	return
}

// FtechC7nUser 查询 c7n 用户 返回唯一值，忽略其他结果
func FtechC7nUser(userName string) (c7nUser C7nUserFields, err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Error("Fail to fetch token,err: ", err)
		return
	}

	// 从缓存取url
	fetchC7nUserUrl, err := cache.HGet("third_party_sys_cfg", "fetch_c7n_user_url")
	if err != nil {
		log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	fetchC7nUserUrlEncoded := fmt.Sprintf(fetchC7nUserUrl, url.QueryEscape(userName)) // url中中文部分编码为url
	respFetchC7nUser, err := req.Get(fetchC7nUserUrlEncoded)
	if err != nil {
		log.Error("Fail to fetch c7n user,err: ", err)
		return
	}
	defer respFetchC7nUser.Close() // 关闭

	// 反序列化
	var c7nFetchUserRes C7nFetchUserRes
	err = respFetchC7nUser.Json(&c7nFetchUserRes)
	if err != nil {
		log.Error("Fail to convert response to json,err: ", err)
		return
	}

	// 数据筛选
	c7nUser = c7nFetchUserRes.Content[0]
	return
}

// FetchC7nRoles 查询 c7n 角色列表
func FetchC7nRoles(roleName string) (c7nRole C7nRoleFields, err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Error("Fail to fetch token,err: ", err)
		return
	}

	// 从缓存取url
	fetchC7nRolesUrl, err := cache.HGet("third_party_sys_cfg", "fetch_c7n_roles_url")
	if err != nil {
		log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	respFetchC7nRoles, err := req.Get(fetchC7nRolesUrl)
	if err != nil {
		log.Error("Fail to fetch c7n user,err: ", err)
		return
	}
	defer respFetchC7nRoles.Close() // 关闭

	// 反序列化
	var c7nFetchRoleRes C7nFetchRoleRes
	err = respFetchC7nRoles.Json(&c7nFetchRoleRes)
	if err != nil {
		log.Error("Fail to convert response to json,err: ", err)
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

// 为c7n用户分配某项目的某角色
func AssignC7nUserProjectRole(c7nProjectId string, c7nUserId string, c7nRoleIds []string) (err error) {
	// 取token
	header, err := FetchToken()
	if err != nil {
		log.Error("Fail to fetch token,err: ", err)
		return
	}

	// 从缓存取url
	assignC7nUserProjectRoleUrl, err := cache.HGet("third_party_sys_cfg", "assign_c7n_user_project_role_url")
	if err != nil {
		log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}

	// 发送请求
	req := HttpRequest.NewRequest()
	req.SetHeaders(header)
	assignC7nUserProjectRoleUrlEncoded := fmt.Sprintf(assignC7nUserProjectRoleUrl, c7nProjectId, c7nUserId)
	respAssignC7nUserProjectRole, err := req.Put(assignC7nUserProjectRoleUrlEncoded, c7nRoleIds)
	log.Info(respAssignC7nUserProjectRole)
	return
}