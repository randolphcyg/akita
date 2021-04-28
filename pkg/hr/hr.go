package hr

import (
	"gitee.com/RandolphCYG/akita/pkg/log"
	"github.com/asmcos/requests"
)

// HrToken 获取token接口返回数据结构体
type HrToken struct {
	// 正确时候
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	// 错误时候
	Code             int    `json:"code"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
	Success          bool   `json:"success"`
}

// HrData HR数据接口返回数据结构体
type HrData struct {
	Content          []HrUser `json:"content"`
	Empty            bool     `json:"empty"`
	Number           int      `json:"number"`
	NumberOfElements int      `json:"numberOfElements"`
	Size             int      `json:"size"`
	TotalElements    int      `json:"totalElements"`
	TotalPages       int      `json:"totalPages"`
	// 出错时候
	Result string `json:"result"`
}

// HrUser 数据接口查询的用户信息结构体
type HrUser struct {
	CompanyCode string `json:"company_code"`
	CompanyName string `json:"company_name"`
	Name        string `json:"ename"`
	Department  string `json:"org_all"`
	Eid         string `json:"pernr"`
	Stat        string `json:"stat2"`
	Mobile      string `json:"usrid"`
	Mail        string `json:"usrid_long"`
	Title       string `json:"zmplans"`
}

// HrDataConn HR数据模型
type HrDataConn struct {
	// 获取 token 的 URL
	UrlGetToken string `json:"url_get_token" gorm:"type:varchar(255);not null;comment:获取token的地址"`
	// 获取 数据 的URL
	UrlGetData string `json:"url_get_data" gorm:"type:varchar(255);not null;comment:获取数据的地址"`
}

// FetchToken 获取token
func FetchToken(h *HrDataConn) (token HrToken) {
	req := requests.Requests()
	respFetchToken, err := req.Post(h.UrlGetToken)
	if err != nil {
		// 抛错
		log.Log().Error("fetch token failed,err:%v\n", err)
		return
	}
	// 反序列化
	err = respFetchToken.Json(&token)
	if err != nil {
		// 抛错
		log.Log().Error("convert response to json failed,err:%v\n", err)
		return
	}
	if !token.Success {
		// 抛错
		log.Log().Error(token.ErrorDescription)
		return
	}
	return
}

// FetchHrData 带着token去获取HR数据
func FetchHrData(h *HrDataConn) (hrUsers []HrUser) {
	req := requests.Requests()
	hrToken := FetchToken(h)
	header := requests.Header{
		"Authorization": hrToken.TokenType + " " + hrToken.AccessToken,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	// 发送请求
	respFetchData, err := req.Post(h.UrlGetData, header)
	if err != nil {
		log.Log().Error("fetch hr data failed,err:%v\n", err)
		return
	}
	var hrdata HrData
	respFetchData.Json(&hrdata)
	// 返回数据是否有报错字段
	if hrdata.Result != "" {
		log.Log().Error("fetch hr data failed,err%v\n", hrdata.Result)
		return
	}
	hrUsers = hrdata.Content
	return
}