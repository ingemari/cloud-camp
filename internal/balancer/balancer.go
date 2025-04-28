package balancer

import (
	"errors"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"
)

type Balancer struct {
	backends []string
	current  int
	mu       sync.Mutex
}

func NewBalancer(backends []string) *Balancer {
	return &Balancer{
		backends: backends,
		current:  0,
	}
}

func (b *Balancer) NextBackend(logger *slog.Logger) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	total := len(b.backends)
	if total == 0 {
		return "", errors.New("нет доступных бэкендов")
	}

	for i := 0; i < total; i++ {
		backend := b.backends[b.current]
		b.current = (b.current + 1) % total

		if isBackendAlive(logger, backend) {
			return backend, nil
		}
	}

	return "", errors.New("все бэкенды недоступны")
}

func isBackendAlive(logger *slog.Logger, target string) bool {
	address := strings.TrimPrefix(target, "http://") // безопаснее
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err != nil {
		logger.Warn("Бэкенд недоступен", "backend", target)
		return false
	}
	_ = conn.Close()
	return true
}
