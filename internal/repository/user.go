package repository

import (
	"cloud/internal/model"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type UserRepository struct {
	rdb    *redis.Client
	logger *slog.Logger
}

func NewUserRepository(rdb *redis.Client, logger *slog.Logger) *UserRepository {
	return &UserRepository{rdb: rdb, logger: logger}
}

func (r *UserRepository) AddUser(ctx context.Context, user model.User) error {
	key := "user:" + user.Id

	err := r.rdb.HSet(ctx, key, map[string]interface{}{
		"capacity":     user.Capacity,
		"rate_per_sec": user.RatePerSec,
	}).Err()
	if err != nil {
		r.logger.Error("Ошибка при сохранении пользователя в Redis", "ip", user.Id, "error", err)
		return fmt.Errorf("failed to add user: %w", err)
	}

	r.logger.Info("Пользователь сохранён в Redis", "ip", user.Id)
	return nil
}

func (r *UserRepository) DelUser(ctx context.Context, user model.User) error {
	key := "user:" + user.Id

	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Ошибка при удалении пользователя из Redis", "user_id", user.Id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.Info("Пользователь удалён из Redis", "user_id", user.Id)
	return nil
}
