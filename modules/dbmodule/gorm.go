package dbmodule

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var om = make(map[string]string)

func initGorm(conf *dbConfig) {
	if conf.LogLevel == 0 {
		conf.LogLevel = 3
	}
	if conf.MaxIdleConn == 0 {
		conf.MaxIdleConn = 10
	}
	if conf.MaxConn == 0 {
		conf.MaxConn = 200
	}
	if conf.SlowThreshold == 0 {
		conf.SlowThreshold = 200
	}

	dsn := ensureTimeout(conf.Dsn, "5s")
	driverArr := strings.Split(conf.Driver, "_")
	var orm gorm.Dialector
	switch strings.ToLower(driverArr[0]) {
	case gina.DbTypeMysql:
		orm = mysql.New(mysql.Config{
			DSN:                      dsn,   // DSN data source name
			DefaultStringSize:        255,   // string 类型字段的默认长度
			DisableDatetimePrecision: false, // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
			DontSupportRenameIndex:   true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
			DontSupportRenameColumn:  true,  // 用 `change` 重命名列，My
		})
	case gina.DbTypeSqlite:
		orm = sqlite.New(sqlite.Config{
			DSN: dsn,
		})
	case gina.DbTypePostgresql:
		orm = postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true, // 禁用 extended protocol
		})
	case gina.DbTypeSqlserver:
		orm = sqlserver.New(sqlserver.Config{
			DSN: dsn,
		})
	case gina.DbTypeOracle:
		console.Echo.Fatalf("❌ 错误: gorm 暂不支持 Oracle \n")
	default:
		console.Echo.Fatalf("❌ 错误: 不支持的数据库驱动类型: %s\n", conf.Driver)
	}
	db, err := gorm.Open(orm, &gorm.Config{
		Logger: getLogger(conf),
	})
	if err != nil {
		console.Echo.Fatalf("❌ 错误: MySQL连接失败: %s\n", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(conf.MaxIdleConn)
	sqlDB.SetMaxOpenConns(conf.MaxConn)

	gina.SetGorm(conf.Driver, db)
	name, ok := om[strings.ToLower(conf.Driver)]
	if !ok {
		name = fmt.Sprintf("gina.GetGorm(%s)", conf.Driver)
	}
	console.Echo.Infof("✅ 提示: `%s` 模块加载成功, 你可以使用 `%s` 进行ORM操作\n", conf.Driver, name)
}

// 切换默认 Logger 使用的 Writer
func getLogger(conf *dbConfig) logger.Interface {
	logLevel := conf.LogLevel
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

	enableWriter := conf.EnableLogWriter
	return logger.New(getLogWriter(enableWriter), logger.Config{
		SlowThreshold:             time.Duration(conf.SlowThreshold) * time.Millisecond,
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
func ensureTimeout(dsn, defaultTimeout string) string {
	if strings.Contains(dsn, "timeout=") {
		return dsn
	}

	if strings.Contains(dsn, "?") {
		return dsn + "&timeout=" + defaultTimeout
	}

	return dsn + "?timeout=" + defaultTimeout
}

func GormPaginate(page, pageSize int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 1000:
			pageSize = 1000
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
