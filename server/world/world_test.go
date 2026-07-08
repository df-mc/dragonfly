package world

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

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

func TestSynchronousExecWorldCanRemoveEntity(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	h := EntitySpawnOpts{Position: mgl64.Vec3{0, 4, 0}}.New(testEntityType{}, testEntityConfig{})
	<-w.Exec(func(tx *Tx) {
		tx.AddEntity(h)
	})

	done := make(chan struct{})
	go func() {
		h.ExecWorld(func(tx *Tx, e Entity) {
			tx.RemoveEntity(e)
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal("ExecWorld self-removal did not complete")
	}
}

func TestSynchronousAdvanceTickTicksViewerlessEntities(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	h := EntitySpawnOpts{Position: mgl64.Vec3{0, 4, 0}}.New(testEntityType{}, testEntityConfig{})
	<-w.Exec(func(tx *Tx) {
		tx.AddEntity(h)
	})

	start := h.data.Pos
	for range 3 {
		w.AdvanceTick()
	}
	if got := h.data.Pos; got == start {
		t.Fatalf("expected entity position to change after ticking, got %v", got)
	}
}

type testEntityConfig struct{}

func (testEntityConfig) Apply(*EntityData) {}

type testEntityType struct{}

func (testEntityType) Open(_ *Tx, handle *EntityHandle, data *EntityData) Entity {
	return &testEntity{handle: handle, data: data}
}

func (testEntityType) EncodeEntity() string {
	return "dragonfly:test_entity"
}

func (testEntityType) BBox(Entity) cube.BBox {
	return cube.Box(0, 0, 0, 1, 1, 1)
}

func (testEntityType) DecodeNBT(map[string]any, *EntityData) {}

func (testEntityType) EncodeNBT(*EntityData) map[string]any {
	return nil
}

type testEntity struct {
	handle *EntityHandle
	data   *EntityData
}

func (e *testEntity) Close() error {
	return nil
}

func (e *testEntity) H() *EntityHandle {
	return e.handle
}

func (e *testEntity) Position() mgl64.Vec3 {
	return e.data.Pos
}

func (e *testEntity) Rotation() cube.Rotation {
	return e.data.Rot
}

func (e *testEntity) Tick(*Tx, int64) {
	e.data.Pos = e.data.Pos.Add(mgl64.Vec3{0, -0.1, 0})
}
