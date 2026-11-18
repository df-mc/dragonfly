package server

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

func TestReservePlayer(t *testing.T) {
	id := uuid.New()
	srv := &Server{joining: make(map[uuid.UUID]struct{}), p: make(map[uuid.UUID]*onlinePlayer)}

	var accepted atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if srv.reservePlayer(id) {
				accepted.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := accepted.Load(); got != 1 {
		t.Fatalf("expected one reservation, got %v", got)
	}
	if srv.reservePlayer(id) {
		t.Fatal("reserved UUID was accepted twice")
	}

	srv.releasePlayer(id)
	if !srv.reservePlayer(id) {
		t.Fatal("UUID not available after a failed join")
	}

	player := &onlinePlayer{}
	srv.admitPlayer(id, player)
	if srv.reservePlayer(id) {
		t.Fatal("online UUID was reserved")
	}
	srv.pmu.RLock()
	_, pending := srv.joining[id]
	got := srv.p[id]
	srv.pmu.RUnlock()
	if pending || got != player {
		t.Fatal("admission did not transfer the reservation")
	}

	srv.pmu.Lock()
	delete(srv.p, id)
	srv.pmu.Unlock()
	if !srv.reservePlayer(id) {
		t.Fatal("UUID not available after disconnect")
	}
}

func TestConcurrentAdmission(t *testing.T) {
	id := uuid.New()
	srv := &Server{joining: make(map[uuid.UUID]struct{}), p: make(map[uuid.UUID]*onlinePlayer)}
	if !srv.reservePlayer(id) {
		t.Fatal("initial reservation failed")
	}

	var accepted atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if srv.reservePlayer(id) {
				accepted.Add(1)
			}
		}()
	}
	close(start)
	srv.admitPlayer(id, &onlinePlayer{})
	wg.Wait()
	if got := accepted.Load(); got != 0 {
		t.Fatalf("expected no reservations during admission, got %v", got)
	}
}
