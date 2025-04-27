package main

import (
	"cloud/internal/balancer"
	"cloud/internal/config"
	"cloud/internal/graceful"
	"cloud/internal/logs"
	"cloud/internal/proxy"
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

	// Создаём балансировщик
	b := balancer.NewBalancer(cfg.Backends)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		maxAttempts := len(cfg.Backends)

		for i := 0; i < maxAttempts; i++ {
			target := b.NextBackend()
			if target == "" {
				http.Error(w, "Нет доступных бэкендов", http.StatusServiceUnavailable)
				return
			}

			proxyServer, err := proxy.NewReverseProxy(target)
			if err != nil {
				logger.Error("Ошибка создания прокси: ", "error", err.Error())
				continue // пробуем следующий сервер
			}

			done := make(chan bool, 1)

			proxyServer.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
				logger.Error("Ошибка проксирования на " + target + ": " + e.Error())
				done <- false
			}

			proxyServer.ServeHTTP(w, r)

			select {
			case success := <-done:
				if !success {
					// Ошибка на этом бэкенде — пробуем следующий
					continue
				}
			default:
				// Если всё прошло нормально — выход
				return
			}
		}

		http.Error(w, "Нет доступных бэкендов", http.StatusServiceUnavailable)
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

	//wg.Add(2)
	//go server.RunBackend(ctx, logger, "9001", &wg)
	//go server.RunBackend(ctx, logger, "9002", &wg)

	server.Run(logger, mux, cfg.Port, 30*time.Second)
}
