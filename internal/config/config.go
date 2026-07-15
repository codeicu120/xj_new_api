package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost            = "0.0.0.0"
	defaultPort            = "8080"
	defaultEnv             = "dev"
	defaultShutdownTimeout = 10 * time.Second
	defaultResourceHost    = "aqsmimg365.sbs:10002"
	defaultIPDBPath        = "/Users/canavs/xjProj/XJBackend/api/data/ipipfree.ipdb"
	defaultMySQLDSN        = "xj_app:xj_app_123456@tcp(127.0.0.1:3306)/xj_comp?charset=utf8mb4&parseTime=true&loc=Local"
	defaultGameResourceURL = "https://image.xjdev.one"
	defaultUploadPath      = "/Users/canavs/xjProj/XJBackend/api/res"
)

type Config struct {
	Env             string
	Host            string
	Port            string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	ResourceBaseURL string
	SMSCaptcha      int
	CaptchaStyle    int
	IPDBPath        string
	MySQLDSN        string
	GameResourceURL string
	VIPDiscount     int
	UploadPath      string
}

func FromEnv() Config {
	return Config{
		Env:             envString("APP_ENV", defaultEnv),
		Host:            envString("HTTP_HOST", defaultHost),
		Port:            envString("HTTP_PORT", defaultPort),
		LogLevel:        envLogLevel("LOG_LEVEL", slog.LevelInfo),
		ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		ResourceBaseURL: envString("RESOURCE_BASE_URL", defaultResourceBaseURL(time.Now())),
		SMSCaptcha:      envInt("SMS_CAPTCHA", 1),
		CaptchaStyle:    envInt("CAPTCHA_STYLE", 0),
		IPDBPath:        envString("IPDB_PATH", defaultIPDBPath),
		MySQLDSN:        envString("MYSQL_DSN", defaultMySQLDSN),
		GameResourceURL: envString("GAME_RESOURCE_BASE_URL", defaultGameResourceURL),
		VIPDiscount:     envInt("VIP_DISCOUNT", 50),
		UploadPath:      envString("UPLOAD_PATH", defaultUploadPath),
	}
}

func (c Config) HTTPAddr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envLogLevel(key string, fallback slog.Level) slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		if level, err := strconv.Atoi(os.Getenv(key)); err == nil {
			return slog.Level(level)
		}
		return fallback
	}
}

func defaultResourceBaseURL(now time.Time) string {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		now = now.In(loc)
	} else {
		now = now.UTC().Add(8 * time.Hour)
	}
	return fmt.Sprintf("https://%s.%s", now.Format("2006010215"), defaultResourceHost)
}
