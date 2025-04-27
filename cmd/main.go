//package cmd
//
//5. Главный файл cmd/main.go
//Загружаем конфиг.
//
//Создаём балансировщик.
//
//Запускаем HTTP-сервер на нужном порту.
//
//На каждый запрос вызываем балансировщик и проксируем его.

package main

import (
	"cloud/internal/balancer"
	"cloud/internal/config"
	"cloud/internal/logs"
	"cloud/internal/proxy"
	"cloud/internal/server"
	"log/slog"
	"strings"
	"sync"
	"time"

	"net/http"
)

func main() {
	logger := logs.SetupLogger("DEBUG")
	slog.SetDefault(logger)

	// Загружаем конфиг
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Error("Ошибка загрузки конфига: " + err.Error())
		return
	}

	// Создаём балансировщик
	b := balancer.NewBalancer(cfg.Backends)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := b.NextBackend()
		if target == "" {
			http.Error(w, "Нет доступных бэкендов", http.StatusServiceUnavailable)
			return
		}

		proxyServer, err := proxy.NewReverseProxy(target)
		if err != nil {
			http.Error(w, "Ошибка проксирования", http.StatusInternalServerError)
			logger.Error("Ошибка создания прокси: ", "error", err.Error())
			return
		}

		logger.Info("Проксируем на: " + target)
		proxyServer.ServeHTTP(w, r)
	})

	var wg sync.WaitGroup
	for _, port := range cfg.Backends {
		port = strings.Split(port, ":")[2]
		wg.Add(1)
		go server.RunBackend(logger, port, &wg)
	}

	server.Run(logger, mux, cfg.Port, 30*time.Second)

	wg.Wait()
}
