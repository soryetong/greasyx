package gina

import (
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"fmt"
	"os"
	"github.com/soryetong/greasyx/console"

	_ "github.com/soryetong/greasyx/tools/automatic"
)

var (
	Db  *gorm.DB
	Rdb redis.Cmdable
	Mdb *mongo.Client
	Log *ILog
)

func Run() {
	if err := console.RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] cmd run err: %s\n", err)
		os.Exit(104)
	}
}
