// Package ratelimit conta requisições por chave numa janela fixa, em memória (spec
// endurecimento, RN-15, RNF-02). Os contadores zeram quando a API reinicia, o que é aceito.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter permite até Limit requisições por chave a cada Window.
type Limiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu        sync.Mutex
	buckets   map[string]*bucket
	lastSweep time.Time
}

type bucket struct {
	count int
	reset time.Time
}

// New devolve um Limiter com o relógio real.
func New(limit int, window time.Duration) *Limiter {
	return NewWithClock(limit, window, time.Now)
}

// NewWithClock devolve um Limiter com o relógio dado, para os testes.
func NewWithClock(limit int, window time.Duration, now func() time.Time) *Limiter {
	return &Limiter{limit: limit, window: window, now: now, buckets: map[string]*bucket{}}
}

// Allow conta uma requisição da chave. Passado o limite, devolve false e quanto falta para
// a janela da chave terminar.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.sweep(now)
	b, ok := l.buckets[key]
	if !ok || !now.Before(b.reset) {
		b = &bucket{reset: now.Add(l.window)}
		l.buckets[key] = b
	}
	if b.count >= l.limit {
		return false, b.reset.Sub(now)
	}
	b.count++
	return true, 0
}

// sweep apaga as janelas vencidas uma vez por janela, para a memória não crescer com IPs
// que não voltam.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for key, b := range l.buckets {
		if !now.Before(b.reset) {
			delete(l.buckets, key)
		}
	}
}

// Len diz quantas chaves estão em memória (para os testes).
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
