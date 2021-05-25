package user

import (
	"fmt"
	"strings"

	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/ldap"
	"gitee.com/RandolphCYG/akita/pkg/log"
	"gitee.com/RandolphCYG/akita/pkg/serializer"
	"gitee.com/RandolphCYG/akita/pkg/wework/api"
	"gitee.com/RandolphCYG/akita/pkg/wework/order"
	"github.com/goinggo/mapstructure"
)

// LdapUserService LDAP用户查询条件
type LdapUserService struct {
	// 用户对象类
	UserClass string `json:"user_class" gorm:"type:varchar(50);not null;comment:用户对象类"`
	// 用户对象过滤
	UserFilter string `json:"user_filter" gorm:"type:varchar(255);not null;comment:用户对象过滤"`
	// 组织架构类
	OrganizationClass string `json:"organization_class" gorm:"type:varchar(50);not null;comment:组织架构类"`
	// 用户名
	Username string `json:"username" gorm:"type:varchar(50);not null;comment:用户名"`
	// 中文名
	DisplayName string `json:"display_name" gorm:"type:varchar(50);not null;comment:中文名"`
	// 邮箱
	Email string `json:"email" gorm:"type:varchar(50);not null;comment:邮箱"`
	// 手机号
	Mobile string `json:"mobile" gorm:"type:varchar(50);not null;comment:手机号"`
}

// 查
func (service *LdapUserService) FetchUser(url string) serializer.Response {
	// 初始化连接
	user := &ldap.LdapAttributes{}
	LdapUsers := ldap.FetchLdapUsers(user)
	for _, user := range LdapUsers {
		fmt.Println(user.DN)
		break
	}
	return serializer.Response{Data: LdapUsers}
}

// 将原始工单转换为UUAP创建工单详情结构体 以后这种方法通用与原始工单的解析
func originalOrderToUuapCreateOrder(weworkOrder map[string]interface{}) (weworkOrderDetails order.WeworkOrderDetails) {
	if err := mapstructure.Decode(weworkOrder, &weworkOrderDetails); err != nil {
		log.Log().Error("原始工单转换错误:%v", err)
	}
	return
}

// 创建用户
func (service *LdapUserService) AddUser(u LdapUserService) serializer.Response {
	// service层处理前端数据，并将数据传给pkg的ldap组件，然后ldap组件处理有关ldap用户的通用逻辑
	// TODO: 这边改成从tabby接受工单request请求
	sp_no := 202105250041
	// 实例化 API 类 TODO: 这边将企业微信配置保存到数据库，系统初始化时加载到环境变量中
	corpAPI := api.NewCorpAPI("XXXXXXX", "XXXXXXX")
	// 发送消息
	response, err := corpAPI.GetApprovalDetail(map[string]interface{}{
		"sp_no": sp_no,
	})

	if err != nil {
		log.Log().Error("%v", err)
	}
	// log.Log().Info("%v", response["info"])

	var weworkOrder order.WeworkOrder
	// 反序列化工单详情
	if err := mapstructure.Decode(response["info"], &weworkOrder); err != nil {
		log.Log().Error("%v", err)
	}
	// 清洗工单
	orderData := make(map[string]interface{})
	orderData["spName"] = weworkOrder.SpName
	orderData["partyid"] = weworkOrder.Applyer.Partyid
	orderData["userid"] = weworkOrder.Applyer.Userid
	for _, con := range weworkOrder.ApplyData.Contents {
		switch con.Control {
		case "Number":
			orderData[con.Title[0].Text] = con.Value.NewNumber
		case "Text":
			orderData[con.Title[0].Text] = con.Value.Text
		case "Textarea":
			orderData[con.Title[0].Text] = con.Value.Text
		case "Date":
			orderData[con.Title[0].Text] = con.Value.Date.STimestamp
		case "Selector":
			if con.Value.Selector.Type == "multi" { // 多选
				tempSelectors := make([]string, len(con.Value.Selector.Options)-2)
				for _, value := range con.Value.Selector.Options {
					tempSelectors = append(tempSelectors, value.Value[0].Text)
				}
				orderData[con.Title[0].Text] = tempSelectors
			} else { // 单选
				orderData[con.Title[0].Text] = con.Value.Selector.Options[0].Value[0].Text
			}
		case "Tips": // 忽略说明类型
			continue
		case "Contact":
			orderData[con.Title[0].Text] = con.Value.Members
		default:
			log.Log().Error("包含未处理工单项类型：%v，请及时补充后端逻辑!", con.Control)
		}
	}

	log.Log().Info("res:%v", orderData)
	// 将原始工单结构体转换为对应要求工单数据
	weworkOrderDetails := originalOrderToUuapCreateOrder(orderData)
	log.Log().Info("%v", weworkOrderDetails)

	company := "XX公司" // 这个数据工单是没有的 需要考虑如何添加
	name := []rune(weworkOrderDetails.Name)
	user := &ldap.LdapAttributes{
		Dn:          "CN=" + string(name) + weworkOrderDetails.Eid + "," + ldap.DepartToDn(weworkOrderDetails.Depart),
		Num:         weworkOrderDetails.Eid,
		Sam:         weworkOrderDetails.Eid,
		AccountCtl:  "544",
		Expire:      ldap.ExpireTime(int64(-1)), // 永不过期
		Sn:          string(name[0]),
		PwdLastSet:  "0",
		DisplayName: string(name),
		GivenName:   string(name[1:]),
		Email:       weworkOrderDetails.Mail,
		Phone:       weworkOrderDetails.Mobile,
		Company:     company,
		Depart:      strings.Split(weworkOrderDetails.Depart, ".")[0],
		Title:       weworkOrderDetails.Title,
	}
	log.Log().Info("%v", user)
	//  TODO 未完善前置逻辑 暂时注释掉创建功能
	// res := ldap.AddUser(user)
	// for index, r := range res {
	// 	fmt.Println(index, r)
	// }
	// TODO 后面一步骤是创建成功发送企业微信消息

	return serializer.Response{Data: 0, Msg: "UUAP账号创建成功"}
}

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

// 查询HR数据
func (service *HrDataService) FetchHrData(h HrDataService) serializer.Response {
	// 初始化连接
	// var c conn.LdapConnService

	// conn, err := c.FetchByConnUrl(url)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// LdapUsers := ldap.FetchUser(&conn)
	// for _, user := range LdapUsers {
	// 	user.PrettyPrint(2)
	// 	fmt.Println(user.GetAttributeValue("displayName"))
	// 	break
	// }
	return serializer.Response{Data: 1111}
}

// 更新HR数据到LDAP实现逻辑
// 此部分逻辑将最终修改为手动和自动两种调用方式
func (service *HrDataService) HrToLdap(h HrDataService) serializer.Response {
	// 获取HR接口数据
	var hrDataConn hr.HrDataConn
	if result := model.DB.First(&hrDataConn); result.Error != nil {
		log.Log().Error("没有外部HR数据连接信息")
	}

	hrConn := &hr.HrDataConn{
		UrlGetToken: hrDataConn.UrlGetToken,
		UrlGetData:  hrDataConn.UrlGetData,
	}

	hrUsers := hr.FetchHrData(hrConn)

	// 全量遍历HR接口数据用户 并 更新LDAP用户
	var userStat, dn string
	var expire int64

	for _, user := range hrUsers {
		if user.Stat == "离职" { //离职员工
			userStat = "546"
			dn = "OU=disabled," + ldap.LdapCfg.BaseDn
			expire = 0 // 账号失效
		} else { // 在职员工
			userStat = "544"
			dn = ldap.DepartToDn(user.Department)
			expire = ldap.ExpireTime(-1) // 账号永久有效
		}
		depart := strings.Split(user.Department, ".")[len(strings.Split(user.Department, "."))-1]
		name := []rune(user.Name)

		// 将hr数据转换为ldap信息格式
		user := &ldap.LdapAttributes{
			Num:         user.Eid,
			Sam:         user.Eid,
			DisplayName: user.Name,
			Email:       user.Mail,
			Phone:       user.Mobile,
			Dn:          dn,
			PwdLastSet:  "0", // 用户下次必须修改密码 0
			AccountCtl:  userStat,
			Expire:      expire,
			Sn:          string(name[0]),
			Name:        user.Name,
			GivenName:   string(name[1:]),
			Company:     user.CompanyName,
			Depart:      depart,
			Title:       user.Title,
		}
		// 更新用户信息 [注意对外部员工OU路径要插入本公司/合作伙伴一层中!!!]
		user.Update()
	}
	return serializer.Response{Data: 1, Msg: "更新成功"}
}
