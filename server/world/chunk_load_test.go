package world

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
)

type notFoundColumnProvider struct {
	NopProvider
}

func (notFoundColumnProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	return nil, leveldb.ErrNotFound
}

type loadErrorColumnProvider struct {
	NopProvider
	err error
}

func (p loadErrorColumnProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	return nil, p.err
}

type storingNotFoundColumnProvider struct {
	NopProvider
	stores int32
}

func (p *storingNotFoundColumnProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	return nil, leveldb.ErrNotFound
}

func (p *storingNotFoundColumnProvider) StoreColumn(ChunkPos, Dimension, *chunk.Column) error {
	atomic.AddInt32(&p.stores, 1)
	return nil
}

type notifyingNotFoundColumnProvider struct {
	NopProvider
	pos    ChunkPos
	loaded chan struct{}
}

func (p notifyingNotFoundColumnProvider) LoadColumn(pos ChunkPos, _ Dimension) (*chunk.Column, error) {
	if pos == p.pos {
		select {
		case p.loaded <- struct{}{}:
		default:
		}
	}
	return nil, leveldb.ErrNotFound
}

type blockingChunkGenerator struct {
	NopGenerator

	active  int32
	max     int32
	started chan struct{}
	release chan struct{}
}

func (g *blockingChunkGenerator) GenerateChunk(ChunkPos, *chunk.Chunk) {
	active := atomic.AddInt32(&g.active, 1)
	for {
		max := atomic.LoadInt32(&g.max)
		if active <= max || atomic.CompareAndSwapInt32(&g.max, max, active) {
			break
		}
	}
	g.started <- struct{}{}
	<-g.release
	atomic.AddInt32(&g.active, -1)
}

type countingChunkGenerator struct {
	NopGenerator
	calls int32
}

func (g *countingChunkGenerator) GenerateChunk(ChunkPos, *chunk.Chunk) {
	atomic.AddInt32(&g.calls, 1)
}

type positionalBlockingChunkGenerator struct {
	NopGenerator
	pos     ChunkPos
	started chan struct{}
	release chan struct{}
}

func (g positionalBlockingChunkGenerator) GenerateChunk(pos ChunkPos, _ *chunk.Chunk) {
	if pos != g.pos {
		return
	}
	select {
	case g.started <- struct{}{}:
	default:
	}
	<-g.release
}

type recordingViewer struct {
	NopViewer
	viewed chan ChunkPos
}

func (v recordingViewer) ViewChunk(pos ChunkPos, _ Dimension, _ map[cube.Pos]Block, _ *chunk.Chunk) {
	v.viewed <- pos
}

func TestChunkLoadWorkersGenerateConcurrently(t *testing.T) {
	gen := &blockingChunkGenerator{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	w := Config{Provider: notFoundColumnProvider{}, Generator: gen, ChunkLoadWorkers: 2}.New()
	defer w.Close()

	done := make(chan struct{}, 2)
	var queued bool
	<-w.exec(func(tx *Tx) {
		queued = w.loadChunkAsync(tx, ChunkPos{0, 0}, func(*Tx, *Column) { done <- struct{}{} }) &&
			w.loadChunkAsync(tx, ChunkPos{1, 0}, func(*Tx, *Column) { done <- struct{}{} })
	})
	if !queued {
		close(gen.release)
		t.Fatal("chunk load requests were not queued")
	}

	for i := 0; i < 2; i++ {
		select {
		case <-gen.started:
		case <-time.After(time.Second):
			close(gen.release)
			t.Fatal("timed out waiting for concurrent GenerateChunk calls")
		}
	}

	close(gen.release)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for generated chunks to be installed")
		}
	}
	if max := atomic.LoadInt32(&gen.max); max < 2 {
		t.Fatalf("GenerateChunk max concurrency = %d, want at least 2", max)
	}
}

func TestChunkLoadProviderErrorDoesNotGenerateOrInstall(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	gen := &countingChunkGenerator{}
	w := Config{
		Provider:         loadErrorColumnProvider{err: providerErr},
		Generator:        gen,
		ChunkLoadWorkers: 1,
	}.New()
	defer w.Close()

	col, err := w.loadChunk(ChunkPos{0, 0})
	if !errors.Is(err, providerErr) {
		t.Fatalf("loadChunk error = %v, want %v", err, providerErr)
	}
	if col != nil {
		t.Fatal("loadChunk returned a generated column after provider error")
	}
	if calls := atomic.LoadInt32(&gen.calls); calls != 0 {
		t.Fatalf("GenerateChunk calls after loadChunk provider error = %d, want 0", calls)
	}

	syncPos := ChunkPos{1, 0}
	<-w.exec(func(tx *Tx) {
		syncCol := tx.chunk(syncPos)
		if syncCol == nil || syncCol.Chunk == nil {
			t.Fatal("tx.chunk returned nil column after provider error")
		}
		if _, ok := w.chunks[syncPos]; ok {
			t.Fatal("tx.chunk installed a transient provider-error column")
		}
	})
	if calls := atomic.LoadInt32(&gen.calls); calls != 0 {
		t.Fatalf("GenerateChunk calls after tx.chunk provider error = %d, want 0", calls)
	}

	gotCallback := make(chan *Column, 1)
	asyncPos := ChunkPos{2, 0}
	<-w.exec(func(tx *Tx) {
		if !w.loadChunkAsync(tx, asyncPos, func(_ *Tx, col *Column) {
			gotCallback <- col
		}) {
			t.Fatal("chunk load request was not queued")
		}
	})

	select {
	case col := <-gotCallback:
		if col != nil {
			t.Fatal("async callback received a generated column after provider error")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async provider-error callback")
	}
	<-w.exec(func(*Tx) {
		if _, ok := w.chunks[asyncPos]; ok {
			t.Fatal("async provider-error callback installed a column")
		}
	})
	if calls := atomic.LoadInt32(&gen.calls); calls != 0 {
		t.Fatalf("GenerateChunk calls after async provider error = %d, want 0", calls)
	}
}

func TestSingleChunkLoadWorkerSerialisesGenerateChunk(t *testing.T) {
	gen := &blockingChunkGenerator{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	syncProviderLoaded := make(chan struct{}, 1)
	w := Config{
		Provider: notifyingNotFoundColumnProvider{
			pos:    ChunkPos{1, 0},
			loaded: syncProviderLoaded,
		},
		Generator:        gen,
		ChunkLoadWorkers: 1,
	}.New()
	defer w.Close()

	asyncDone := make(chan struct{}, 1)
	<-w.exec(func(tx *Tx) {
		if !w.loadChunkAsync(tx, ChunkPos{0, 0}, func(*Tx, *Column) { asyncDone <- struct{}{} }) {
			t.Fatal("async chunk load request was not queued")
		}
	})
	select {
	case <-gen.started:
	case <-time.After(time.Second):
		close(gen.release)
		t.Fatal("timed out waiting for first GenerateChunk call")
	}

	syncDone := make(chan struct{}, 1)
	go func() {
		_, _ = w.loadChunk(ChunkPos{1, 0})
		syncDone <- struct{}{}
	}()

	select {
	case <-syncProviderLoaded:
	case <-time.After(time.Second):
		close(gen.release)
		t.Fatal("timed out waiting for synchronous load to reach provider")
	}

	close(gen.release)
	select {
	case <-asyncDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async load to return")
	}
	select {
	case <-syncDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for synchronous loadChunk to return")
	}
	if max := atomic.LoadInt32(&gen.max); max != 1 {
		t.Fatalf("GenerateChunk max concurrency = %d, want 1", max)
	}
}

func TestSingleChunkLoadWorkerDoesNotDeadlockWhenSignalQueueIsFull(t *testing.T) {
	posA := ChunkPos{0, 0}
	posB := ChunkPos{1, 0}
	gen := positionalBlockingChunkGenerator{
		pos:     posA,
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	w := Config{Provider: notFoundColumnProvider{}, Generator: gen, ChunkLoadWorkers: 1}.New()

	<-w.exec(func(tx *Tx) {
		if !w.loadChunkAsync(tx, posA, func(*Tx, *Column) {}) {
			t.Fatal("first async chunk load request was not queued")
		}
		if !w.loadChunkAsync(tx, posB, func(*Tx, *Column) {}) {
			t.Fatal("second async chunk load request was not queued")
		}
	})
	select {
	case <-gen.started:
	case <-time.After(time.Second):
		close(gen.release)
		_ = w.Close()
		t.Fatal("timed out waiting for first GenerateChunk call")
	}

	txStarted := make(chan struct{})
	waitForB := make(chan struct{})
	txChunkReturned := make(chan struct{})
	go func() {
		<-w.exec(func(tx *Tx) {
			close(txStarted)
			<-waitForB
			tx.chunk(posB)
			close(txChunkReturned)
		})
	}()
	select {
	case <-txStarted:
	case <-time.After(time.Second):
		close(gen.release)
		_ = w.Close()
		t.Fatal("timed out waiting for blocking transaction to start")
	}

	for {
		select {
		case w.queue <- normalTransaction{c: make(chan struct{}), f: func(*Tx) {}}:
		default:
			goto queueFull
		}
	}

queueFull:
	close(waitForB)
	close(gen.release)

	select {
	case <-txChunkReturned:
	case <-time.After(time.Second):
		t.Fatal("tx.chunk waiting on the second async request deadlocked")
	}

	closed := make(chan struct{})
	go func() {
		_ = w.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not return after saturated signal queue cleared")
	}
}

func TestWorldCloseWithInFlightAsyncChunkLoadDoesNotInstall(t *testing.T) {
	provider := &storingNotFoundColumnProvider{}
	gen := &blockingChunkGenerator{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	viewer := recordingViewer{viewed: make(chan ChunkPos, 1)}
	w := Config{Provider: provider, Generator: gen, ChunkLoadWorkers: 1}.New()
	loader := NewLoader(1, w, viewer)

	<-w.exec(func(tx *Tx) {
		loader.Load(tx, 1)
	})
	select {
	case <-gen.started:
	case <-time.After(time.Second):
		close(gen.release)
		_ = w.Close()
		t.Fatal("timed out waiting for async GenerateChunk call")
	}

	closed := make(chan struct{})
	go func() {
		_ = w.Close()
		close(closed)
	}()

	select {
	case <-w.closing:
	case <-time.After(time.Second):
		close(gen.release)
		t.Fatal("timed out waiting for world close to begin")
	}
	close(gen.release)

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not return with an in-flight async chunk load")
	}
	select {
	case pos := <-viewer.viewed:
		t.Fatalf("viewer saw chunk %v after world close began", pos)
	default:
	}
	if stores := atomic.LoadInt32(&provider.stores); stores != 0 {
		t.Fatalf("provider stored %d chunks after closing, want 0", stores)
	}
	if _, ok := w.chunks[ChunkPos{0, 0}]; ok {
		t.Fatal("in-flight async chunk was installed after closing")
	}
}

// TestChunkDistanceLargeCoordinates ensures chunk distances do not overflow
// for positions far apart, such as after teleporting across the world.
func TestChunkDistanceLargeCoordinates(t *testing.T) {
	a, b := ChunkPos{1_875_000, 1_875_000}, ChunkPos{-1_875_000, -1_875_000}
	if d := chunkDistance(a, b); d != 5_303_301 {
		t.Fatalf("chunkDistance(%v, %v) = %d, want 5303301", a, b, d)
	}
}
