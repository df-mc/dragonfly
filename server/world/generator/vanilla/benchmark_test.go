package vanilla

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

func BenchmarkGenerateChunkFresh(b *testing.B) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	benchGeneratorFreshChunks(b, world.Overworld, func(g Generator, pos world.ChunkPos) {
		c := chunk.New(g.airRID, cube.Range{-64, 319})
		g.GenerateChunk(pos, c)
	})
}

func BenchmarkGenerateColumnFresh(b *testing.B) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	benchGeneratorFreshChunks(b, world.Overworld, func(g Generator, pos world.ChunkPos) {
		col := &chunk.Column{Chunk: chunk.New(g.airRID, cube.Range{-64, 319})}
		g.GenerateColumn(pos, col)
	})
}

func benchGeneratorFreshChunks(b *testing.B, dim world.Dimension, fn func(Generator, world.ChunkPos)) {
	g := NewForDimension(0, dim)
	positions := benchChunkPositions(max(b.N, 256))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(g, positions[i])
	}
}

func benchChunkPositions(n int) []world.ChunkPos {
	positions := make([]world.ChunkPos, n)
	baseX := int32(32768)
	baseZ := int32(-32768)
	for i := 0; i < n; i++ {
		// Walk a wide diagonal band so each iteration hits a distinct fresh chunk far from spawn.
		positions[i] = world.ChunkPos{
			baseX + int32(i*3),
			baseZ + int32(i*5),
		}
	}
	return positions
}
