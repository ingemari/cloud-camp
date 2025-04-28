package main

import (
	"cloud/internal/balancer"
	"cloud/internal/config"
	"cloud/internal/graceful"
	"cloud/internal/logs"
	"cloud/internal/proxy"
	"cloud/internal/ratelimit"
	"cloud/internal/server"
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

	// создаем ограничитель запросов
	rateLimiter := ratelimit.NewRateLimiter()

	// создаём балансировщик
	b := balancer.NewBalancer(cfg.Backends)

	// мультиплексор
	mux := http.NewServeMux()
	// эндпоинт
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleProxyRequest(w, r, logger, rateLimiter, b, cfg.Backends)
	})

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ловим сигнал завершения
	go graceful.WaitForShutdown(cancel, logger)

	for _, port := range cfg.Backends {
		port = strings.Split(port, ":")[2]
		wg.Add(1)
		go server.RunBackend(ctx, logger, port, &wg)
	}

	server.Run(logger, mux, cfg.Port, 30*time.Second)
}

func handleProxyRequest(w http.ResponseWriter, r *http.Request, logger *slog.Logger, rateLimiter *ratelimit.RateLimiter, b *balancer.Balancer, backends []string) {
	clientIP := strings.Split(r.RemoteAddr, ":")[0]

	if !rateLimiter.AllowRequest(clientIP) {
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
