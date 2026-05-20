package world

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

type blockingProvider struct {
	NopProvider
	started chan struct{}
	release chan struct{}
	loads   atomic.Int32
}

func newBlockingProvider() *blockingProvider {
	return &blockingProvider{started: make(chan struct{}), release: make(chan struct{})}
}

func (p *blockingProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	if p.loads.Add(1) == 1 {
		close(p.started)
	}
	<-p.release
	return &chunk.Column{Chunk: chunk.New(DefaultBlockRegistry, Overworld.Range())}, nil
}

type countingGenerator struct {
	calls atomic.Int32
}

func (g *countingGenerator) GenerateChunk(ChunkPos, *chunk.Chunk) {
	g.calls.Add(1)
}

func (g *countingGenerator) DefaultSpawn(Dimension) cube.Pos { return cube.Pos{} }

type recordingViewer struct {
	NopViewer
	chunks atomic.Int32
}

func (v *recordingViewer) ViewChunk(ChunkPos, Dimension, map[cube.Pos]Block, *chunk.Chunk) {
	v.chunks.Add(1)
}

func TestLoaderLoadRequestsMissingChunkAsynchronously(t *testing.T) {
	provider := newBlockingProvider()
	generator := &countingGenerator{}
	w := Config{Provider: provider, Generator: generator}.New()
	defer w.Close()

	viewer := &recordingViewer{}
	loader := NewLoader(2, w, viewer)

	done := make(chan struct{})
	w.Exec(func(tx *Tx) {
		loader.Load(tx, 1)
		close(done)
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Loader.Load blocked on chunk preparation")
	}
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("chunk preparation was not requested")
	}
	if viewer.chunks.Load() != 0 {
		t.Fatal("pending chunk was sent before preparation completed")
	}
	if loader.Loaded(ChunkPos{}) {
		t.Fatal("pending chunk was marked loaded")
	}

	close(provider.release)
	assertEventually(t, time.Second, func() bool {
		ch := w.Exec(func(tx *Tx) {
			loader.Load(tx, 1)
		})
		<-ch
		return viewer.chunks.Load() == 1 && loader.Loaded(ChunkPos{})
	})
}

func TestLoaderLoadDeduplicatesPendingRequests(t *testing.T) {
	provider := newBlockingProvider()
	generator := &countingGenerator{}
	w := Config{Provider: provider, Generator: generator}.New()
	defer w.Close()

	<-w.Exec(func(tx *Tx) {
		if !tx.World().requestChunk(ChunkPos{}) {
			t.Fatal("first chunk request was rejected")
		}
		if !tx.World().requestChunk(ChunkPos{}) {
			t.Fatal("duplicate chunk request was rejected")
		}
	})

	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("chunk preparation was not requested")
	}
	if got := provider.loads.Load(); got != 1 {
		t.Fatalf("expected one provider load, got %v", got)
	}
	close(provider.release)
}

func assertEventually(t *testing.T, timeout time.Duration, f func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	t.Fatal("condition was not met before timeout")
}
