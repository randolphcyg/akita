package email

import (
	"crypto/tls"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var log = logrus.New()

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

// 初始化连接
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

func SendMailHtml(address []string, subject string, htmlContent string) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", EmailCfg.Username)
	m.SetHeader("To", address[0])
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlContent)
	port, err := strconv.Atoi(EmailCfg.Port)
	d := gomail.NewDialer(EmailCfg.Host, port, EmailCfg.Address, EmailCfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Send the email
	if err = d.DialAndSend(m); err != nil {
		return
	}
	return
}

// // 发送 html 邮件
// func SendMailHtml(address []string, subject string, htmlContent string) (err error) {
// 	fmt.Println("发邮件")
// 	e := email.NewEmail()
// 	auth := smtp.PlainAuth("", EmailCfg.Username, EmailCfg.Password, EmailCfg.Host)
// 	fmt.Println(auth)
// 	//设置发送方的邮箱
// 	e.From = EmailCfg.Address
// 	//设置主题
// 	e.Subject = subject
// 	//设置文件发送的内容
// 	e.HTML = []byte(htmlContent)

// 	for _, v := range address {
// 		// 设置接收方的邮箱
// 		e.To = []string{v}
// 		//设置服务器相关的配置
// 		addr := fmt.Sprintf("%s:%s", EmailCfg.Host, EmailCfg.Port)
// 		fmt.Println(addr)
// 		// 发送
// 		err = e.SendWithTLS(addr, auth, &tls.Config{InsecureSkipVerify: true})
// 		if err != nil {
// 			return
// 		}
// 	}
// 	return
// }

// // 发送 text 邮件
// func SendMailText(address []string, subject string, body string) (err error) {
// 	// 通常身份应该是空字符串，填充用户名.
// 	auth := smtp.PlainAuth("", EmailCfg.Username, EmailCfg.Password, EmailCfg.Host)
// 	contentType := "Content-Type: text/html; charset=UTF-8"

// 	for _, v := range address {
// 		s := fmt.Sprintf("To:%s\r\nFrom:%s<%s>\r\nSubject:%s\r\n%s\r\n\r\n%s",
// 			v, EmailCfg.NickName, EmailCfg.Username, subject, contentType, body)
// 		msg := []byte(s)
// 		addr := fmt.Sprintf("%s:%s", EmailCfg.Host, EmailCfg.Host)
// 		err = smtp.SendMail(addr, auth, EmailCfg.Username, []string{v}, msg)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return
// }
