package gina

import (
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/modules/cachemodule"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	_ "github.com/soryetong/greasyx/tools/automatic"
)

var (
	odbMap sync.Map
	xdbMap sync.Map
	Rdb    redis.Cmdable
	Mdb    *mongo.Client
	Log    *ILog
	Casbin *casbin.SyncedEnforcer
	Cache  *cachemodule.Cache
)

func Run() {
	if err := console.RootCmd.Execute(); err != nil {
		console.Echo.Fatalf("❌ 错误: cmd run err: %s\n", err)
	}
}

// === 注册数据库实例 ===
const (
	DbTypeMysql      = "mysql"
	DbTypePostgresql = "postgresql"
	DbTypeSqlite     = "sqlite"
	DbTypeSqlserver  = "sqlserver"
	DbTypeOracle     = "oracle"
)

func SetGorm(driver string, db *gorm.DB) {
	odbMap.Store(strings.ToLower(driver), db)
}

func GetGorm(driver string) *gorm.DB {
	val, ok := odbMap.Load(strings.ToLower(driver))
	if !ok {
		return nil
	}
	return val.(*gorm.DB)
}

func SetSqlx(driver string, db *sqlx.DB) {
	xdbMap.Store(strings.ToLower(driver), db)
}

func GetSqlx(driver string) *sqlx.DB {
	val, ok := xdbMap.Load(strings.ToLower(driver))
	if !ok {
		return nil
	}
	return val.(*sqlx.DB)
}

// === （GORM） ===
func GMySQL() *gorm.DB     { return GetGorm(DbTypeMysql) }
func GPostgres() *gorm.DB  { return GetGorm(DbTypePostgresql) }
func GSqlite() *gorm.DB    { return GetGorm(DbTypeSqlite) }
func GSqlserver() *gorm.DB { return GetGorm(DbTypeSqlserver) }
func GOracle() *gorm.DB    { return GetGorm(DbTypeOracle) }

// === （sqlx） ===
func XMySQL() *sqlx.DB     { return GetSqlx(DbTypeMysql) }
func XPostgres() *sqlx.DB  { return GetSqlx(DbTypePostgresql) }
func XSqlite() *sqlx.DB    { return GetSqlx(DbTypeSqlite) }
func XSqlserver() *sqlx.DB { return GetSqlx(DbTypeSqlserver) }
func XOracle() *sqlx.DB    { return GetSqlx(DbTypeOracle) }
