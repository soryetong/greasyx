package dbmodule

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
)

var sm = make(map[string]string)

func initSqlx(conf *dbConfig) {
	if gina.GetSqlx(conf.Driver) != nil {
		return
	}
	if conf.MaxIdleConn == 0 {
		conf.MaxIdleConn = 10
	}
	if conf.MaxConn == 0 {
		conf.MaxConn = 200
	}

	dsn := ensureTimeout(conf.Dsn, "5s")
	driverArr := strings.Split(conf.Driver, "_")
	var db *sqlx.DB
	var err error
	switch strings.ToLower(driverArr[0]) {
	case gina.DbTypeMysql:
		db, err = sqlx.Open("mysql", dsn)
	case gina.DbTypePostgresql:
		db, err = sqlx.Open("postgres", dsn)
	case gina.DbTypeSqlite:
		db, err = sqlx.Open("sqlite3", dsn)
	case gina.DbTypeSqlserver:
		db, err = sqlx.Open("sqlserver", dsn)
	case gina.DbTypeOracle:
		db, err = sqlx.Open("oracle", dsn)
	default:
		console.Echo.Fatalf("❌ 错误: 不支持的数据库驱动类型: %s\n", conf.Driver)
	}

	if err != nil {
		console.Echo.Fatalf("❌ 错误: %s 数据库连接失败: %s\n", conf.Driver, err)
	}

	if err = db.Ping(); err != nil {
		console.Echo.Fatalf("❌ 错误: %s 数据库无法访问: %s\n", conf.Driver, err)
	}

	db.SetMaxIdleConns(conf.MaxIdleConn)
	db.SetMaxOpenConns(conf.MaxConn)
	gina.SetSqlx(conf.Driver, db)
	name, ok := sm[strings.ToLower(conf.Driver)]
	if !ok {
		name = fmt.Sprintf("gina.GetSqlx(%s)", conf.Driver)
	}
	console.Echo.Infof("✅ 提示: `%s` 模块加载成功, 你可以使用 `%s` 进行SQL操作\n", conf.Driver, name)
}
