// Package mailer 腾讯 SMTP 邮件发送（smtp.qq.com:465，隐式 TLS）。
package mailer

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// FromEnv 从环境变量读取 SMTP 配置；未配置时 Send 走开发降级（返回错误由调用方处理）。
func FromEnv() Config {
	return Config{
		Host:     env("SMTP_HOST", "smtp.qq.com"),
		Port:     env("SMTP_PORT", "465"),
		Username: env("SMTP_USER", ""),
		Password: env("SMTP_PASSWORD", ""),
		From:     env("SMTP_FROM", ""),
	}
}

func (c Config) Enabled() bool {
	return c.Username != "" && c.Password != ""
}

// Send 发送一封纯文本邮件。
func (c Config) Send(to, subject, body string) error {
	from := c.From
	if from == "" {
		from = c.Username
	}
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
		"",
	}, "\r\n")

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", c.Host+":"+c.Port, &tls.Config{MinVersion: tls.VersionTLS12, ServerName: c.Host})
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return fmt.Errorf("smtp deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
