package mysqlmodule

import (
	"fmt"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	driver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
	"github.com/soryetong/greasyx/console"
)

func init() {
	console.Append(mysqlCmd)
}

var mysqlCmd = &cobra.Command{
	Use:   "MySQL",
	Short: "Init MySQL",
	Long:  `加载MySQL模块之后，可以通过 gina.Db 进行ORM操作`,
	Run: func(cmd *cobra.Command, args []string) {
		dsn := viper.GetString("MySQL.Dsn")
		if dsn == "" {
			console.Echo.Fatalf("❌ 错误: 你正在加载MySQL模块，但是你未配置MySQL.Dsn，请先添加配置\n")
		}

		initGorm(dsn)
	},
}

func initGorm(dsn string) {
	viper.SetDefault("MySQL.LogLevel", 3)
	viper.SetDefault("MySQL.EnableLogWriter", true)
	viper.SetDefault("MySQL.MaxIdleConn", 10)
	viper.SetDefault("MySQL.MaxConn", 200)
	viper.SetDefault("MySQL.SlowThreshold", 200)
	dsn = ensureTimeout(dsn, "5s")
	mysqlConfig := driver.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         255,   // string 类型字段的默认长度
		DisableDatetimePrecision:  false, // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}
	db, err := gorm.Open(driver.New(mysqlConfig), &gorm.Config{
		Logger: getLogger(),
	})
	if err != nil {
		console.Echo.Fatalf("❌ 错误: MySQL连接失败: %s\n", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(viper.GetInt("MySQL.MaxIdleConn"))
	sqlDB.SetMaxOpenConns(viper.GetInt("MySQL.MaxConn"))

	gina.Db = db
	console.Echo.Info("ℹ️ 提示: MySQL模块加载成功, 你可以使用 `gina.Db` 进行ORM操作\n")
}

// 切换默认 Logger 使用的 Writer
func getLogger() logger.Interface {
	logLevel := viper.GetInt("MySQL.LogLevel")
	var logMode logger.LogLevel
	switch logLevel {
	case 1:
		logMode = logger.Error
	case 2:
		logMode = logger.Warn
	case 3:
		logMode = logger.Info
	default:
		logMode = logger.Info
	}

	enableWriter := viper.GetBool("MySQL.EnableLogWriter")
	return logger.New(getLogWriter(enableWriter), logger.Config{
		SlowThreshold:             time.Duration(viper.GetInt("MySQL.SlowThreshold")) * time.Millisecond,
		LogLevel:                  logMode,
		IgnoreRecordNotFoundError: true,
		Colorful:                  !enableWriter,
	})
}

// 自定义 Writer
func getLogWriter(enableWriter bool) logger.Writer {
	logPath := viper.GetString("Log.Path")
	var writer io.Writer
	if enableWriter {
		if !strings.HasSuffix(logPath, "/") {
			logPath += "/"
		}
		fileName := fmt.Sprintf("%s%s/mysqlmodule.log", logPath, time.Now().Format("2006-01-02"))
		writer = &lumberjack.Logger{
			Filename:   path.Join(logPath, fileName),
			MaxSize:    viper.GetInt("Log.MaxSize"),    // 单文件最大容量, 单位是MB
			MaxBackups: viper.GetInt("Log.MaxBackups"), // 最大保留过期文件个数
			MaxAge:     viper.GetInt("Log.MaxAge"),     // 保留过期文件的最大时间间隔, 单位是天
			Compress:   viper.GetBool("Log.Compress"),  // 是否需要压缩滚动日志, 使用的gzip压缩
			LocalTime:  true,                           // 是否使用计算机的本地时间, 默认UTC
		}
	} else {
		writer = os.Stdout
	}

	return log.New(writer, "\r\n", log.LstdFlags)
}

// 确保连接字符串中存在 timeout 参数
func ensureTimeout(connStr, defaultTimeout string) string {
	parts := strings.SplitN(connStr, "?", 2)
	if len(parts) == 1 {
		return connStr + "?timeout=" + defaultTimeout
	}

	base := parts[0]
	query := parts[1]
	params, err := url.ParseQuery(query)
	if err != nil {
		return connStr
	}

	if _, exists := params["timeout"]; !exists {
		params.Set("timeout", defaultTimeout)
	}

	return base + "?" + params.Encode()
}
