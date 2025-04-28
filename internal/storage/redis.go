package storage

import (
	"github.com/redis/go-redis/v9"
	"os"
)

func NewRedisStorage() {
	opt := &redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       os.Getenv("REDIS_DB"),
	}

	rdb := redis.NewClient(opt)
}
