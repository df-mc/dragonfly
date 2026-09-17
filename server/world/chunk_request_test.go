package world

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// TestChunkCallbackDoesNotReenterLoader checks that a chunk callback reading a
// block in a chunk that still has a request pending does not re-enter
// Loader.viewChunk and deadlock the world owner.
func TestChunkCallbackDoesNotReenterLoader(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()

	nested := ChunkPos{1, 0}
	v := &reentrantViewer{pos: cube.Pos{int(nested[0]) * 16, 0, int(nested[1]) * 16}}
	l := NewLoader(1, w, v)

	h := EntitySpawnOpts{Position: mgl64.Vec3{0.5, 4, 0.5}}.New(reentrantEntityType{}, reentrantEntityConfig{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-w.exec(func(tx *Tx) {
			tx.AddEntity(h)
			w.loadChunkAsync(tx, nested, func(tx *Tx, col *Column) {
				l.viewChunk(tx, nested, col)
			})
			l.pending[nested] = struct{}{}
			l.Load(tx, 1)
		})
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		t.Fatal("chunk callback re-entered the Loader and deadlocked the world owner")
	}
	if !v.read {
		t.Fatal("viewer never read a block in the pending chunk, test did not reproduce the case")
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close world: %v", err)
	}
}

// reentrantViewer reads a block at pos the first time it is shown an entity.
type reentrantViewer struct {
	NopViewer
	pos  cube.Pos
	read bool
}

func (v *reentrantViewer) ViewEntity(e Entity) {
	if v.read {
		return
	}
	v.read = true
	e.(*reentrantEntity).tx.Block(v.pos)
}

type reentrantEntityConfig struct{}

func (reentrantEntityConfig) Apply(*EntityData) {}

type reentrantEntityType struct{}

func (reentrantEntityType) Open(tx *Tx, handle *EntityHandle, data *EntityData) Entity {
	return &reentrantEntity{tx: tx, handle: handle, data: data}
}

func (reentrantEntityType) EncodeEntity() string { return "dragonfly:reentrant_test_entity" }

func (reentrantEntityType) BBox(Entity) cube.BBox { return cube.Box(0, 0, 0, 1, 1, 1) }

func (reentrantEntityType) DecodeNBT(map[string]any, *EntityData) {}

func (reentrantEntityType) EncodeNBT(*EntityData) map[string]any { return nil }

type reentrantEntity struct {
	tx     *Tx
	handle *EntityHandle
	data   *EntityData
}

func (e *reentrantEntity) Close() error { return nil }

func (e *reentrantEntity) H() *EntityHandle { return e.handle }

func (e *reentrantEntity) Position() mgl64.Vec3 { return e.data.Pos }

func (e *reentrantEntity) Rotation() cube.Rotation { return e.data.Rot }
