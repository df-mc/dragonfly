package world_test

import (
	"sync"
	"testing"
	"time"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/dragonfly/server/world/generator/vanilla"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/go-gl/mathgl/mgl64"
)

var finaliseBenchmarkBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()

func BenchmarkLoaderLoadFreshCenterChunk(b *testing.B) {
	benchmarkLoaderLoadFresh(b, 0, func(center world.ChunkPos) []world.ChunkPos {
		return []world.ChunkPos{center}
	})
}

func BenchmarkLoaderLoadFreshImmediateNeighbors(b *testing.B) {
	benchmarkLoaderLoadFresh(b, 1, immediateChunkPositions)
}

func BenchmarkLoaderLoadProviderHitCenterChunk(b *testing.B) {
	benchmarkLoaderLoadProviderHit(b, 0, func(center world.ChunkPos) []world.ChunkPos {
		return []world.ChunkPos{center}
	})
}

func BenchmarkLoaderLoadProviderHitImmediateNeighbors(b *testing.B) {
	benchmarkLoaderLoadProviderHit(b, 1, immediateChunkPositions)
}

func benchmarkLoaderLoadFresh(b *testing.B, radius int, required func(world.ChunkPos) []world.ChunkPos) {
	finaliseBenchmarkBlocksOnce.Do(worldFinaliseBlockRegistry)

	b.StopTimer()
	generator := vanilla.New(0)
	positions := benchmarkChunkPositions(max(b.N, 32))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		target := positions[i]
		w := world.Config{
			Provider:            world.NopProvider{},
			Generator:           generator,
			SaveInterval:        -1,
			ChunkUnloadInterval: time.Hour,
			MetricsLogThreshold: -1,
		}.New()
		loader := world.NewLoader(radius, w, world.NopViewer{})
		targetPos := mgl64.Vec3{float64(target[0]*16 + 8), 96, float64(target[1]*16 + 8)}

		b.StartTimer()
		ok := waitForLoaderChunks(w, loader, targetPos, required(target), 5*time.Second)
		b.StopTimer()

		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
		_ = w.Close()

		if !ok {
			b.Fatalf("timed out loading chunks around %v", target)
		}
	}
}

func benchmarkLoaderLoadProviderHit(b *testing.B, radius int, required func(world.ChunkPos) []world.ChunkPos) {
	finaliseBenchmarkBlocksOnce.Do(worldFinaliseBlockRegistry)

	b.StopTimer()
	positions := benchmarkChunkPositions(max(b.N, 32))
	provider := newBenchmarkProvider(positions, required)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		target := positions[i]
		w := world.Config{
			Provider:            provider,
			Generator:           world.NopGenerator{},
			SaveInterval:        -1,
			ChunkUnloadInterval: time.Hour,
			MetricsLogThreshold: -1,
		}.New()
		loader := world.NewLoader(radius, w, world.NopViewer{})
		targetPos := mgl64.Vec3{float64(target[0]*16 + 8), 96, float64(target[1]*16 + 8)}

		b.StartTimer()
		ok := waitForLoaderChunks(w, loader, targetPos, required(target), 5*time.Second)
		b.StopTimer()

		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
		_ = w.Close()

		if !ok {
			b.Fatalf("timed out loading provider-backed chunks around %v", target)
		}
	}
}

func waitForLoaderChunks(w *world.World, loader *world.Loader, pos mgl64.Vec3, required []world.ChunkPos, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		<-w.Exec(func(tx *world.Tx) {
			loader.Move(tx, pos)
			loader.Load(tx, len(required))
		})
		if loaderHasChunks(loader, required) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func loaderHasChunks(loader *world.Loader, required []world.ChunkPos) bool {
	for _, pos := range required {
		if _, ok := loader.Chunk(pos); !ok {
			return false
		}
	}
	return true
}

func immediateChunkPositions(center world.ChunkPos) []world.ChunkPos {
	return []world.ChunkPos{
		center,
		{center[0] - 1, center[1]},
		{center[0], center[1] - 1},
		{center[0], center[1] + 1},
		{center[0] + 1, center[1]},
	}
}

func benchmarkChunkPositions(n int) []world.ChunkPos {
	positions := make([]world.ChunkPos, n)
	baseX := int32(32768)
	baseZ := int32(-32768)
	for i := 0; i < n; i++ {
		positions[i] = world.ChunkPos{
			baseX + int32(i*3),
			baseZ + int32(i*5),
		}
	}
	return positions
}

type benchmarkProvider struct {
	world.NopProvider
	columns map[world.ChunkPos]*chunk.Column
}

func newBenchmarkProvider(centres []world.ChunkPos, required func(world.ChunkPos) []world.ChunkPos) benchmarkProvider {
	generator := vanilla.New(0)
	columns := make(map[world.ChunkPos]*chunk.Column)
	for _, center := range centres {
		for _, pos := range required(center) {
			if _, ok := columns[pos]; ok {
				continue
			}
			col := &chunk.Column{Chunk: chunk.New(0, cube.Range{-64, 319})}
			generator.GenerateColumn(pos, col)
			columns[pos] = col
		}
	}
	return benchmarkProvider{columns: columns}
}

func (p benchmarkProvider) LoadColumn(pos world.ChunkPos, _ world.Dimension) (*chunk.Column, error) {
	col, ok := p.columns[pos]
	if !ok {
		return nil, leveldb.ErrNotFound
	}
	return col, nil
}
