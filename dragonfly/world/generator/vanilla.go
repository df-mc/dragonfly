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
	Smoothness, ForestSize, ChanceForTrees float64

	TerrainPerlin *perlin.Perlin
	TreesPerlin   *perlin.Perlin
}

func NewVanillaGenerator(seed int64, alpha, beta, smoothness, forestsize, chancefortrees float64) (v Vanilla) {
	v.TerrainPerlin = perlin.NewPerlin(alpha, beta, 2, seed)
	v.TreesPerlin = perlin.NewPerlin(alpha, beta, 2, seed/2)
	v.ForestSize = forestsize
	v.Smoothness = smoothness
	v.ChanceForTrees = chancefortrees

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
	v.GenerateTrees(pos, chunk)
}

func (v Vanilla) GrassLevel(x, z uint8, pos world.ChunkPos) uint8 {
	return uint8(52+(v.TerrainPerlin.Noise2D(((16*(float64(pos.X())))+float64(x))/v.Smoothness, ((16*(float64(pos.Z())))+float64(z))/v.Smoothness)*15)) + 3
}

func (v Vanilla) GenerateTrees(pos world.ChunkPos, chunk *chunk.Chunk) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chance := v.TreesPerlin.Noise2D((float64(pos.X())+float64(x))/v.ForestSize, (float64(pos.Z())+float64(z))/v.ForestSize)
			randomChance := rand.Float64()
			if chance < v.ChanceForTrees && randomChance < .1 {
				v.GenerateTree(x, v.GrassLevel(x, z, pos), z, chunk)
			}
		}
	}
}

func (v Vanilla) GenerateTree(x, y, z uint8, chunk *chunk.Chunk) {
	max := y + 3 + uint8(rand.Intn(3))
	for y < max {
		chunk.SetRuntimeID(x, y, z, 0, log)
		y++
	}
}
