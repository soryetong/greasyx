package gina

import (
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"
)

var (
	Db  *gorm.DB
	Rdb redis.Cmdable
)
