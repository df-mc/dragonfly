package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/dragonfly/server/world/generator/vanilla"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const seaLevel = 63

type overworldSample struct {
	ChunkX           int32
	ChunkZ           int32
	Height           int
	Biome            gen.Biome
	Top              string
	Sea              string
	UndergroundAir   int
	UndergroundWater int
	UndergroundLava  int
	Vegetation       int
	Trees            int
	Ores             int
	Magma            int
	StructureBlocks  int
}

type overworldReport struct {
	Samples          []overworldSample
	MinHeight        int
	MaxHeight        int
	UndergroundAir   int
	UndergroundWater int
	UndergroundLava  int
	Vegetation       int
	Trees            int
	Ores             int
	Magma            int
	StructureBlocks  int
	UniqueBiomes     int
}

type netherReport struct {
	Netherrack   int
	Lava         int
	Ores         int
	Features     int
	UniqueBiomes int
}

type endReport struct {
	EndStone     int
	Obsidian     int
	Chorus       int
	UniqueBiomes int
}

func main() {
	const seed = int64(0)
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	overworld, err := verifyOverworld(seed)
	if err != nil {
		failf("worldgen verification failed: %v", err)
	}
	nether, err := verifyNether(seed)
	if err != nil {
		failf("worldgen verification failed: %v", err)
	}
	end, err := verifyEnd(seed)
	if err != nil {
		failf("worldgen verification failed: %v", err)
	}

	fmt.Printf(
		"overworld seed=%d min_height=%d max_height=%d variation=%d underground_air=%d underground_water=%d underground_lava=%d ores=%d magma=%d vegetation=%d trees=%d structure_blocks=%d unique_biomes=%d\n",
		seed,
		overworld.MinHeight,
		overworld.MaxHeight,
		overworld.MaxHeight-overworld.MinHeight,
		overworld.UndergroundAir,
		overworld.UndergroundWater,
		overworld.UndergroundLava,
		overworld.Ores,
		overworld.Magma,
		overworld.Vegetation,
		overworld.Trees,
		overworld.StructureBlocks,
		overworld.UniqueBiomes,
	)
	for _, sample := range overworld.Samples {
		fmt.Printf(
			"overworld_sample chunk_x=%d chunk_z=%d height=%d biome=%d top=%s sea=%s air=%d water=%d lava=%d ores=%d magma=%d vegetation=%d trees=%d structure_blocks=%d\n",
			sample.ChunkX,
			sample.ChunkZ,
			sample.Height,
			sample.Biome,
			sample.Top,
			sample.Sea,
			sample.UndergroundAir,
			sample.UndergroundWater,
			sample.UndergroundLava,
			sample.Ores,
			sample.Magma,
			sample.Vegetation,
			sample.Trees,
			sample.StructureBlocks,
		)
	}
	fmt.Printf(
		"nether seed=%d netherrack=%d lava=%d ores=%d features=%d unique_biomes=%d\n",
		seed,
		nether.Netherrack,
		nether.Lava,
		nether.Ores,
		nether.Features,
		nether.UniqueBiomes,
	)
	fmt.Printf(
		"end seed=%d end_stone=%d obsidian=%d chorus=%d unique_biomes=%d\n",
		seed,
		end.EndStone,
		end.Obsidian,
		end.Chorus,
		end.UniqueBiomes,
	)
}

func verifyOverworld(seed int64) (overworldReport, error) {
	g := vanilla.New(seed)
	source, err := gen.NewBiomeSource(seed, gen.NewWorldgenRegistry(), "overworld")
	if err != nil {
		return overworldReport{}, err
	}
	spawnChunk, ok := g.FindSpawnChunk(128)
	if !ok {
		return overworldReport{}, fmt.Errorf("could not find spawn hint chunk")
	}
	plannedStructure, hasPlannedStructure := vanilla.FindPlannedStructureStart(seed, "villages", 24)
	if !hasPlannedStructure {
		return overworldReport{}, fmt.Errorf("could not plan a village structure start")
	}
	structurePalette := make(map[string]struct{}, len(plannedStructure.PaletteNames))
	for _, name := range plannedStructure.PaletteNames {
		structurePalette[name] = struct{}{}
	}

	chunks := []world.ChunkPos{
		spawnChunk,
		{spawnChunk[0] + 1, spawnChunk[1]},
		{spawnChunk[0] - 1, spawnChunk[1]},
		{spawnChunk[0], spawnChunk[1] + 1},
		{spawnChunk[0], spawnChunk[1] - 1},
		{0, 0},
		{16, 16},
		{32, 0},
		{64, 32},
		{-48, 16},
		{96, -32},
		{int32(plannedStructure.Origin.X() >> 4), int32(plannedStructure.Origin.Z() >> 4)},
		{int32((plannedStructure.Origin.X() + plannedStructure.Size[0] - 1) >> 4), int32((plannedStructure.Origin.Z() + plannedStructure.Size[2] - 1) >> 4)},
	}
	chunks = uniqueChunkPositions(chunks)

	report := overworldReport{
		MinHeight: int(^uint(0) >> 1),
		MaxHeight: -int(^uint(0)>>1) - 1,
	}
	seenBiomes := map[gen.Biome]struct{}{}
	airRID := world.BlockRuntimeID(block.Air{})
	minY, maxY := world.Overworld.Range()[0], world.Overworld.Range()[1]

	for i, chunkPos := range chunks {
		c := chunk.New(airRID, world.Overworld.Range())
		g.GenerateChunk(chunkPos, c)

		if i == 0 {
			bottom, ok := blockAt(c, 0, minY, 0)
			if !ok {
				return overworldReport{}, fmt.Errorf("failed to read generated overworld bedrock sample")
			}
			if _, ok := bottom.(block.Bedrock); !ok {
				return overworldReport{}, fmt.Errorf("expected bedrock at world minimum, got %T", bottom)
			}
		}

		collectSourceBiomes(seenBiomes, source, chunkPos, minY, maxY, 16)

		height := int(c.HighestBlock(8, 8))
		for height > minY {
			b, ok := blockAt(c, 8, height, 8)
			if !ok {
				break
			}
			if _, ok := b.(block.Water); !ok {
				break
			}
			height--
		}

		topBlock, _ := blockAt(c, 8, height, 8)
		seaBlock, _ := blockAt(c, 8, seaLevel, 8)
		biome := source.GetBiome(int(chunkPos[0])*16+8, height, int(chunkPos[1])*16+8)

		var undergroundAir, undergroundWater, undergroundLava, vegetation, trees, ores, magma, structureBlocks int
		for y := minY + 1; y < seaLevel; y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := blockAt(c, x, y, z)
					if !ok {
						continue
					}
					switch b.(type) {
					case block.Air:
						undergroundAir++
					case block.Water:
						undergroundWater++
					case block.Lava:
						undergroundLava++
					}
					name, _ := b.EncodeBlock()
					name = strings.TrimPrefix(name, "minecraft:")
					if strings.HasSuffix(name, "_ore") || strings.HasPrefix(name, "infested_") || name == "ancient_debris" {
						ores++
					}
					if name == "magma" {
						magma++
					}
				}
			}
		}

		for y := minY + 1; y <= maxY; y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := blockAt(c, x, y, z)
					if !ok {
						continue
					}
					switch b.(type) {
					case block.ShortGrass, block.DoubleTallGrass, block.Flower, block.Pumpkin:
						vegetation++
					case block.Log, block.Leaves:
						trees++
						continue
					}
					name, _ := b.EncodeBlock()
					name = strings.TrimPrefix(name, "minecraft:")
					if strings.HasSuffix(name, "_log") || strings.HasSuffix(name, "_leaves") {
						trees++
					}
				}
			}
		}

		if intersectsStructureChunk(plannedStructure, int(chunkPos[0]), int(chunkPos[1]), minY, maxY) {
			for y := plannedStructure.Origin.Y(); y < plannedStructure.Origin.Y()+plannedStructure.Size[1]; y++ {
				if y < minY || y > maxY {
					continue
				}
				for x := max(int(chunkPos[0])*16, plannedStructure.Origin.X()); x <= min(int(chunkPos[0])*16+15, plannedStructure.Origin.X()+plannedStructure.Size[0]-1); x++ {
					for z := max(int(chunkPos[1])*16, plannedStructure.Origin.Z()); z <= min(int(chunkPos[1])*16+15, plannedStructure.Origin.Z()+plannedStructure.Size[2]-1); z++ {
						b, ok := blockAt(c, x-int(chunkPos[0])*16, y, z-int(chunkPos[1])*16)
						if !ok {
							continue
						}
						name, _ := b.EncodeBlock()
						if _, ok := structurePalette[name]; ok {
							structureBlocks++
						}
					}
				}
			}
		}

		report.UndergroundAir += undergroundAir
		report.UndergroundWater += undergroundWater
		report.UndergroundLava += undergroundLava
		report.Vegetation += vegetation
		report.Trees += trees
		report.Ores += ores
		report.Magma += magma
		report.StructureBlocks += structureBlocks
		if height < report.MinHeight {
			report.MinHeight = height
		}
		if height > report.MaxHeight {
			report.MaxHeight = height
		}

		report.Samples = append(report.Samples, overworldSample{
			ChunkX:           chunkPos[0],
			ChunkZ:           chunkPos[1],
			Height:           height,
			Biome:            biome,
			Top:              fmt.Sprintf("%T", topBlock),
			Sea:              fmt.Sprintf("%T", seaBlock),
			UndergroundAir:   undergroundAir,
			UndergroundWater: undergroundWater,
			UndergroundLava:  undergroundLava,
			Vegetation:       vegetation,
			Trees:            trees,
			Ores:             ores,
			Magma:            magma,
			StructureBlocks:  structureBlocks,
		})
	}

	report.UniqueBiomes = len(seenBiomes)
	switch {
	case len(report.Samples) == 0:
		return overworldReport{}, fmt.Errorf("no overworld samples collected")
	case report.MaxHeight-report.MinHeight < 10:
		return overworldReport{}, fmt.Errorf("overworld terrain variation too small (min=%d max=%d)", report.MinHeight, report.MaxHeight)
	case report.UndergroundAir == 0:
		return overworldReport{}, fmt.Errorf("no overworld underground air cavities were generated in sampled chunks")
	case report.UndergroundWater == 0 && report.UndergroundLava == 0:
		return overworldReport{}, fmt.Errorf("no overworld aquifer fluids were generated in sampled chunks")
	case report.UniqueBiomes < 2:
		return overworldReport{}, fmt.Errorf("overworld samples only produced one source biome")
	case report.Vegetation == 0:
		return overworldReport{}, fmt.Errorf("overworld sampled chunks had no vegetation or simple placed features")
	case report.Trees == 0:
		return overworldReport{}, fmt.Errorf("overworld sampled chunks had no trees")
	case report.Ores == 0:
		return overworldReport{}, fmt.Errorf("overworld sampled chunks had no underground ore features")
	case report.StructureBlocks == 0:
		return overworldReport{}, fmt.Errorf("planned village structure blocks did not appear in generated chunks")
	}
	return report, nil
}

func verifyNether(seed int64) (netherReport, error) {
	g := vanilla.NewForDimension(seed, world.Nether)
	source, err := gen.NewBiomeSource(seed, gen.NewWorldgenRegistry(), "nether")
	if err != nil {
		return netherReport{}, err
	}

	chunks := []world.ChunkPos{
		{0, 0},
		{8, 0},
		{0, 8},
		{16, 16},
		{-16, 24},
		{32, -32},
		{64, 32},
	}
	chunks = uniqueChunkPositions(chunks)

	report := netherReport{}
	seenBiomes := map[gen.Biome]struct{}{}
	airRID := world.BlockRuntimeID(block.Air{})
	minY, maxY := world.Nether.Range()[0], world.Nether.Range()[1]

	for _, chunkPos := range chunks {
		c := chunk.New(airRID, world.Nether.Range())
		g.GenerateChunk(chunkPos, c)
		collectSourceBiomes(seenBiomes, source, chunkPos, minY, maxY, 16)

		for y := minY; y <= maxY; y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := blockAt(c, x, y, z)
					if !ok {
						continue
					}
					switch b.(type) {
					case block.Netherrack:
						report.Netherrack++
					case block.Lava:
						report.Lava++
					}
					name, _ := b.EncodeBlock()
					name = strings.TrimPrefix(name, "minecraft:")
					switch {
					case strings.HasSuffix(name, "_ore") || name == "ancient_debris":
						report.Ores++
					case name == "glowstone" || name == "magma" || strings.Contains(name, "fungus") || strings.Contains(name, "roots") || strings.Contains(name, "vines") || strings.Contains(name, "nylium") || name == "basalt" || name == "blackstone" || name == "soul_sand" || name == "soul_soil":
						report.Features++
					}
				}
			}
		}
	}

	report.UniqueBiomes = len(seenBiomes)
	switch {
	case report.Netherrack == 0:
		return netherReport{}, fmt.Errorf("sampled Nether chunks had no netherrack terrain")
	case report.Lava == 0:
		return netherReport{}, fmt.Errorf("sampled Nether chunks had no lava")
	case report.Ores == 0:
		return netherReport{}, fmt.Errorf("sampled Nether chunks had no ore features")
	case report.Features == 0:
		return netherReport{}, fmt.Errorf("sampled Nether chunks had no surface or vegetation features")
	case report.UniqueBiomes < 2:
		return netherReport{}, fmt.Errorf("sampled Nether chunks only produced one source biome")
	}
	return report, nil
}

func verifyEnd(seed int64) (endReport, error) {
	g := vanilla.NewForDimension(seed, world.End)
	source, err := gen.NewBiomeSource(seed, gen.NewWorldgenRegistry(), "end")
	if err != nil {
		return endReport{}, err
	}

	chunks := []world.ChunkPos{
		{0, 0},
		{4, 4},
		{6, 0},
		{80, 80},
		{96, 96},
		{112, 80},
		{-80, 80},
		{-96, 96},
	}
	chunks = uniqueChunkPositions(chunks)

	report := endReport{}
	seenBiomes := map[gen.Biome]struct{}{}
	airRID := world.BlockRuntimeID(block.Air{})
	minY, maxY := world.End.Range()[0], world.End.Range()[1]

	for _, chunkPos := range chunks {
		c := chunk.New(airRID, world.End.Range())
		g.GenerateChunk(chunkPos, c)
		collectSourceBiomes(seenBiomes, source, chunkPos, minY, maxY, 32)

		for y := minY; y <= maxY; y++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					b, ok := blockAt(c, x, y, z)
					if !ok {
						continue
					}
					switch b.(type) {
					case block.EndStone:
						report.EndStone++
					case block.Obsidian:
						report.Obsidian++
					case block.ChorusPlant, block.ChorusFlower:
						report.Chorus++
					}
				}
			}
		}
	}

	report.UniqueBiomes = len(seenBiomes)
	switch {
	case report.EndStone == 0:
		return endReport{}, fmt.Errorf("sampled End chunks had no end stone terrain")
	case report.Obsidian == 0:
		return endReport{}, fmt.Errorf("sampled End chunks had no obsidian spike or platform blocks")
	case report.Chorus == 0:
		return endReport{}, fmt.Errorf("sampled End chunks had no chorus features")
	case report.UniqueBiomes < 2:
		return endReport{}, fmt.Errorf("sampled End chunks only produced one source biome")
	}
	return report, nil
}

func collectSourceBiomes(seen map[gen.Biome]struct{}, source gen.BiomeSource, chunkPos world.ChunkPos, minY, maxY, yStep int) {
	for localX := 0; localX < 16; localX += 4 {
		for localZ := 0; localZ < 16; localZ += 4 {
			for y := minY; y <= maxY; y += yStep {
				seen[source.GetBiome(int(chunkPos[0])*16+localX, y, int(chunkPos[1])*16+localZ)] = struct{}{}
			}
		}
	}
}

func uniqueChunkPositions(chunks []world.ChunkPos) []world.ChunkPos {
	out := make([]world.ChunkPos, 0, len(chunks))
	seen := make(map[world.ChunkPos]struct{}, len(chunks))
	for _, pos := range chunks {
		if _, ok := seen[pos]; ok {
			continue
		}
		seen[pos] = struct{}{}
		out = append(out, pos)
	}
	return out
}

func blockAt(c *chunk.Chunk, x, y, z int) (world.Block, bool) {
	return world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func intersectsStructureChunk(info vanilla.PlannedStructureInfo, chunkX, chunkZ, minY, maxY int) bool {
	minBlockX := chunkX * 16
	maxBlockX := minBlockX + 15
	minBlockZ := chunkZ * 16
	maxBlockZ := minBlockZ + 15
	structureMaxY := info.Origin.Y() + info.Size[1] - 1
	return !(info.Origin.X()+info.Size[0]-1 < minBlockX ||
		info.Origin.X() > maxBlockX ||
		info.Origin.Z()+info.Size[2]-1 < minBlockZ ||
		info.Origin.Z() > maxBlockZ ||
		structureMaxY < minY ||
		info.Origin.Y() > maxY)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var finaliseBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()
