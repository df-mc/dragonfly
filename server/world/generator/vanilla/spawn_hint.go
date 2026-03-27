package vanilla

import (
	"math"

	"github.com/df-mc/dragonfly/server/world"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

// FindSpawnChunk returns a nearby overworld chunk whose surface biome is suitable for a visible land spawn.
func (g Generator) FindSpawnChunk(maxRadius int) (world.ChunkPos, bool) {
	best := world.ChunkPos{}
	bestScore := -1
	bestDist := math.MaxFloat64

	for radius := 0; radius <= maxRadius; radius++ {
		for _, pos := range spawnHintRing(radius) {
			biome := g.biomeSource.GetBiome(int(pos[0])*16+8, seaLevel, int(pos[1])*16+8)
			score := spawnHintBiomeScore(biome)
			if score <= 0 {
				continue
			}

			dist := math.Hypot(float64(pos[0]), float64(pos[1]))
			if score > bestScore || (score == bestScore && dist < bestDist) {
				best = pos
				bestScore = score
				bestDist = dist
				if score >= 100 {
					return best, true
				}
			}
		}
	}
	return best, bestScore > 0
}

func spawnHintRing(radius int) []world.ChunkPos {
	if radius == 0 {
		return []world.ChunkPos{{}}
	}

	out := make([]world.ChunkPos, 0, radius*8)
	for x := -radius; x <= radius; x++ {
		out = append(out, world.ChunkPos{int32(x), int32(-radius)})
		out = append(out, world.ChunkPos{int32(x), int32(radius)})
	}
	for z := -radius + 1; z <= radius-1; z++ {
		out = append(out, world.ChunkPos{int32(-radius), int32(z)})
		out = append(out, world.ChunkPos{int32(radius), int32(z)})
	}
	return out
}

func spawnHintBiomeScore(biome gen.Biome) int {
	switch biome {
	case gen.BiomeForest, gen.BiomeBirchForest, gen.BiomeDarkForest, gen.BiomeFlowerForest, gen.BiomeTaiga, gen.BiomeSnowyTaiga, gen.BiomeOldGrowthPineTaiga, gen.BiomeOldGrowthSpruceTaiga, gen.BiomeJungle, gen.BiomeSparseJungle, gen.BiomeBambooJungle, gen.BiomeMangroveSwamp, gen.BiomeCherryGrove:
		return 100
	case gen.BiomePlains, gen.BiomeSunflowerPlains, gen.BiomeSavanna, gen.BiomeSavannaPlateau, gen.BiomeWindsweptSavanna, gen.BiomeMeadow:
		return 70
	case gen.BiomeBeach, gen.BiomeSnowyBeach, gen.BiomeStonyShore, gen.BiomeRiver, gen.BiomeFrozenRiver:
		return 0
	case gen.BiomeOcean, gen.BiomeDeepOcean, gen.BiomeColdOcean, gen.BiomeDeepColdOcean, gen.BiomeFrozenOcean, gen.BiomeDeepFrozenOcean, gen.BiomeLukewarmOcean, gen.BiomeDeepLukewarmOcean, gen.BiomeWarmOcean, gen.BiomeDeepWarmOcean:
		return 0
	default:
		return 40
	}
}
