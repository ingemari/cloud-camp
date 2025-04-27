package logs

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
)

func SetupLogger(logLevel string) *slog.Logger {
	var level slog.Level

	switch logLevel {
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	fileWriter := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    100, // MB
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// Мультиплексор (stdout + файл)
	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

	// Создаем текстовый обработчик с уровнем логирования
	handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level:     level, // Уровень логирования (Debug, Info, Error и wт.д.)
		AddSource: true,  // Добавляем информацию о месте в коде
	})

	// Создаем и возвращаем логгер
	return slog.New(handler)
}
