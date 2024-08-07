package test

import (
	"crypto/tls"
	"github.com/jordan-wright/email"
	"net/smtp"
	"testing"
)

func TestSendEmail(t *testing.T) {
	e := email.NewEmail()
	e.From = "Cz <15637662613@163.com>"
	e.To = []string{"1508847610@qq.com"}
	e.Subject = "验证码发送测试"
	e.HTML = []byte("您的验证码是：<b>123456</b>")
	//err := e.Send("smtp.163.com:465", smtp.PlainAuth("", "15637662613@163.com", "OOYTFVQUSZNBVFVM", "smtp.163.com"))
	//返回EOF时，关闭SSL重试
	err := e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth("", "15637662613@163.com", "OOYTFVQUSZNBVFVM", "smtp.163.com"),
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "smtp.163.com",
		})
	if err != nil {
		t.Fatal(err)
		return
	}
}
