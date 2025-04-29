package ratelimit

import (
	"cloud/internal/config"
	"cloud/internal/storage"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type Bucket struct {
	capacity     int           // Максимальное количество токенов
	tokens       int           // Текущее количество токенов
	refillRate   int           // Количество токенов, добавляемых за интервал (для refill)
	refillPeriod time.Duration // Период пополнения (для refill)
	lastRefill   time.Time     // Время последнего пополнения (для refill)
	mu           sync.Mutex    // Мьютекс бакета
}

type RateLimiter struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
	rdb     *redis.Client
	logger  *slog.Logger
}

func NewRateLimiter(logger *slog.Logger, rdb *redis.Client) *RateLimiter {
	r := &RateLimiter{
		buckets: make(map[string]*Bucket),
		rdb:     rdb,
		logger:  logger,
	}
	err := r.initFromRedis()
	if err != nil {
		r.logger.Error("Ошибка выгрузки юзеров из redis")
	}
	return r
}

func (rl *RateLimiter) AllowRequest(cfg *config.Config, clientID string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[clientID]
	if !exists {
		capacity, _ := strconv.Atoi(cfg.Capacity)
		rate, _ := strconv.Atoi(cfg.Rate)
		// Если клиента нет — автоматически создать ему bucket и добавить в redis
		bucket = &Bucket{
			capacity:     capacity,    // лимит по умолчанию из конфига
			tokens:       capacity,    // начальное количество токенов
			refillRate:   rate,        // сколько токенов пополняется из конфига
			refillPeriod: time.Second, // интервал пополнения
			lastRefill:   time.Now(),
		}
		rl.buckets[clientID] = bucket
	}
	rl.mu.Unlock()

	return bucket.takeToken()
}

func (b *Bucket) takeToken() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func (b *Bucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	if elapsed >= b.refillPeriod {
		newTokens := int(elapsed/b.refillPeriod) * b.refillRate
		if newTokens > 0 {
			b.tokens += newTokens
			if b.tokens > b.capacity {
				b.tokens = b.capacity
			}
			b.lastRefill = now
		}
	}
}

func (rl *RateLimiter) initFromRedis() error {
	users, err := storage.GetAllUserLimits(rl.rdb)
	if err != nil {
		return err
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, user := range users {
		rl.buckets[user.IP] = &Bucket{
			capacity:     user.Capacity,
			tokens:       user.Capacity,
			refillRate:   user.RatePerSec,
			refillPeriod: time.Second,
			lastRefill:   time.Now(),
		}
	}

	return nil
}
