package hr

import (
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"github.com/pkg/errors"

	"github.com/kirinlabs/HttpRequest"
)

// TokenResp 获取token接口返回数据结构体
type TokenResp struct {
	// 正确时候
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	// 错误时候
	Code             string `json:"code"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
	Success          bool   `json:"success"`
}

// DataResp HR数据接口返回数据结构体
type DataResp struct {
	Content          []User `json:"content"`
	Empty            bool   `json:"empty"`
	Number           int    `json:"number"`
	NumberOfElements int    `json:"numberOfElements"`
	Size             int    `json:"size"`
	TotalElements    int    `json:"totalElements"`
	TotalPages       int    `json:"totalPages"`
	// 出错时候
	Result string `json:"result"`
}

// User 数据接口查询的用户信息结构体
type User struct {
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
func (h *HrDataConn) FetchToken() (tokenResp TokenResp, err error) {
	req := HttpRequest.NewRequest()
	respFetchToken, err := req.Post(h.UrlGetToken)
	if err != nil {
		// 抛错
		err = errors.Wrap(err, serializer.ErrGetToken)
		return
	}
	// 反序列化
	err = respFetchToken.Json(&tokenResp)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrConvertRespToJson)
		return
	}
	if !tokenResp.Success && tokenResp.ErrorDescription != "" {
		// 抛错
		err = errors.New(tokenResp.ErrorDescription)
		return
	}
	return
}

// FetchData 带着token去获取HR数据
func (h *HrDataConn) FetchData() (users []User, err error) {
	req := HttpRequest.NewRequest()
	hrToken, err := h.FetchToken()
	if err != nil {
		return
	}
	header := map[string]string{
		"Authorization": hrToken.TokenType + " " + hrToken.AccessToken,
		"Content-Type":  "application/json;charset=UTF-8",
	}
	// 发送请求
	req.SetHeaders(header)
	respFetchData, err := req.Post(h.UrlGetData)
	if err != nil {
		err = errors.Wrap(err, serializer.ErrFetchHrData)
		return
	}

	var dataResp DataResp
	err = respFetchData.Json(&dataResp)
	if err != nil {
		return nil, err
	}
	// 返回数据是否有报错字段
	if dataResp.Result != "" {
		err = errors.Wrap(err, serializer.ErrFetchHrData+dataResp.Result)
		return
	}
	users = dataResp.Content
	return
}
