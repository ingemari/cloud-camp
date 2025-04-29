package service

import (
	"cloud/internal/model"
	"context"
	"log/slog"
)

type UserRepository interface {
	AddUser(ctx context.Context, user model.User) error
	DelUser(ctx context.Context, user model.User) error
}

type UserService struct {
	userRepo UserRepository
	logger   *slog.Logger
}

func NewUserService(r UserRepository, logger *slog.Logger) *UserService {
	return &UserService{userRepo: r, logger: logger}
}

func (s *UserService) CreateUser(ctx context.Context, user model.User) error {
	return s.userRepo.AddUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, user model.User) error {
	return s.userRepo.DelUser(ctx, user)
}
