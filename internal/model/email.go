package model

/*
* 邮件模板 缓存
*
 */

// 邮件模板
type EmailTemplate struct {
	Key   string `json:"key" gorm:"type:varchar(255);not null;comment:模板键值"`
	Value string `json:"value" gorm:"type:varchar(4000);not null;comment:模板内容"`
}

// CreateEmailTemplate 增 邮件模板
func CreateEmailTemplate(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create((&EmailTemplate{Key: key, Value: value}))
}

// DeleteEmailTemplate 删 邮件模板
func DeleteEmailTemplate(key string) {
	DB.Where("key = ?", key).Delete(&EmailTemplate{})
}

// UpdateEmailTemplate 改 邮件模板
func UpdateEmailTemplate(key, value string) {
	DB.Model(&WeworkMsgTemplate{}).Create((&EmailTemplate{Key: key, Value: value}))
}

// FetchEmailTemplate 查 邮件模板
func FetchEmailTemplate(key string) (emailTemplate EmailTemplate, err error) {
	DB.Where("key = ?", key).First(&emailTemplate)
	return
}

// FetchEmailTemplates 查 邮件模板列表
func FetchEmailTemplates() (emailTemplates []EmailTemplate, err error) {
	DB.Find(&emailTemplates)
	return
}
