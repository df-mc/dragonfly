package world

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

func TestCallRethrowsPanic(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	t.Cleanup(func() { _ = w.Close() })

	panicValue := &struct{ message string }{"call panic"}
	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("Call panic = %v, want original value %v", recovered, panicValue)
		}
	}()
	_, _ = Call(context.Background(), w, func(*Tx) (struct{}, error) {
		panic(panicValue)
	})
}

func TestCallRefRethrowsPanic(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	t.Cleanup(func() { _ = w.Close() })
	h := NewEntity(taskTestEntityType{}, taskTestEntityConfig{})
	if err := w.Do(func(tx *Tx) { tx.AddEntity(h) }).Wait(context.Background()); err != nil {
		t.Fatalf("add entity: %v", err)
	}

	panicValue := &struct{ message string }{"call ref panic"}
	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("CallRef panic = %v, want original value %v", recovered, panicValue)
		}
	}()
	_, _ = CallRef(context.Background(), NewEntityRef[Entity](h), func(*Tx, Entity) (struct{}, error) {
		panic(panicValue)
	})
}

func TestCallRethrowsPanicWhenCancellationLoses(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	t.Cleanup(func() { _ = w.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	release := make(chan struct{})
	panicValue := &struct{ message string }{"call panic after cancellation"}
	type outcome struct {
		panicValue any
		err        error
	}
	result := make(chan outcome, 1)
	go func() {
		var out outcome
		defer func() {
			out.panicValue = recover()
			result <- out
		}()
		_, out.err = Call(ctx, w, func(*Tx) (struct{}, error) {
			close(started)
			<-release
			panic(panicValue)
		})
	}()

	<-started
	cancel()
	select {
	case out := <-result:
		close(release)
		t.Fatalf("Call returned before running callback completed: panic = %v, err = %v", out.panicValue, out.err)
	case <-time.After(20 * time.Millisecond):
		close(release)
	}
	out := <-result
	if out.panicValue != panicValue {
		t.Fatalf("Call panic = %v, want original value %v (err = %v)", out.panicValue, panicValue, out.err)
	}
}

func TestDoCapturesPanic(t *testing.T) {
	w := Config{Synchronous: true, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	t.Cleanup(func() { _ = w.Close() })

	panicValue := &struct{ message string }{"do panic"}
	task := w.Do(func(*Tx) { panic(panicValue) })
	var panicErr *PanicError
	if err := task.Err(); !errors.As(err, &panicErr) {
		t.Fatalf("Do error = %v, want *PanicError", err)
	}
	if panicErr.Value != panicValue {
		t.Fatalf("PanicError.Value = %v, want original value %v", panicErr.Value, panicValue)
	}
}

func TestEntityDoCancelAfterInvalidatedWeakTransactionDoesNotPoisonHandle(t *testing.T) {
	w := New()
	defer w.Close()

	h := NewEntity(taskTestEntityType{}, taskTestEntityConfig{})
	<-w.exec(func(tx *Tx) { tx.AddEntity(h) })

	started := make(chan struct{})
	release := make(chan struct{})
	w.exec(func(*Tx) {
		close(started)
		<-release
	})
	<-started

	removeDone := w.exec(func(tx *Tx) {
		e, ok := h.Entity(tx)
		if !ok {
			t.Error("entity missing before remove")
			return
		}
		tx.RemoveEntity(e)
	})
	task := h.Do(func(*Tx, Entity) {
		t.Error("cancelled task ran")
	})
	// The fast path in schedule may queue directly without a weak
	// transaction. Either way, the task must still be cancellable while
	// pending.
	if !task.Cancel() {
		t.Fatal("expected pending task to cancel")
	}
	close(release)
	<-removeDone
	if err := task.Wait(testContext(t)); !errors.Is(err, ErrTaskCancelled) {
		t.Fatalf("expected ErrTaskCancelled, got %v", err)
	}

	<-w.exec(func(tx *Tx) { tx.AddEntity(h) })
	task = h.Do(func(*Tx, Entity) {})
	if err := task.Wait(testContext(t)); err != nil {
		t.Fatalf("handle poisoned after cancelling invalidated weak transaction: %v", err)
	}
}

func TestDoDoesNotBlockOwnerWhenQueueFull(t *testing.T) {
	w := New()
	defer w.Close()

	done := make(chan struct{})
	go func() {
		<-w.exec(func(tx *Tx) {
			for i := 0; i < cap(w.queue)+32; i++ {
				w.Do(func(*Tx) {})
				tx.Defer(func(*Tx) {})
			}
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Do or Context.Defer blocked the world owner")
	}
}

// TestWeakExecDoesNotBlockOwnerWhenQueueFull ensures an off-owner weak
// transaction blocked on a full queue does not hold scheduleMu, which would

func TestWeakExecDoesNotBlockOwnerWhenQueueFull(t *testing.T) {
	w := New()

	h := NewEntity(taskTestEntityType{}, taskTestEntityConfig{})
	<-w.exec(func(tx *Tx) { tx.AddEntity(h) })

	entered := make(chan struct{})
	proceed := make(chan struct{})
	release := make(chan struct{})

	// Occupy the world owner goroutine.
	w.exec(func(tx *Tx) {
		close(entered)
		<-proceed
		// Owner-side fire-and-forget must not block on scheduleMu.
		w.Do(func(tx *Tx) {})
		close(release)
	})
	<-entered

	// Fill the queue to capacity while the owner is busy.
	for i := 0; i < cap(w.queue); i++ {
		w.queue <- normalTransaction{c: make(chan struct{}), f: func(tx *Tx) {}}
	}

	// Entity task whose weak transaction ends up in World.weakExec with the
	// queue full.
	task := h.Do(func(tx *Tx, e Entity) {})
	time.Sleep(200 * time.Millisecond)

	close(proceed)
	select {
	case <-release:
	case <-time.After(3 * time.Second):
		t.Fatal("owner-side World.Do deadlocked on scheduleMu held by weakExec blocked on full queue")
	}
	if err := task.Wait(testContext(t)); err != nil {
		t.Fatalf("entity task failed after queue drained: %v", err)
	}
	_ = w.Close()
}

func TestDoQueuedBeforeCloseDoesNotRunAfterHandleClose(t *testing.T) {
	w := New()

	var closeHandled atomic.Bool
	var ranAfterClose atomic.Bool
	w.Handle(closeOrderHandler{closed: &closeHandled})

	started := make(chan struct{})
	release := make(chan struct{})
	w.exec(func(*Tx) {
		close(started)
		<-release
	})
	<-started

	for i := 0; i < cap(w.queue); i++ {
		w.exec(func(*Tx) {})
	}
	task := w.Do(func(*Tx) {
		if closeHandled.Load() {
			ranAfterClose.Store(true)
		}
	})

	closed := make(chan struct{})
	go func() {
		_ = w.Close()
		close(closed)
	}()
	close(release)

	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("world close did not complete")
	}
	if err := task.Wait(testContext(t)); err != nil && !errors.Is(err, ErrWorldClosed) {
		t.Fatalf("scheduled task failed with unexpected error: %v", err)
	}
	if ranAfterClose.Load() {
		t.Fatal("scheduled task ran after HandleClose")
	}
}

func TestEntityDoScheduledDuringWorldCloseRunsBeforeQueueShutdown(t *testing.T) {
	var task *Task
	w := New()
	h := NewEntity(closeSchedulingEntityType{}, closeSchedulingEntityConfig{onClose: func(h *EntityHandle) {
		task = h.Do(func(*Tx, Entity) {})
	}})
	<-w.exec(func(tx *Tx) { tx.AddEntity(h) })

	if err := w.Close(); err != nil {
		t.Fatalf("close world: %v", err)
	}
	if task == nil {
		t.Fatal("entity close did not schedule cleanup task")
	}
	if err := task.Wait(testContext(t)); err != nil {
		t.Fatalf("close-time entity task failed: %v", err)
	}
}

func TestEntityDoBlockedBeforeWorldCloseFailsPromptly(t *testing.T) {
	w := New()
	h := NewEntity(closeSchedulingEntityType{}, closeSchedulingEntityConfig{})
	<-w.exec(func(tx *Tx) { tx.AddEntity(h) })

	h.cond.L.Lock()
	h.weakTxActive = true
	h.cond.L.Unlock()
	task := h.Do(func(*Tx, Entity) {
		t.Error("entity task ran after world close")
	})

	if err := w.Close(); err != nil {
		t.Fatalf("close world: %v", err)
	}
	h.cond.L.Lock()
	h.weakTxActive = false
	h.cond.Broadcast()
	h.cond.L.Unlock()

	if err := task.Wait(testContext(t)); !errors.Is(err, ErrWorldClosed) {
		t.Fatalf("expected ErrWorldClosed, got %v", err)
	}
}

type taskTestEntityConfig struct{}

type closeOrderHandler struct {
	NopHandler
	closed *atomic.Bool
}

func (h closeOrderHandler) HandleClose(*Tx) { h.closed.Store(true) }

type closeSchedulingEntityConfig struct {
	onClose func(*EntityHandle)
}

func (c closeSchedulingEntityConfig) Apply(data *EntityData) { data.Data = c.onClose }

type closeSchedulingEntityType struct{}

func (closeSchedulingEntityType) Open(_ *Tx, handle *EntityHandle, data *EntityData) Entity {
	onClose, _ := data.Data.(func(*EntityHandle))
	return closeSchedulingEntity{h: handle, onClose: onClose}
}

func (closeSchedulingEntityType) EncodeEntity() string { return "dragonfly:close_scheduling_entity" }

func (closeSchedulingEntityType) BBox(Entity) cube.BBox { return cube.BBox{} }

func (closeSchedulingEntityType) DecodeNBT(map[string]any, *EntityData) {}

func (closeSchedulingEntityType) EncodeNBT(*EntityData) map[string]any { return nil }

type closeSchedulingEntity struct {
	h       *EntityHandle
	onClose func(*EntityHandle)
}

func (e closeSchedulingEntity) Close() error {
	if e.onClose != nil {
		e.onClose(e.h)
	}
	return nil
}

func (e closeSchedulingEntity) H() *EntityHandle { return e.h }

func (closeSchedulingEntity) Position() mgl64.Vec3 { return mgl64.Vec3{} }

func (closeSchedulingEntity) Rotation() cube.Rotation { return cube.Rotation{} }

func (taskTestEntityConfig) Apply(*EntityData) {}

type taskTestEntityType struct{}

func (taskTestEntityType) Open(tx *Tx, handle *EntityHandle, _ *EntityData) Entity {
	return taskTestEntity{h: handle, tx: tx}
}

func (taskTestEntityType) EncodeEntity() string { return "dragonfly:test_entity" }

func (taskTestEntityType) BBox(Entity) cube.BBox { return cube.BBox{} }

func (taskTestEntityType) DecodeNBT(map[string]any, *EntityData) {}

func (taskTestEntityType) EncodeNBT(*EntityData) map[string]any { return nil }

type taskTestEntity struct {
	h  *EntityHandle
	tx *Tx
}

func (e taskTestEntity) Close() error {
	if e.tx != nil {
		if ent, ok := e.h.Entity(e.tx); ok {
			e.tx.RemoveEntity(ent)
		}
	}
	return e.h.Close()
}

func (e taskTestEntity) H() *EntityHandle { return e.h }

func (taskTestEntity) Position() mgl64.Vec3 { return mgl64.Vec3{} }

func (taskTestEntity) Rotation() cube.Rotation { return cube.Rotation{} }

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
