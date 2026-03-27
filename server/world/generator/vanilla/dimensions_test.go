package vanilla

import (
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func TestGenerateNetherChunksContainTerrainBiomesAndFeatures(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.Nether)
	positions := []world.ChunkPos{
		{0, 0},
		{8, 0},
		{0, 8},
		{16, 16},
		{-16, 24},
		{32, -32},
		{64, 32},
	}

	var (
		totalNetherrack int
		totalLava       int
		totalOres       int
		totalFeatures   int
		seenBiomes      = map[gen.Biome]struct{}{}
	)

	for _, pos := range positions {
		c := chunk.New(g.airRID, world.Nether.Range())
		g.GenerateChunk(pos, c)

		for y := c.Range().Min(); y <= c.Range().Max(); y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
					if !ok {
						continue
					}
					switch b.(type) {
					case block.Netherrack:
						totalNetherrack++
					case block.Lava:
						totalLava++
					}
					name, _ := b.EncodeBlock()
					name = strings.TrimPrefix(name, "minecraft:")
					switch {
					case strings.HasSuffix(name, "_ore") || name == "ancient_debris":
						totalOres++
					case name == "glowstone" || name == "magma" || strings.Contains(name, "fungus") || strings.Contains(name, "roots") || strings.Contains(name, "vines") || strings.Contains(name, "nylium") || name == "basalt" || name == "blackstone" || name == "soul_sand" || name == "soul_soil":
						totalFeatures++
					}
				}
			}
		}

		for localX := uint8(0); localX < 16; localX += 4 {
			for localZ := uint8(0); localZ < 16; localZ += 4 {
				for y := c.Range().Min(); y <= c.Range().Max(); y += 16 {
					seenBiomes[g.biomeSource.GetBiome(int(pos[0])*16+int(localX), y, int(pos[1])*16+int(localZ))] = struct{}{}
				}
			}
		}
	}

	t.Logf("nether totals: netherrack=%d lava=%d ores=%d features=%d biomes=%d", totalNetherrack, totalLava, totalOres, totalFeatures, len(seenBiomes))

	if totalNetherrack == 0 {
		t.Fatal("expected generated Nether chunks to contain netherrack terrain")
	}
	if totalLava == 0 {
		t.Fatal("expected generated Nether chunks to contain lava")
	}
	if totalOres == 0 {
		t.Fatal("expected generated Nether chunks to contain Nether ore features")
	}
	if totalFeatures == 0 {
		t.Fatal("expected generated Nether chunks to contain Nether surface/vegetation features")
	}
	if len(seenBiomes) < 2 {
		t.Fatalf("expected generated Nether chunks to contain multiple biomes, got %d", len(seenBiomes))
	}
}

func TestGenerateEndChunksContainTerrainBiomesAndFeatures(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.End)
	positions := []world.ChunkPos{
		{0, 0},
		{2, 0},
		{0, 2},
		{80, 80},
		{96, 96},
		{112, 80},
		{-80, 80},
		{-96, 96},
	}

	var (
		totalEndStone int
		totalObsidian int
		totalChorus   int
		seenBiomes    = map[gen.Biome]struct{}{}
	)

	for _, pos := range positions {
		c := chunk.New(g.airRID, world.End.Range())
		g.GenerateChunk(pos, c)

		for y := c.Range().Min(); y <= c.Range().Max(); y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
					if !ok {
						continue
					}
					switch b.(type) {
					case block.EndStone:
						totalEndStone++
					case block.Obsidian:
						totalObsidian++
					case block.ChorusPlant, block.ChorusFlower:
						totalChorus++
					}
				}
			}
		}

		for localX := uint8(0); localX < 16; localX += 4 {
			for localZ := uint8(0); localZ < 16; localZ += 4 {
				for y := c.Range().Min(); y <= c.Range().Max(); y += 32 {
					seenBiomes[g.biomeSource.GetBiome(int(pos[0])*16+int(localX), y, int(pos[1])*16+int(localZ))] = struct{}{}
				}
			}
		}
	}

	t.Logf("end totals: end_stone=%d obsidian=%d chorus=%d biomes=%d", totalEndStone, totalObsidian, totalChorus, len(seenBiomes))

	if totalEndStone == 0 {
		t.Fatal("expected generated End chunks to contain end stone terrain")
	}
	if totalObsidian == 0 {
		t.Fatal("expected generated End chunks to contain obsidian spike/platform blocks")
	}
	if totalChorus == 0 {
		t.Fatal("expected generated End chunks to contain chorus features")
	}
	if len(seenBiomes) < 2 {
		t.Fatalf("expected generated End chunks to contain multiple biomes, got %d", len(seenBiomes))
	}
}

func TestGenerateEndMainIslandContainsInactivePodium(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.End)
	c := chunk.New(g.airRID, world.End.Range())
	g.GenerateChunk(world.ChunkPos{0, 0}, c)

	podiumY := clamp(g.preliminarySurfaceLevelAt(0, 0, c.Range().Min(), c.Range().Max()), c.Range().Min()+1, c.Range().Max()-endPodiumPillarHeight)

	pillar, ok := world.BlockByRuntimeID(c.Block(0, int16(podiumY), 0, 0))
	if !ok {
		t.Fatal("expected center podium block runtime ID to decode")
	}
	if _, ok := pillar.(block.Bedrock); !ok {
		t.Fatalf("expected center of End podium to be bedrock, got %T", pillar)
	}

	ring, ok := world.BlockByRuntimeID(c.Block(3, int16(podiumY), 0, 0))
	if !ok {
		t.Fatal("expected outer podium ring runtime ID to decode")
	}
	if _, ok := ring.(block.Bedrock); !ok {
		t.Fatalf("expected outer End podium ring to be bedrock, got %T", ring)
	}

	support, ok := world.BlockByRuntimeID(c.Block(3, int16(podiumY-1), 0, 0))
	if !ok {
		t.Fatal("expected podium support runtime ID to decode")
	}
	if _, ok := support.(block.EndStone); !ok {
		t.Fatalf("expected podium support to be end stone, got %T", support)
	}
}
