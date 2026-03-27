package vanilla

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func TestCarveTerrainCarvesSolidChunks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	positions := []world.ChunkPos{
		{0, 0},
		{16, 16},
		{32, 0},
		{64, 32},
		{-48, 16},
	}

	carved := 0
	for _, pos := range positions {
		c := chunk.New(g.airRID, cube.Range{-64, 319})
		minY := c.Range().Min()
		maxY := c.Range().Max()
		chunkX := int(pos[0])
		chunkZ := int(pos[1])

		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				c.SetBlock(uint8(x), int16(minY), uint8(z), 0, g.bedrockRID)
				for y := minY + 1; y <= maxY; y++ {
					c.SetBlock(uint8(x), int16(y), uint8(z), 0, g.baseRuntimeID(y))
				}
			}
		}

		biomes := g.populateBiomeVolume(c, chunkX, chunkZ, minY, maxY)
		flat := g.graph.NewFlatCacheGrid(chunkX, chunkZ, g.noises)
		aquifer := gen.NewNoiseBasedAquifer(
			g.graph,
			chunkX,
			chunkZ,
			minY,
			maxY,
			g.noises,
			flat,
			g.seed,
			gen.OverworldFluidPicker{SeaLevel: seaLevel},
		)

		g.carveTerrain(c, biomes, chunkX, chunkZ, minY, maxY, aquifer)
		carved += countCarvedBlocks(c, g)
	}

	if carved == 0 {
		t.Fatal("expected carver pass to hollow at least one sampled solid chunk")
	}
}

func countCarvedBlocks(c *chunk.Chunk, g Generator) int {
	total := 0
	for y := c.Range().Min() + 1; y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				switch c.Block(uint8(x), int16(y), uint8(z), 0) {
				case g.airRID, g.waterRID, g.lavaRID:
					total++
				}
			}
		}
	}
	return total
}
