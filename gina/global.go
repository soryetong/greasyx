package gina

import (
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var (
	Db  *gorm.DB
	Rdb redis.Cmdable
	Mdb *mongo.Client
	Log *ILog
)
