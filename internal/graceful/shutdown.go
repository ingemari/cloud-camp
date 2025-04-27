package graceful

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(cancel context.CancelFunc, logger *slog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Получен сигнал завершения. Завершаем работу...")
	cancel()
}
