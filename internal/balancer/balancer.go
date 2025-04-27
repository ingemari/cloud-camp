//2. Балансировщик balancer/balancer.go
//Структура Balancer, где будет:
//
//список серверов,
//
//индекс текущего сервера (round-robin),
//
//мьютекс для потокобезопасности.
//
//Метод NextBackend() string — выбираем следующий сервер.

package balancer

import (
	"sync"
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

func (b *Balancer) NextBackend() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backends) == 0 {
		return ""
	}

	backend := b.backends[b.current]
	b.current = (b.current + 1) % len(b.backends)
	return backend
}
