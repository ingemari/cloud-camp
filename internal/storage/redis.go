package storage

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
	"strings"
)

type UserLimit struct {
	IP         string
	Capacity   int
	RatePerSec int
}

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})
	return rdb
}

func GetAllUserLimits(rdb *redis.Client) ([]UserLimit, error) {
	var cursor uint64
	var users []UserLimit

	var ctx = context.Background()

	for {
		// SCAN ищет ключи user:*
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "user:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("redis scan error: %w", err)
		}

		for _, key := range keys {
			data, err := rdb.HGetAll(ctx, key).Result()
			if err != nil || len(data) == 0 {
				continue
			}

			capacity, _ := strconv.Atoi(data["capacity"])
			rate, _ := strconv.Atoi(data["rate_per_sec"])
			ip := strings.TrimPrefix(key, "user:")

			users = append(users, UserLimit{
				IP:         ip,
				Capacity:   capacity,
				RatePerSec: rate,
			})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return users, nil
}
