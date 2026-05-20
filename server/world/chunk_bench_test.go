package world

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/go-gl/mathgl/mgl64"
)

type benchmarkProvider struct {
	NopProvider
	delay time.Duration
	loads atomic.Int64
}

func (p *benchmarkProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	p.loads.Add(1)
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	return nil, leveldb.ErrNotFound
}

type benchmarkGenerator struct {
	work  int
	calls atomic.Int64
}

func (g *benchmarkGenerator) GenerateChunk(pos ChunkPos, c *chunk.Chunk) {
	g.calls.Add(1)
	var v uint64 = uint64(pos[0])<<32 | uint64(uint32(pos[1]))
	for i := 0; i < g.work; i++ {
		v ^= v << 13
		v ^= v >> 7
		v ^= v << 17
	}
	if v == 0 {
		c.SetBlock(0, 0, 0, 0, 0)
	}
}

func (g *benchmarkGenerator) DefaultSpawn(Dimension) cube.Pos { return cube.Pos{} }

type benchmarkViewer struct {
	NopViewer
	chunks atomic.Int64
}

func (v *benchmarkViewer) ViewChunk(ChunkPos, Dimension, map[cube.Pos]Block, *chunk.Chunk) {
	v.chunks.Add(1)
}

func BenchmarkLoaderLoadSlowProviderLatency(b *testing.B) {
	provider := &benchmarkProvider{delay: time.Millisecond}
	w := Config{Provider: provider, Generator: &benchmarkGenerator{}}.New()
	defer w.Close()

	loader := NewLoader(2, w, &benchmarkViewer{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := mgl64.Vec3{float64(i * 64), 64, 0}
		<-w.Exec(func(tx *Tx) {
			loader.Move(tx, pos)
			loader.Load(tx, 1)
		})
	}
}

func BenchmarkLoaderChunkVisibilityThroughput(b *testing.B) {
	provider := &benchmarkProvider{delay: time.Millisecond}
	viewer := &benchmarkViewer{}
	w := Config{Provider: provider, Generator: &benchmarkGenerator{}}.New()
	defer w.Close()

	loader := NewLoader(16, w, viewer)
	b.ReportAllocs()
	b.ResetTimer()
	for viewer.chunks.Load() < int64(b.N) {
		<-w.Exec(func(tx *Tx) {
			loader.Load(tx, 8)
		})
	}
}

func BenchmarkWorldExecLatencyDuringChunkLoad(b *testing.B) {
	provider := &benchmarkProvider{delay: time.Millisecond}
	w := Config{Provider: provider, Generator: &benchmarkGenerator{}}.New()
	defer w.Close()

	loader := NewLoader(2, w, &benchmarkViewer{})
	<-w.Exec(func(tx *Tx) {
		loader.Load(tx, 1)
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-w.Exec(func(tx *Tx) {})
	}
}
