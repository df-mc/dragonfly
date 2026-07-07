package world

import (
	"runtime"
	"testing"
	"time"
)

// TestSynchronousWorldNoGoroutines verifies that a synchronous World starts no
// background goroutines.
func TestSynchronousWorldNoGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()
	w := Config{Synchronous: true}.New()
	if after := runtime.NumGoroutine(); after != before {
		t.Errorf("expected no new goroutines after New(), had %v, got %v", before, after)
	}
	_ = w.Close()
}

// TestSynchronousWorldExec verifies that Exec on a synchronous World runs the
// transaction on the calling goroutine, with the returned channel closed once
// Exec returns.
func TestSynchronousWorldExec(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	var ran bool
	c := w.Exec(func(tx *Tx) { ran = true })
	if !ran {
		t.Fatal("expected transaction to have run when Exec returned")
	}
	select {
	case <-c:
	default:
		t.Fatal("expected channel returned by Exec to be closed when Exec returned")
	}
}

// TestSynchronousWorldAdvanceTick verifies that a synchronous World does not
// tick on its own and that AdvanceTick advances the current tick exactly once
// per call, even without any viewers.
func TestSynchronousWorldAdvanceTick(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	current := func() int64 {
		w.set.Lock()
		defer w.set.Unlock()
		return w.set.CurrentTick
	}
	start := current()
	time.Sleep(time.Second / 10)
	if got := current(); got != start {
		t.Fatalf("expected no automatic ticking, tick advanced from %v to %v", start, got)
	}
	for range 5 {
		w.AdvanceTick()
	}
	if got := current(); got != start+5 {
		t.Fatalf("expected current tick %v after 5 AdvanceTick calls, got %v", start+5, got)
	}
}

// TestSynchronousWorldClose verifies that closing a synchronous World returns
// promptly instead of waiting for goroutines that were never started.
func TestSynchronousWorldClose(t *testing.T) {
	w := Config{Synchronous: true}.New()

	done := make(chan struct{})
	go func() {
		_ = w.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal("Close did not return within 5 seconds")
	}
}
