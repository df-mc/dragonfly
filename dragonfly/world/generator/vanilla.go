package generator

import (
	"github.com/aquilax/go-perlin"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/chunk"
)

var (
	stone, _ = world.BlockRuntimeID(block.Stone{})
)

type Vanilla struct {
	Seed int

	Perlin *perlin.Perlin
}

func NewVanillaGenerator(seed int64, alpha, beta float64) (v Vanilla) {
	v.Perlin = perlin.NewPerlin(alpha, beta, 2, seed)

	return
}

// GenerateChunk ...
func (v Vanilla) GenerateChunk(pos world.ChunkPos, chunk *chunk.Chunk) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetRuntimeID(x, 0, z, 0, bedrock)
			for y := uint8(1); y < uint8(54+(v.Perlin.Noise2D(float64(pos.X())+float64(x), float64(pos.Z())+float64(z))*15)); y++ {
				chunk.SetRuntimeID(x, y, z, 0, stone)
			}
		}
	}
}
