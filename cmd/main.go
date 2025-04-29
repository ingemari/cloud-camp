package main

import (
	"cloud/internal/balancer"
	"cloud/internal/config"
	"cloud/internal/graceful"
	"cloud/internal/handler"
	"cloud/internal/logs"
	"cloud/internal/proxy"
	"cloud/internal/ratelimit"
	"cloud/internal/repository"
	"cloud/internal/server"
	"cloud/internal/service"
	"cloud/internal/storage"
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"net/http"
)

func main() {
	logger := logs.SetupLogger("DEBUG")
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Error("Ошибка загрузки конфига: " + err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// клиент redis
	rdb := storage.NewRedisClient()
	v, err := rdb.Ping(ctx).Result()
	logger.Info("Database ping", "result", v)
	if err != nil {
		logger.Error("Database init error!")
	}
	defer rdb.Close()

	// создаем ограничитель запросов
	rateLimiter := ratelimit.NewRateLimiter(logger, rdb)
	// создаём балансировщик
	b := balancer.NewBalancer(cfg.Backends)

	// этот блок тоже нужно вынести в router
	// мультиплексор
	mux := http.NewServeMux()
	// эндпоинт
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleProxyRequest(cfg, w, r, logger, rateLimiter, b, cfg.Backends)
	})
	// repo
	userRepo := repository.NewUserRepository(rdb, logger)
	// service
	userService := service.NewUserService(userRepo, logger)
	// handler
	userHandler := handler.NewAuthHandler(userService, logger)
	// endpoints
	// ручки с добавлением уникальных лимитов в редис а так же их удаления
	mux.HandleFunc("/clients", userHandler.HandleAddUser)
	mux.HandleFunc("/clients/delete", userHandler.HandleDeleteUser)

	var wg sync.WaitGroup

	// ловим сигнал завершения
	go graceful.WaitForShutdown(cancel, logger)

	for _, port := range cfg.Backends {
		port = strings.Split(port, ":")[2]
		wg.Add(1)
		go server.RunBackend(ctx, logger, port, &wg)
	}

	server.Run(logger, mux, cfg.Port, 30*time.Second)
}

// не хватило времени ее вынести в роутинг пакет
func handleProxyRequest(cfg *config.Config, w http.ResponseWriter, r *http.Request, logger *slog.Logger, rateLimiter *ratelimit.RateLimiter, b *balancer.Balancer, backends []string) {
	clientIP := strings.Split(r.RemoteAddr, ":")[0]

	if !rateLimiter.AllowRequest(cfg, clientIP) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		logger.Warn("Rate limit exceeded", "client", clientIP)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	for i := 0; i < len(backends); i++ {
		target, err := b.NextBackend(logger)
		if err != nil {
			logger.Warn("Недоступный бэкенд сервер", "address", backends[i])
			continue
		}

		proxyServer, err := proxy.NewReverseProxy(target)
		if err != nil {
			logger.Error("Ошибка создания прокси", "target", target, "error", err)
			continue
		}

		proxyServer.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
			logger.Error("Ошибка проксирования", "target", target, "error", e)
			http.Error(rw, "Проблема на бэкенде", http.StatusBadGateway)
		}

		proxyServer.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Нет доступных бэкендов", http.StatusServiceUnavailable)
	logger.Warn("Нет доступных бэкендов")
}
