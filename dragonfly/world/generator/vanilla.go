package generator

import (
	"github.com/aquilax/go-perlin"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/block/grass"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/chunk"
	"math/rand"
)

var (
	stone, _     = world.BlockRuntimeID(block.Stone{})
	log, _       = world.BlockRuntimeID(block.Log{Wood: wood.Oak(), Stripped: false, Axis: world.Y})
	tallGrass, _ = world.BlockRuntimeID(block.TallGrass{Type: grass.Tall()})
)

type Vanilla struct {
	Smoothness, ForestFrequency, ChanceForTrees float64

	TerrainPerlin *perlin.Perlin
	BiomePerlin   *perlin.Perlin
	TreeRand      *rand.Rand
}

const (
	HEIGHT = 15
)

func NewVanillaGenerator(seed int64, alpha, beta, smoothness, forestSize, chanceForTree float64) (v Vanilla) {
	v.TerrainPerlin = perlin.NewPerlin(alpha, beta, 2, seed)
	v.Smoothness = smoothness
	v.BiomePerlin = perlin.NewPerlin(2, 2, 2, seed/2)
	v.TreeRand = rand.New(rand.NewSource(seed / 3))
	v.ForestFrequency = forestSize
	v.ChanceForTrees = chanceForTree
	return
}

// GenerateChunk ...
func (v Vanilla) GenerateChunk(pos world.ChunkPos, chunk *chunk.Chunk) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetRuntimeID(x, 0, z, 0, bedrock)
			max := uint8(52 + (v.Perlin2DAt(x, z, v.Smoothness, pos, v.TerrainPerlin) * HEIGHT))
			for y := uint8(1); y < max; y++ {
				chunk.SetRuntimeID(x, y, z, 0, stone)
			}

			dirtLevel := max
			for ; dirtLevel < max+2; dirtLevel++ {
				chunk.SetRuntimeID(x, dirtLevel, z, 0, dirt)
			}

			chunk.SetRuntimeID(x, dirtLevel, z, 0, grassBlock)
		}
	}
	chance := v.BiomePerlin.Noise2D(float64(pos.X())/5, float64(pos.Z())/5)

	if chance < v.ForestFrequency {
		v.GenerateTrees(chunk, pos)
	}
}

func (v Vanilla) GrassLevel(x, z uint8, pos world.ChunkPos) uint8 {
	return uint8(52 + (v.Perlin2DAt(x, z, v.Smoothness, pos, v.TerrainPerlin) * HEIGHT) + 2)
}

func (v Vanilla) GenerateTrees(chunk *chunk.Chunk, pos world.ChunkPos) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			h := v.TreeRand.Float64()
			//TODO: Also take into account if there's another tree growing nearby
			if h < v.ChanceForTrees {
				v.GenerateTree(x, z, chunk, v.GrassLevel(x, z, pos))
			}
		}
	}
}

func (v Vanilla) GenerateTree(x, z uint8, chunk *chunk.Chunk, grassLevel uint8) {
	max := 3 + uint8(rand.Intn(3)) + grassLevel

	for i := grassLevel + 1; i < max; i++ {
		chunk.SetRuntimeID(x, i, z, 0, log)
	}
}

func (v Vanilla) Perlin2DAt(x, z uint8, smoothness float64, pos world.ChunkPos, perlin *perlin.Perlin) float64 {
	// We add .4 because the range is -.4 to .4
	return perlin.Noise2D((float64(pos.X()*16)+float64(x))/smoothness, (float64(pos.Z()*16)+float64(z))/smoothness) + .4
}
