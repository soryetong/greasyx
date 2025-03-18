package gina

import (
	"github.com/casbin/casbin/v2"
	"github.com/go-redis/redis/v8"
	"github.com/soryetong/greasyx/console"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	_ "github.com/soryetong/greasyx/tools/automatic"
)

var (
	Db     *gorm.DB
	Rdb    redis.Cmdable
	Mdb    *mongo.Client
	Log    *ILog
	Casbin *casbin.SyncedEnforcer
)

func Run() {
	if err := console.RootCmd.Execute(); err != nil {
		console.Echo.Fatalf("❌ 错误: cmd run err: %s\n", err)
	}
}
