package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func RunBackend(ctx context.Context, logger *slog.Logger, port string, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Ответ от сервера", "port", port)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		logger.Info("Запуск сервера", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Ошибка запуска сервера", "port", port, "error", err)
		}
	}()

	<-ctx.Done() // ждём сигнал отмены контекста

	logger.Info("Shutting down server", "port", port)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Ошибка при остановке сервера", "port", port, "error", err)
	}
}

func Run(logger *slog.Logger, handler http.Handler, port string, shutdownTimeout time.Duration) {
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("Received shutdown signal", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	select {
	case <-ctx.Done():
		logger.Warn("Shutdown timeout exceeded")
	default:
		logger.Info("Server exited gracefully")
	}
}
