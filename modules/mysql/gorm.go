package mysql

import (
	"gorm.io/gorm/logger"
	"gorm.io/gorm"
	"fmt"
	"time"
	"os"
	"log"
	"io"
	driver "gorm.io/driver/mysql"
	"github.com/soryetong/greasyx/console"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
	"github.com/soryetong/greasyx/utils"
	"github.com/soryetong/greasyx/gina"
	"strings"
	"net/url"
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
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] "+
				"你正在加载MySQL模块，但是你未配置MySQL.Dsn，请先添加配置\n")
			os.Exit(124)
		}

		gina.Db = initGorm(dsn, "./")
		_, _ = fmt.Fprintf(os.Stderr, "\n\033[32m [GREASYX-info] "+
			"MySQL模块加载成功, 你可以使用 `gina.Db` 进行ORM操作\033[0m\n")
	},
}

func initGorm(dsn, logPath string) *gorm.DB {
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
		Logger: getLogger(3, true, logPath),
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-error] "+
			"MySQL连接失败: %s\n", err))
		os.Exit(124)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(200)

	return db
}

// 切换默认 Logger 使用的 Writer
func getLogger(logLevel int64, enableWriter bool, logPath string) logger.Interface {
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

	return logger.New(getLogWriter(enableWriter, logPath), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logMode,
		IgnoreRecordNotFoundError: true,
		Colorful:                  !enableWriter,
	})
}

// 自定义 Writer
func getLogWriter(enableWriter bool, logPath string) logger.Writer {
	var writer io.Writer
	if enableWriter {
		fileName := fmt.Sprintf("%s_sql.log", time.Now().Format("20060102"))
		logFileName := path.Join(logPath, fileName)
		writer = utils.Tool().NewLumberjack(logFileName)
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
