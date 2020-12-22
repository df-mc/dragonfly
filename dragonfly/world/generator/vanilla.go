package generator

import (
	"github.com/aquilax/go-perlin"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/chunk"
	"math/rand"
)

var (
	stone, _ = world.BlockRuntimeID(block.Stone{})
	log, _   = world.BlockRuntimeID(block.CoalBlock{})
)

type Vanilla struct {
	Seed                                   int64
	Smoothness, ForestSize, ChanceForTrees float64

	TerrainPerlin *perlin.Perlin

	sRand *rand.Rand
}

func NewVanillaGenerator(seed int64, alpha, beta, smoothness float64) (v Vanilla) {
	v.Seed = seed
	v.TerrainPerlin = perlin.NewPerlin(alpha, beta, 2, v.Seed)
	v.Smoothness = smoothness
	v.sRand = rand.New(rand.NewSource(v.Seed))
	return
}

// GenerateChunk ...
func (v Vanilla) GenerateChunk(pos world.ChunkPos, chunk *chunk.Chunk) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetRuntimeID(x, 0, z, 0, bedrock)
			max := uint8(52 + (v.TerrainPerlin.Noise2D(((16*(float64(pos.X())))+float64(x))/v.Smoothness, ((16*(float64(pos.Z())))+float64(z))/v.Smoothness) * 15))
			for y := uint8(1); y < max; y++ {
				chunk.SetRuntimeID(x, y, z, 0, stone)
			}

			dirtLevel := max
			for ; dirtLevel < max+2; dirtLevel++ {
				chunk.SetRuntimeID(x, dirtLevel, z, 0, dirt)
			}

			chunk.SetRuntimeID(x, dirtLevel, z, 0, grass)
		}
	}
}

func (v Vanilla) GrassLevel(x, z uint8, pos world.ChunkPos) uint8 {
	return uint8(52+(v.TerrainPerlin.Noise2D(((16*(float64(pos.X())))+float64(x))/v.Smoothness, ((16*(float64(pos.Z())))+float64(z))/v.Smoothness)*15)) + 2
}

func (v Vanilla) GenerateWater(pos world.ChunkPos, chunk *chunk.Chunk) {

}
