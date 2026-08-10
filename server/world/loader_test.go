package world

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
)

// TestLoaderViewChunkReentrant verifies that viewChunk returns when the Viewer it calls reads a block from a chunk
// that has a background load in flight. Completing that load runs the request's callbacks inline on the same
// goroutine, and the callback a Loader registers is viewChunk itself, so holding the Loader's lock across the Viewer
// deadlocks the world's transaction goroutine against itself.
//
// A real session reaches the same read through Player.Breathing, which reads the liquid at the player's eye when the
// player is shown to a viewer. That position need not be in the chunk being viewed.
func TestLoaderViewChunkReentrant(t *testing.T) {
	tests := []struct {
		name string
		// probeFromChunk reads the block from ViewChunk rather than from ViewEntity, covering the other call-out the
		// Loader used to make while holding its lock.
		probeFromChunk bool
	}{
		{name: "viewer reads a block while an entity is shown to it"},
		{name: "viewer reads a block while a chunk is shown to it", probeFromChunk: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The World must not be synchronous: loadChunkAsync never registers a request on one, so nothing can be
			// in flight to re-enter the Loader.
			w := newLoaderTestWorld(t)
			// A block position inside the chunk that is left to load in the background.
			v := &probeViewer{probe: cube.Pos{20, 64, 4}, fromChunk: tt.probeFromChunk}
			l := NewLoader(8, w, v)
			viewed, background := ChunkPos{0, 0}, ChunkPos{1, 0}

			returned := make(chan struct{})
			w.Do(func(tx *Tx) {
				defer close(returned)
				v.tx = tx

				// The viewed chunk must be resident and hold an entity, so that adding the Viewer to it shows one.
				c := tx.chunk(viewed)
				tx.AddEntity(EntitySpawnOpts{Position: mgl64.Vec3{8, 64, 8}}.New(taskTestEntityType{}, taskTestEntityConfig{}))

				// Put a load of the background chunk in flight, registered by this Loader exactly as Load does.
				l.mu.Lock()
				l.pending[background] = struct{}{}
				l.mu.Unlock()
				if !w.loadChunkAsync(tx, background, func(tx *Tx, c *Column) { l.viewChunk(tx, background, c) }) {
					t.Error("loadChunkAsync() = false, want true: no background load was scheduled")
					return
				}

				l.viewChunk(tx, viewed, c)

				if !v.probed {
					t.Error("the Viewer never read a block: it was not called")
				}
			})

			select {
			case <-returned:
			case <-time.After(time.Second * 10):
				t.Fatal("viewChunk did not return: the Loader deadlocked against itself")
			}
		})
	}
}

// newLoaderTestWorld returns a World that discards its log output. It is closed on a separate goroutine when the test
// ends, because a World whose transaction goroutine is stuck can never complete a Close.
func newLoaderTestWorld(t *testing.T) *World {
	t.Helper()

	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	t.Cleanup(func() {
		closed := make(chan struct{})
		go func() {
			_ = w.Close()
			close(closed)
		}()
		select {
		case <-closed:
		case <-time.After(time.Second * 10):
			t.Log("world could not be closed: its transaction goroutine is stuck")
		}
	})
	return w
}

// probeViewer reads a block from the world when it is shown a chunk or an entity, as a session does when it encodes
// the metadata of a player being shown to it.
type probeViewer struct {
	NopViewer
	tx        *Tx
	probe     cube.Pos
	fromChunk bool
	probed    bool
}

func (v *probeViewer) ViewChunk(ChunkPos, Dimension, map[cube.Pos]Block, *chunk.Chunk) {
	if v.fromChunk {
		v.read()
	}
}

func (v *probeViewer) ViewEntity(Entity) {
	if !v.fromChunk {
		v.read()
	}
}

func (v *probeViewer) read() {
	v.probed = true
	_, _ = v.tx.Liquid(v.probe)
}
