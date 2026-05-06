package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/k3a/html2text"
	"github.com/vanng822/go-premailer/premailer"
	"gopkg.in/gomail.v2"

	g "gin-blog/internal/global"
)

type EmailData struct {
	UserName string
	Subject  string
	Code     string
}

// SendCodeEmail 发送含验证码的邮件
func SendCodeEmail(email string, data *EmailData) error {
	conf := g.GetConfig().Email
	host := conf.Host
	port := conf.Port
	user := conf.SmtpUser
	pass := conf.SmtpPass
	from := conf.From

	slog.Info(fmt.Sprintf("发送验证码邮件 to=%s code=%s", email, data.Code))

	htmlBody := buildCodeEmailHTML(data.UserName, data.Code)

	prem, _ := premailer.NewPremailerFromString(htmlBody, nil)
	htmlInline, _ := prem.Transform()

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", htmlInline)
	m.AddAlternative("text/plain", html2text.HTML2Text(htmlBody))

	d := gomail.NewDialer(host, port, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}

func buildCodeEmailHTML(username, code string) string {
	const tpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>注册验证码</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;padding:20px;">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:8px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <h2 style="color:#333;">您好，{{.Username}}</h2>
    <p style="color:#555;">您正在注册账号，验证码为：</p>
    <div style="font-size:36px;font-weight:bold;letter-spacing:8px;color:#4caf50;text-align:center;margin:24px 0;">{{.Code}}</div>
    <p style="color:#999;font-size:13px;">验证码10分钟内有效，请勿泄露给他人。</p>
  </div>
</body>
</html>`

	t, _ := template.New("code").Parse(tpl)
	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{"Username": username, "Code": code})
	return buf.String()
}
