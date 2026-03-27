package vanilla

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func TestPopulateBiomeVolumeStoresVerticalBiomeSplits(t *testing.T) {
	biomeSource, err := gen.NewBiomeSource(0, gen.NewWorldgenRegistry(), "overworld")
	if err != nil {
		t.Fatalf("create biome source: %v", err)
	}
	g := Generator{biomeSource: biomeSource}

	surfaceY := 80
	caveYs := []int{-64, -32, 0}
	split, ok := findBiomeSplit(g.biomeSource, surfaceY, caveYs)
	if !ok {
		t.Fatal("did not find a column with differing surface and underground biomes")
	}

	c := chunk.New(0, cube.Range{-64, 319})
	g.populateBiomeVolume(c, split.chunkX, split.chunkZ, c.Range().Min(), c.Range().Max())

	gotSurface := g.biomeAt(c, split.localX, surfaceY, split.localZ)
	if gotSurface != split.surfaceBiome {
		t.Fatalf("expected surface biome %v at (%d,%d,%d), got %v", split.surfaceBiome, split.worldX, surfaceY, split.worldZ, gotSurface)
	}

	gotCave := g.biomeAt(c, split.localX, split.caveY, split.localZ)
	if gotCave != split.caveBiome {
		t.Fatalf("expected underground biome %v at (%d,%d,%d), got %v", split.caveBiome, split.worldX, split.caveY, split.worldZ, gotCave)
	}
	if gotSurface == gotCave {
		t.Fatalf("expected differing surface and underground biomes at (%d,%d), got %v", split.worldX, split.worldZ, gotSurface)
	}
}

type biomeSplitSample struct {
	worldX       int
	worldZ       int
	chunkX       int
	chunkZ       int
	localX       int
	localZ       int
	caveY        int
	surfaceBiome gen.Biome
	caveBiome    gen.Biome
}

func findBiomeSplit(source gen.BiomeSource, surfaceY int, caveYs []int) (biomeSplitSample, bool) {
	for _, step := range []int{128, 64, 16, 4} {
		for worldX := -8192; worldX <= 8192; worldX += step {
			for worldZ := -8192; worldZ <= 8192; worldZ += step {
				surfaceBiome := source.GetBiome(worldX, surfaceY, worldZ)
				for _, caveY := range caveYs {
					caveBiome := source.GetBiome(worldX, caveY, worldZ)
					if caveBiome == surfaceBiome || !isUndergroundBiome(caveBiome) {
						continue
					}
					chunkX := alignDown(worldX, 16) / 16
					chunkZ := alignDown(worldZ, 16) / 16
					return biomeSplitSample{
						worldX:       worldX,
						worldZ:       worldZ,
						chunkX:       chunkX,
						chunkZ:       chunkZ,
						localX:       worldX - chunkX*16,
						localZ:       worldZ - chunkZ*16,
						caveY:        caveY,
						surfaceBiome: surfaceBiome,
						caveBiome:    caveBiome,
					}, true
				}
			}
		}
	}
	return biomeSplitSample{}, false
}

func isUndergroundBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeDripstoneCaves, gen.BiomeLushCaves, gen.BiomeDeepDark:
		return true
	default:
		return false
	}
}
