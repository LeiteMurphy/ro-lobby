package ratelimit

import (
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newClock() *clock { return &clock{t: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)} }

// CA-05.2: até o limite passa; o seguinte é recusado com o tempo que falta.
func TestAllow_CA05_2_LimitThenRefuse(t *testing.T) {
	c := newClock()
	l := NewWithClock(30, time.Minute, c.now)
	for i := range 30 {
		if ok, _ := l.Allow("usuario-a"); !ok {
			t.Fatalf("requisição %d recusada antes do limite", i+1)
		}
	}
	c.t = c.t.Add(20 * time.Second)
	ok, retry := l.Allow("usuario-a")
	if ok {
		t.Fatal("a 31ª foi aceita")
	}
	if retry != 40*time.Second {
		t.Fatalf("retry = %s, esperado 40s", retry)
	}
}

// CA-05.2: a chave de outro usuário tem a própria contagem.
func TestAllow_CA05_2_OtherKeyUnaffected(t *testing.T) {
	l := NewWithClock(1, time.Minute, newClock().now)
	l.Allow("usuario-a")
	if ok, _ := l.Allow("usuario-a"); ok {
		t.Fatal("usuario-a passou do limite")
	}
	if ok, _ := l.Allow("usuario-b"); !ok {
		t.Fatal("usuario-b foi afetado pelo limite de usuario-a")
	}
}

// CA-05.4: depois da janela, a chave volta a ser aceita.
func TestAllow_CA05_4_WindowResets(t *testing.T) {
	c := newClock()
	l := NewWithClock(2, time.Minute, c.now)
	l.Allow("ip")
	l.Allow("ip")
	ok, retry := l.Allow("ip")
	if ok {
		t.Fatal("passou do limite")
	}
	c.t = c.t.Add(retry)
	if ok, _ := l.Allow("ip"); !ok {
		t.Fatal("continuou bloqueado depois do Retry-After")
	}
}

// RNF-02: as janelas vencidas são apagadas, e a memória não cresce sem fim.
func TestAllow_RNF02_SweepsExpired(t *testing.T) {
	c := newClock()
	l := NewWithClock(5, time.Minute, c.now)
	for _, k := range []string{"a", "b", "c"} {
		l.Allow(k)
	}
	c.t = c.t.Add(2 * time.Minute)
	l.Allow("d")
	if n := l.Len(); n != 1 {
		t.Fatalf("chaves em memória = %d, esperado 1", n)
	}
}
