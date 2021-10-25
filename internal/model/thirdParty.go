package model

/*
* 第三方系统配置
*
 */

// 第三方系统配置
type ThirdPartyCfg struct {
	Key   string `json:"key" gorm:"type:varchar(255);not null;comment:模板键值"`
	Value string `json:"value" gorm:"type:varchar(4000);not null;comment:模板内容"`
}

// CreateThirdPartyCfg 增 第三方系统配置
func CreateThirdPartyCfg(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create((&ThirdPartyCfg{Key: key, Value: value}))
}

// DeleteThirdPartyCfg 删 第三方系统配置
func DeleteThirdPartyCfg(key string) {
	DB.Where("key = ?", key).Delete(&ThirdPartyCfg{})
}

// UpdateThirdPartyCfg 改 第三方系统配置
func UpdateThirdPartyCfg(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create((&ThirdPartyCfg{Key: key, Value: value}))
}

// FetchThirdPartyCfg 查 第三方系统配置
func FetchThirdPartyCfg(key string) (thirdPartyCfg ThirdPartyCfg, err error) {
	DB.Where("key = ?", key).First(&thirdPartyCfg)
	return
}

// FetchThirdPartyCfgs 查 第三方系统配置列表
func FetchThirdPartyCfgs() (thirdPartyCfgs []ThirdPartyCfg, err error) {
	DB.Find(&thirdPartyCfgs)
	return
}
