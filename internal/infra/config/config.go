// Package config 统一加载运行时配置：main 加载后注入 wire，wire 不再写死默认值。
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DB        DB
	HTTPAddr  string
	JWTSecret string
	JWTExpire time.Duration
	Admin     Admin
	Demo      Demo
}

// DB 数据库连接配置；DSNOverride 设置时优先使用完整 DSN。
type DB struct {
	DSNOverride string
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	SSLMode     string
	TimeZone    string
}

type Admin struct {
	Username string
	Password string
}

type Demo struct {
	Username string
	Password string
}

// Load 从环境变量读取配置；未设置时使用本机开发默认值。
func Load() Config {
	return Config{
		DB: DB{
			DSNOverride: env("DB_DSN", ""),
			Host:        env("DB_HOST", "localhost"),
			Port:        env("DB_PORT", "5432"),
			User:        env("DB_USER", "4ge0"),
			Password:    env("DB_PASSWORD", ""),
			Name:        env("DB_NAME", "cinema"),
			SSLMode:     env("DB_SSLMODE", "disable"),
			TimeZone:    env("DB_TIMEZONE", "Asia/Shanghai"),
		},
		HTTPAddr:  env("HTTP_ADDR", ":8080"),
		JWTSecret: env("JWT_SECRET", "dev-secret"),
		JWTExpire: time.Duration(envInt("JWT_EXPIRE_HOURS", 24)) * time.Hour,
		Admin: Admin{
			Username: env("ADMIN_USERNAME", "admin"),
			Password: env("ADMIN_PASSWORD", "admin123"),
		},
		Demo: Demo{
			Username: env("DEMO_USERNAME", "demo"),
			Password: env("DEMO_PASSWORD", "demo123"),
		},
	}
}

func (d DB) DSN() string {
	if d.DSNOverride != "" {
		return d.DSNOverride
	}
	parts := []string{
		"host=" + d.Host,
		"user=" + d.User,
		"port=" + d.Port,
		"dbname=" + d.Name,
		"sslmode=" + d.SSLMode,
		"TimeZone=" + d.TimeZone,
	}
	if d.Password != "" {
		parts = append(parts, "password="+d.Password)
	}
	return strings.Join(parts, " ")
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
