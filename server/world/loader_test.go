package world

import (
	"io"
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"time"
)

// asyncWorld returns a World that ticks on its own goroutines and discards its log output.
func asyncWorld(conf Config) *World {
	conf.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	return conf.New()
}

// TestChangeWorldDoesNotBlockTheOwner verifies the shape of
// session.handleWorldSwitch: Loader.ChangeWorld runs on the *new* world's owner
// goroutine and does a blocking send on the *old* world's transaction queue
// (loader.go:52, l.w.exec). Two players swapping worlds in opposite directions
// while both queues are saturated makes the two owner goroutines block on each
// other's queue, freezing both worlds permanently.
func TestChangeWorldDoesNotBlockTheOwner(t *testing.T) {
	w1, w2 := asyncWorld(Config{}), asyncWorld(Config{})

	// a lives in w1 and is moved to w2 by a transaction on w2; b vice versa.
	a := NewLoader(2, w1, NopViewer{})
	b := NewLoader(2, w2, NopViewer{})

	park := func(w *World) chan struct{} {
		gate, started := make(chan struct{}), make(chan struct{})
		w.Do(func(tx *Tx) {
			close(started)
			<-gate
		})
		<-started
		return gate
	}
	gate1, gate2 := park(w1), park(w2)

	done := make(chan struct{}, 2)
	// Queue the world switches as the next item on each owner.
	w2.Do(func(tx *Tx) {
		a.ChangeWorld(tx, w2)
		done <- struct{}{}
	})
	w1.Do(func(tx *Tx) {
		b.ChangeWorld(tx, w1)
		done <- struct{}{}
	})

	// Saturate both owner queues: 128 buffered entries plus parked senders, so
	// a slot freed by an owner is immediately taken again.
	for range 400 {
		go w1.Save()
		go w2.Save()
	}
	time.Sleep(time.Millisecond * 300)

	close(gate1)
	close(gate2)

	deadline := time.After(time.Second * 10)
	for range 2 {
		select {
		case <-done:
		case <-deadline:
			buf := make([]byte, 1<<20)
			buf = buf[:runtime.Stack(buf, true)]
			for _, g := range strings.Split(string(buf), "\n\n") {
				if strings.Contains(g, "Loader).ChangeWorld") {
					t.Log("blocked world owner goroutine:\n" + g)
				}
			}
			t.Fatal("Loader.ChangeWorld blocked both world owner goroutines on each other's transaction queue")
		}
	}
	w1.Close()
	w2.Close()
}
