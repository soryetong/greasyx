package ginahelper

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ServerAddr  string
	ServerIsTLS bool
)

func InitSugaredLogger() *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = nil
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.DisableStacktrace = true

	logger, _ := config.Build()

	return logger.Sugar()
}

// 获取操作系统
func GetPlatform(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// 移动端
	if strings.Contains(ua, "android") {
		return "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return "iOS"
	}

	// 桌面端
	if strings.Contains(ua, "windows") {
		return "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		return "MacOS"
	} else if strings.Contains(ua, "linux") {
		return "Linux"
	}

	return "Unknown"
}

// 获取浏览器类型
func GetBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Google Chrome"
	} else if strings.Contains(ua, "edg") {
		return "Microsoft Edge"
	} else if strings.Contains(ua, "firefox") {
		return "Mozilla Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Apple Safari"
	} else if strings.Contains(ua, "opr") || strings.Contains(ua, "opera") {
		return "Opera"
	} else if strings.Contains(ua, "msie") || strings.Contains(ua, "trident") {
		return "Internet Explorer"
	}

	return "Unknown"
}

func GetServerAddr() string {
	prefix := "http"
	if ServerIsTLS {
		prefix = "https"
	}

	addr := ServerAddr
	addrArr := strings.Split(ServerAddr, ":")
	if addrArr[0] == "" {
		addrArr[0] = GetLocalIP()
		addr = strings.Join(addrArr, ":")
	}

	return fmt.Sprintf("%s://%s", prefix, addr)
}
