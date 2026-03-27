package world

import (
	"maps"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
)

func TestLoaderPopulateLoadQueuePrioritisesNearbyChunks(t *testing.T) {
	loader := &Loader{
		r:      2,
		pos:    ChunkPos{0, 0},
		loaded: map[ChunkPos]*Column{},
	}

	loader.populateLoadQueue()

	if len(loader.loadQueue) < 9 {
		t.Fatalf("expected non-trivial load queue, got %d entries", len(loader.loadQueue))
	}
	if loader.loadQueue[0] != (ChunkPos{0, 0}) {
		t.Fatalf("expected center chunk first, got %v", loader.loadQueue[0])
	}

	cardinals := map[ChunkPos]struct{}{
		{-1, 0}: {},
		{1, 0}:  {},
		{0, -1}: {},
		{0, 1}:  {},
	}
	for i := 1; i <= 4; i++ {
		if _, ok := cardinals[loader.loadQueue[i]]; !ok {
			t.Fatalf("expected immediate neighbor at index %d, got %v", i, loader.loadQueue[i])
		}
	}
	for _, diagonal := range []ChunkPos{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} {
		if indexOfChunkPos(loader.loadQueue, diagonal) < 5 {
			t.Fatalf("expected diagonal chunk %v after cardinals, queue=%v", diagonal, loader.loadQueue[:9])
		}
	}
}

func TestLoaderLoadPrefetchesMissingGeneratedChunkAsync(t *testing.T) {
	t.Parallel()

	generator := &blockingGenerator{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	w := Config{
		Provider:            NopProvider{},
		Generator:           generator,
		SaveInterval:        -1,
		ChunkUnloadInterval: time.Hour,
	}.New()
	defer func() {
		generator.unblock()
		_ = w.Close()
	}()

	loader := NewLoader(0, w, NopViewer{})
	target := ChunkPos{0, 0}

	done := w.Exec(func(tx *Tx) {
		loader.Move(tx, mgl64.Vec3{8, 64, 8})
		loader.Load(tx, 1)
	})
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("loader.Load blocked on missing chunk generation")
	}

	if _, ok := loader.Chunk(target); ok {
		t.Fatal("expected chunk miss to remain unloaded until prefetch installs it")
	}

	select {
	case <-generator.started:
	case <-time.After(time.Second):
		t.Fatal("expected async generator prefetch to start")
	}
	generator.unblock()

	if !waitForChunkInstall(w, target, time.Second) {
		t.Fatal("expected prefetched chunk to be installed")
	}

	<-w.Exec(func(tx *Tx) {
		loader.Load(tx, 1)
	})
	if _, ok := loader.Chunk(target); !ok {
		t.Fatal("expected loader to deliver prefetched chunk on retry")
	}
}

func TestLoaderLoadPrefetchesInitialPriorityWindow(t *testing.T) {
	t.Parallel()

	generator := &blockingGenerator{
		started: make(chan struct{}, 16),
		release: make(chan struct{}),
	}
	w := Config{
		Provider:            NopProvider{},
		Generator:           generator,
		SaveInterval:        -1,
		ChunkUnloadInterval: time.Hour,
	}.New()
	defer func() {
		generator.unblock()
		_ = w.Close()
	}()

	loader := NewLoader(2, w, NopViewer{})

	var inFlight map[ChunkPos]struct{}
	done := w.Exec(func(tx *Tx) {
		loader.Load(tx, 1)
		tx.World().prefetchMu.Lock()
		inFlight = maps.Clone(tx.World().prefetchInFlight)
		tx.World().prefetchMu.Unlock()
	})
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("loader.Load blocked while queueing nearby misses")
	}

	expected := []ChunkPos{
		{0, 0},
		{-1, 0}, {0, -1}, {0, 1}, {1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}
	if len(inFlight) != len(expected) {
		t.Fatalf("expected %d prefetched chunks, got %d (%v)", len(expected), len(inFlight), inFlight)
	}
	for _, pos := range expected {
		if _, ok := inFlight[pos]; !ok {
			t.Fatalf("expected chunk %v to be queued for initial prefetch, got %v", pos, inFlight)
		}
	}
}

func TestLoaderLoadKeepsPrefetchFrontierPinnedUntilInitialChunksInstall(t *testing.T) {
	t.Parallel()

	generator := &blockingGenerator{
		started: make(chan struct{}, 32),
		release: make(chan struct{}),
	}
	w := Config{
		Provider:            NopProvider{},
		Generator:           generator,
		SaveInterval:        -1,
		ChunkUnloadInterval: time.Hour,
	}.New()
	defer func() {
		generator.unblock()
		_ = w.Close()
	}()

	loader := NewLoader(4, w, NopViewer{})

	var inFlight map[ChunkPos]struct{}
	<-w.Exec(func(tx *Tx) {
		loader.Load(tx, 4)
		loader.Load(tx, 4)
		loader.Load(tx, 4)
		tx.World().prefetchMu.Lock()
		inFlight = maps.Clone(tx.World().prefetchInFlight)
		tx.World().prefetchMu.Unlock()
	})

	expected := []ChunkPos{
		{0, 0},
		{-1, 0}, {0, -1}, {0, 1}, {1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}
	if len(inFlight) != len(expected) {
		t.Fatalf("expected %d prefetched chunks after repeated loads, got %d (%v)", len(expected), len(inFlight), inFlight)
	}
	for _, pos := range expected {
		if _, ok := inFlight[pos]; !ok {
			t.Fatalf("expected chunk %v to stay in the initial prefetch frontier, got %v", pos, inFlight)
		}
	}
}

func TestWorldMetricsTrackAsyncPrefetchStages(t *testing.T) {
	t.Parallel()

	generator := &blockingGenerator{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	w := Config{
		Provider:            NopProvider{},
		Generator:           generator,
		SaveInterval:        -1,
		ChunkUnloadInterval: time.Hour,
	}.New()
	defer func() {
		generator.unblock()
		_ = w.Close()
	}()

	loader := NewLoader(0, w, NopViewer{})
	target := ChunkPos{0, 0}

	<-w.Exec(func(tx *Tx) {
		loader.Move(tx, mgl64.Vec3{8, 64, 8})
		loader.Load(tx, 1)
	})
	select {
	case <-generator.started:
	case <-time.After(time.Second):
		t.Fatal("expected async generator prefetch to start")
	}
	generator.unblock()

	if !waitForChunkInstall(w, target, time.Second) {
		t.Fatal("expected prefetched chunk to be installed")
	}

	snapshot := w.Metrics()
	if snapshot.ProviderMisses != 1 {
		t.Fatalf("expected 1 provider miss, got %d", snapshot.ProviderMisses)
	}
	if snapshot.Generation.Count != 1 {
		t.Fatalf("expected 1 generated chunk, got %d", snapshot.Generation.Count)
	}
	if snapshot.Lighting.Count != 1 {
		t.Fatalf("expected 1 lit chunk, got %d", snapshot.Lighting.Count)
	}
	if snapshot.Installation.Count != 1 {
		t.Fatalf("expected 1 installed chunk, got %d", snapshot.Installation.Count)
	}
	if snapshot.PrefetchQueuedTotal != 1 {
		t.Fatalf("expected 1 prefetched chunk request, got %d", snapshot.PrefetchQueuedTotal)
	}
}

func indexOfChunkPos(queue []ChunkPos, target ChunkPos) int {
	for i, pos := range queue {
		if pos == target {
			return i
		}
	}
	return -1
}

func waitForChunkInstall(w *World, pos ChunkPos, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var installed bool
		<-w.Exec(func(tx *Tx) {
			_, installed = tx.World().chunks[pos]
		})
		if installed {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

type blockingGenerator struct {
	started     chan struct{}
	release     chan struct{}
	releaseOnce sync.Once
	calls       atomic.Int32
}

func (g *blockingGenerator) GenerateChunk(pos ChunkPos, c *chunk.Chunk) {
	g.calls.Add(1)
	select {
	case g.started <- struct{}{}:
	default:
	}
	<-g.release
}

func (g *blockingGenerator) GenerateColumn(pos ChunkPos, col *chunk.Column) {
	g.GenerateChunk(pos, col.Chunk)
}

func (g *blockingGenerator) unblock() {
	g.releaseOnce.Do(func() {
		close(g.release)
	})
}
