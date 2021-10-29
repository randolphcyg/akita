package email

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Config email config
type Config struct {
	Host      string
	Port      string
	Username  string
	Password  string
	NickName  string
	Address   string
	ReplyTo   string
	KeepAlive int
}

var (
	EmailCfg *Config
)

// Init 初始化邮件配置
func Init(c *Config) (err error) {
	EmailCfg = &Config{
		Host:      c.Host,
		Port:      c.Port,
		Username:  c.Username,
		Password:  c.Password,
		NickName:  c.NickName,
		Address:   c.Address,
		ReplyTo:   c.ReplyTo,
		KeepAlive: c.KeepAlive,
	}
	return
}

// SendMailHtml 发送 html 邮件
func SendMailHtml(address []string, subject string, htmlContent string) (err error) {
	e := email.NewEmail()
	auth := smtp.PlainAuth("", EmailCfg.Username, EmailCfg.Password, EmailCfg.Host)
	//设置发送方的邮箱
	e.From = EmailCfg.Address
	//设置主题
	e.Subject = subject
	//设置文件发送的内容
	e.HTML = []byte(htmlContent)

	for _, v := range address {
		// 设置接收方的邮箱
		e.To = []string{v}
		//设置服务器相关的配置
		addr := fmt.Sprintf("%s:%s", EmailCfg.Host, EmailCfg.Port)
		// 发送
		err = e.Send(addr, auth) // , &tls.Config{InsecureSkipVerify: true}
		if err != nil {
			return
		}
	}
	return
}

// SendMailText 发送 text 邮件
func SendMailText(address []string, subject string, body string) (err error) {
	// 通常身份应该是空字符串，填充用户名.
	auth := smtp.PlainAuth("", EmailCfg.Username, EmailCfg.Password, EmailCfg.Host)
	contentType := "Content-Type: text/html; charset=UTF-8"

	for _, v := range address {
		s := fmt.Sprintf("To:%s\r\nFrom:%s<%s>\r\nSubject:%s\r\n%s\r\n\r\n%s",
			v, EmailCfg.NickName, EmailCfg.Username, subject, contentType, body)
		msg := []byte(s)
		addr := fmt.Sprintf("%s:%s", EmailCfg.Host, EmailCfg.Host)
		err = smtp.SendMail(addr, auth, EmailCfg.Username, []string{v}, msg)
		if err != nil {
			return
		}
	}
	return
}
