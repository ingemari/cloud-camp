package handler

import (
	"cloud/internal/handler/dto"
	"cloud/internal/handler/mapper"
	"cloud/internal/model"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, user model.User) error
}

type AuthHandler struct {
	userService UserService
	logger      *slog.Logger
}

func NewAuthHandler(as UserService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{userService: as, logger: logger}
}

func (h *AuthHandler) HandleAddUser(w http.ResponseWriter, r *http.Request) {
	var req dto.UserAddReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		h.logger.Error("Некорректный запрос")
		return
	}
	// без валидации
	user := mapper.AddReqToUser(req)

	err := h.userService.CreateUser(r.Context(), user)
	if err != nil {
		h.logger.Error("Failed create user")
		http.Error(w, "Ошибка добавления юзера", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь добавлен",
	})
}

func (h *AuthHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req dto.UserDelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		h.logger.Error("Некорректный запрос")
		return
	}

	user := mapper.DelReqToUser(req)

	err := h.userService.DeleteUser(r.Context(), user)
	if err != nil {
		h.logger.Error("Failed delete user")
		http.Error(w, "Ошибка удаления юзера", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь удален",
	})
}
