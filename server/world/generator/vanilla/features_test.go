package vanilla

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func TestGenerateChunkDecoratesVegetation(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	positions := []world.ChunkPos{
		{0, 0},
		{16, 16},
		{32, 0},
		{64, 32},
		{-48, 16},
	}

	totalDecor := 0
	for _, pos := range positions {
		c := chunk.New(g.airRID, cube.Range{-64, 319})
		g.GenerateChunk(pos, c)
		totalDecor += countDecorativeBlocks(c)
	}

	if totalDecor == 0 {
		t.Fatal("expected generated sample chunks to contain vegetation or simple placed features")
	}
}

func TestGenerateChunkPlacesUndergroundFeatures(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	positions := []world.ChunkPos{
		{0, 0},
		{16, 16},
		{32, 0},
		{64, 32},
		{-48, 16},
	}

	totalOres := 0
	for _, pos := range positions {
		c := chunk.New(g.airRID, cube.Range{-64, 319})
		g.GenerateChunk(pos, c)
		totalOres += countOreBlocks(c)
	}

	if totalOres == 0 {
		t.Fatal("expected generated sample chunks to contain underground ore features")
	}
}

func TestExecuteSpringFeaturePlacesSource(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	y := 32
	center := cube.Pos{8, y, 8}
	stoneRID := world.BlockRuntimeID(block.Stone{})

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for dy := y - 2; dy <= y+2; dy++ {
				c.SetBlock(uint8(x), int16(dy), uint8(z), 0, stoneRID)
			}
		}
	}
	c.SetBlock(uint8(center[0]), int16(center[1]), uint8(center[2]), 0, g.airRID)
	c.SetBlock(uint8(center[0]+1), int16(center[1]), uint8(center[2]), 0, g.airRID)

	cfg, err := g.features.Configured("spring_water")
	if err != nil {
		t.Fatalf("failed to load spring_water: %v", err)
	}
	spring, err := cfg.SpringFeature()
	if err != nil {
		t.Fatalf("failed to decode spring_water: %v", err)
	}
	if !g.executeSpringFeature(c, center, spring, 0, 0, c.Range().Min(), c.Range().Max(), nil) {
		t.Fatal("expected spring feature to place a fluid source")
	}

	b, ok := world.BlockByRuntimeID(c.Block(uint8(center[0]), int16(center[1]), uint8(center[2]), 0))
	if !ok {
		t.Fatal("expected spring center block to exist")
	}
	if _, ok := b.(block.Water); !ok {
		t.Fatalf("expected spring to place water, got %T", b)
	}
}

func TestFeatureBlockFromStateNormalizesTreeStates(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)

	logBlock, ok := g.featureBlockFromState(gen.BlockState{
		Name: "oak_log",
		Properties: map[string]string{
			"axis": "y",
		},
	}, nil)
	if !ok {
		t.Fatal("expected oak_log state to resolve")
	}
	if _, ok := logBlock.(block.Log); !ok {
		t.Fatalf("expected oak_log to resolve to block.Log, got %T", logBlock)
	}

	leafBlock, ok := g.featureBlockFromState(gen.BlockState{
		Name: "oak_leaves",
		Properties: map[string]string{
			"distance":    "7",
			"persistent":  "false",
			"waterlogged": "false",
		},
	}, nil)
	if !ok {
		t.Fatal("expected oak_leaves state to resolve")
	}
	if _, ok := leafBlock.(block.Leaves); !ok {
		t.Fatalf("expected oak_leaves to resolve to block.Leaves, got %T", leafBlock)
	}
}

func TestFeatureBlockFromStateFallsBackForUnsupportedJavaBlocks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	tests := []struct {
		name       string
		state      gen.BlockState
		expectType any
	}{
		{
			name:       "bamboo",
			state:      gen.BlockState{Name: "minecraft:bamboo"},
			expectType: block.SugarCane{},
		},
		{
			name:       "rooted_dirt",
			state:      gen.BlockState{Name: "minecraft:rooted_dirt"},
			expectType: block.Dirt{},
		},
		{
			name: "leaf_litter",
			state: gen.BlockState{
				Name: "minecraft:leaf_litter",
				Properties: map[string]string{
					"facing":         "north",
					"segment_amount": "4",
				},
			},
			expectType: block.ShortGrass{},
		},
		{
			name: "big_dripleaf",
			state: gen.BlockState{
				Name: "minecraft:big_dripleaf",
				Properties: map[string]string{
					"facing":      "east",
					"tilt":        "none",
					"waterlogged": "false",
				},
			},
			expectType: block.SugarCane{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, ok := g.featureBlockFromState(tt.state, nil)
			if !ok {
				t.Fatalf("expected %s fallback state to resolve", tt.name)
			}
			if reflect.TypeOf(b) != reflect.TypeOf(tt.expectType) {
				t.Fatalf("expected %s to resolve to %T, got %T", tt.name, tt.expectType, b)
			}
		})
	}
}

func TestMangrovePropaguleSurvivesOnMud(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	c.SetBlock(8, 0, 8, 0, world.BlockRuntimeID(block.Mud{}))

	ok := g.canBlockStateSurvive(c, cube.Pos{8, 1, 8}, gen.BlockState{
		Name: "minecraft:mangrove_propagule",
		Properties: map[string]string{
			"age":         "0",
			"hanging":     "true",
			"stage":       "0",
			"waterlogged": "false",
		},
	}, nil, c.Range().Min(), c.Range().Max())
	if !ok {
		t.Fatal("expected mangrove propagule to survive on mud")
	}
}

func TestExecuteKelpPreservesWaterLayer(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	pos := cube.Pos{8, 4, 8}
	c.SetBlock(uint8(pos[0]), int16(pos[1]-1), uint8(pos[2]), 0, world.BlockRuntimeID(block.Stone{}))
	for y := pos[1]; y <= pos[1]+8; y++ {
		c.SetBlock(uint8(pos[0]), int16(y), uint8(pos[2]), 0, g.waterRID)
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	if !g.executeKelp(c, pos, c.Range().Min(), c.Range().Max(), &rng) {
		t.Fatal("expected kelp feature to place at least one kelp block")
	}

	placed, ok := world.BlockByRuntimeID(c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0))
	if !ok {
		t.Fatal("expected placed kelp block to resolve")
	}
	if _, ok := placed.(block.Kelp); !ok {
		t.Fatalf("expected kelp at placement position, got %T", placed)
	}
	if c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 1) != g.waterRID {
		t.Fatal("expected kelp to preserve source water in chunk layer 1")
	}
}

func TestExecuteSeaPicklePreservesWaterLayer(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	pos := cube.Pos{8, 4, 8}
	c.SetBlock(uint8(pos[0]), int16(pos[1]-1), uint8(pos[2]), 0, world.BlockRuntimeID(block.Stone{}))
	c.SetBlock(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, g.waterRID)

	rng := gen.NewXoroshiro128FromSeed(1)
	if !g.executeSeaPickle(c, pos, gen.SeaPickleConfig{Count: 1}, 0, 0, c.Range().Min(), c.Range().Max(), &rng) {
		t.Fatal("expected sea pickle feature to place a plant")
	}

	if c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0) == g.waterRID {
		t.Fatal("expected sea pickle block to replace foreground water")
	}
	if c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 1) != g.waterRID {
		t.Fatal("expected sea pickle block to preserve source water in chunk layer 1")
	}
}

func TestExecuteConfiguredOakTreePlacesBlocks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, world.BlockRuntimeID(block.Grass{}))
		}
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomePlains)
	if !g.executeConfiguredFeature(c, biomes, cube.Pos{8, 1, 8}, gen.ConfiguredFeatureRef{Name: "oak"}, "plains", 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0) {
		t.Fatal("expected oak configured feature to place a tree")
	}
	if countTreeBlocks(c) == 0 {
		t.Fatal("expected oak configured feature to create logs or leaves")
	}
}

func TestExecuteNetherQuartzOrePlacesBlocks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.Nether)
	c := chunk.New(g.airRID, world.Nether.Range())
	for y := c.Range().Min(); y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				c.SetBlock(uint8(x), int16(y), uint8(z), 0, world.BlockRuntimeID(block.Netherrack{}))
			}
		}
	}

	cfg, err := g.features.Configured("ore_quartz")
	if err != nil {
		t.Fatalf("failed to load ore_quartz: %v", err)
	}
	ore, err := cfg.Ore()
	if err != nil {
		t.Fatalf("failed to decode ore_quartz: %v", err)
	}
	rng := gen.NewXoroshiro128FromSeed(1)
	if !g.executeOre(c, cube.Pos{8, 32, 8}, ore, 0, 0, c.Range().Min(), c.Range().Max(), &rng, false) {
		t.Fatal("expected nether quartz ore feature to place at least one block")
	}

	totalOres := 0
	for y := c.Range().Min(); y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				name, _ := b.EncodeBlock()
				if strings.HasSuffix(strings.TrimPrefix(name, "minecraft:"), "_ore") {
					totalOres++
				}
			}
		}
	}
	if totalOres == 0 {
		t.Fatal("expected executeOre to leave quartz ore blocks in the chunk")
	}
}

func TestRunPlacedFeatureNetherQuartzPlacesBlocks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := NewForDimension(0, world.Nether)
	c := chunk.New(g.airRID, world.Nether.Range())
	for y := c.Range().Min(); y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				c.SetBlock(uint8(x), int16(y), uint8(z), 0, world.BlockRuntimeID(block.Netherrack{}))
				c.SetBiome(uint8(x), int16(y), uint8(z), biomeRuntimeID(gen.BiomeNetherWastes))
			}
		}
	}

	placed, err := g.features.Placed("ore_quartz_nether")
	if err != nil {
		t.Fatalf("failed to load ore_quartz_nether: %v", err)
	}
	biomeName := biomeKey(gen.BiomeNetherWastes)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomeNetherWastes)
	rng := g.featureRNG(0, 0, biomeName, "ore_quartz_nether")
	positions, ok := g.applyPlacementModifiers(c, biomes, []cube.Pos{{0, c.Range().Min(), 0}}, placed.Placement, biomeName, 0, 0, c.Range().Min(), c.Range().Max(), &rng)
	if !ok {
		t.Fatal("expected placement modifiers for ore_quartz_nether to be supported")
	}
	if len(positions) == 0 {
		t.Fatal("expected ore_quartz_nether placement modifiers to produce candidate positions")
	}
	for _, pos := range positions {
		g.executeConfiguredFeature(c, biomes, pos, placed.Feature, biomeName, 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0)
	}

	totalOres := 0
	for y := c.Range().Min(); y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				name, _ := b.EncodeBlock()
				if strings.HasSuffix(strings.TrimPrefix(name, "minecraft:"), "_ore") {
					totalOres++
				}
			}
		}
	}
	if totalOres == 0 {
		t.Fatal("expected placed feature ore_quartz_nether to leave ore blocks in the chunk")
	}
}

func TestExecuteConfiguredOakTreeHasRoundedTop(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, world.BlockRuntimeID(block.Grass{}))
		}
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomePlains)
	if !g.executeConfiguredFeature(c, biomes, cube.Pos{8, 1, 8}, gen.ConfiguredFeatureRef{Name: "oak"}, "plains", 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0) {
		t.Fatal("expected oak configured feature to place a tree")
	}

	top, ok := highestTreeLog(c)
	if !ok {
		t.Fatal("expected placed oak tree to contain a log")
	}
	cardinalLeaves := 0
	for _, off := range []cube.Pos{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {
		if isLeafBlockAt(c, top.Add(off)) {
			cardinalLeaves++
		}
	}
	if cardinalLeaves < 3 {
		t.Fatalf("expected oak canopy around top log to place cardinal leaves, got %d", cardinalLeaves)
	}
	for _, off := range []cube.Pos{{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1}} {
		if isLeafBlockAt(c, top.Add(off)) {
			t.Fatalf("expected oak canopy diagonals around top log to be open, found leaves at %v", top.Add(off))
		}
	}
}

func TestExecutePlacedOakCheckedPlacesBlocks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, world.BlockRuntimeID(block.Grass{}))
		}
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomePlains)
	if !g.executePlacedFeatureRef(c, biomes, cube.Pos{8, 1, 8}, gen.PlacedFeatureRef{Name: "oak_checked"}, "plains", 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0) {
		t.Fatal("expected oak_checked placed feature to place a tree")
	}
	if countTreeBlocks(c) == 0 {
		t.Fatal("expected oak_checked placed feature to create logs or leaves")
	}
}

func TestHeightmapPlacementYCountsWaterAboveOceanFloor(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	c.SetBlock(8, 0, 8, 0, world.BlockRuntimeID(block.Grass{}))
	c.SetBlock(8, 1, 8, 0, g.waterRID)
	c.SetBlock(8, 2, 8, 0, g.waterRID)

	worldSurface := g.heightmapPlacementY(c, 8, 8, "WORLD_SURFACE", c.Range().Min(), c.Range().Max())
	if worldSurface != 3 {
		t.Fatalf("expected world surface height 3, got %d", worldSurface)
	}
	oceanFloor := g.heightmapPlacementY(c, 8, 8, "OCEAN_FLOOR", c.Range().Min(), c.Range().Max())
	if oceanFloor != 1 {
		t.Fatalf("expected ocean floor height 1, got %d", oceanFloor)
	}
	if depth := g.surfaceWaterDepthAt(c, 8, 8, c.Range().Min()); depth != 2 {
		t.Fatalf("expected surface water depth 2, got %d", depth)
	}
}

func TestTreesBirchPlacementSkipsWaterColumns(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	grassRID := world.BlockRuntimeID(block.Grass{})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, grassRID)
			c.SetBlock(uint8(x), 1, uint8(z), 0, g.waterRID)
			c.SetBlock(uint8(x), 2, uint8(z), 0, g.waterRID)
		}
	}

	placed, err := g.features.Placed("trees_birch")
	if err != nil {
		t.Fatalf("failed to load trees_birch: %v", err)
	}
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomeBirchForest)
	rng := g.featureRNG(0, 0, biomeKey(gen.BiomeBirchForest), "trees_birch")
	positions, ok := g.applyPlacementModifiers(c, biomes, []cube.Pos{{0, c.Range().Min(), 0}}, placed.Placement, biomeKey(gen.BiomeBirchForest), 0, 0, c.Range().Min(), c.Range().Max(), &rng)
	if !ok {
		t.Fatal("expected trees_birch placement modifiers to be supported")
	}
	if len(positions) != 0 {
		t.Fatalf("expected water depth filter to reject birch tree placements, got %d position(s)", len(positions))
	}
}

func TestGenerateChunkAtSpawnHintPlacesTrees(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	hint, ok := g.FindSpawnChunk(128)
	if !ok {
		t.Fatal("expected to find a spawn hint chunk")
	}

	c := chunk.New(g.airRID, cube.Range{-64, 319})
	g.GenerateChunk(hint, c)
	if countTreeBlocks(c) == 0 {
		t.Fatalf("expected spawn hint chunk %v to contain trees", hint)
	}
}

func TestExecuteConfiguredBambooPlacesStalks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, world.BlockRuntimeID(block.Grass{}))
		}
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomeBambooJungle)
	if !g.executeConfiguredFeature(c, biomes, cube.Pos{8, 1, 8}, gen.ConfiguredFeatureRef{Name: "bamboo_some_podzol"}, "bamboo_jungle", 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0) {
		t.Fatal("expected bamboo configured feature to place stalks")
	}

	found := false
	for y := 1; y <= c.Range().Max(); y++ {
		b, ok := world.BlockByRuntimeID(c.Block(8, int16(y), 8, 0))
		if !ok {
			continue
		}
		if _, ok := b.(block.SugarCane); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected bamboo fallback stalks to be present")
	}
}

func TestExecuteConfiguredMossPatchPlacesGround(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	g := New(0)
	c := chunk.New(g.airRID, cube.Range{-64, 319})
	stoneRID := world.BlockRuntimeID(block.Stone{})
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(uint8(x), 0, uint8(z), 0, stoneRID)
		}
	}

	rng := gen.NewXoroshiro128FromSeed(1)
	biomes := filledTestBiomeVolume(c.Range().Min(), c.Range().Max(), gen.BiomeLushCaves)
	if !g.executeConfiguredFeature(c, biomes, cube.Pos{8, 1, 8}, gen.ConfiguredFeatureRef{Name: "moss_patch"}, "lush_caves", 0, 0, c.Range().Min(), c.Range().Max(), &rng, 0) {
		t.Fatal("expected moss_patch configured feature to place ground blocks")
	}

	found := false
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			b, ok := world.BlockByRuntimeID(c.Block(uint8(x), 0, uint8(z), 0))
			if !ok {
				continue
			}
			name, _ := b.EncodeBlock()
			if strings.TrimPrefix(name, "minecraft:") == "moss_block" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected vegetation patch to replace ground with moss")
	}
}

func countDecorativeBlocks(c *chunk.Chunk) int {
	total := 0
	for y := c.Range().Min() + 1; y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				switch b.(type) {
				case block.ShortGrass, block.DoubleTallGrass, block.Flower, block.Pumpkin:
					total++
				}
			}
		}
	}
	return total
}

func countOreBlocks(c *chunk.Chunk) int {
	total := 0
	for y := c.Range().Min() + 1; y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				name, _ := b.EncodeBlock()
				name = strings.TrimPrefix(name, "minecraft:")
				if strings.HasSuffix(name, "_ore") || strings.HasPrefix(name, "infested_") || name == "ancient_debris" {
					total++
				}
			}
		}
	}
	return total
}

func countTreeBlocks(c *chunk.Chunk) int {
	total := 0
	for y := c.Range().Min() + 1; y <= c.Range().Max(); y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				switch b.(type) {
				case block.Log, block.Leaves:
					total++
					continue
				}
				name, _ := b.EncodeBlock()
				name = strings.TrimPrefix(name, "minecraft:")
				if strings.HasSuffix(name, "_log") || strings.HasSuffix(name, "_leaves") {
					total++
				}
			}
		}
	}
	return total
}

func highestTreeLog(c *chunk.Chunk) (cube.Pos, bool) {
	for y := c.Range().Max(); y >= c.Range().Min()+1; y-- {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				b, ok := world.BlockByRuntimeID(c.Block(uint8(x), int16(y), uint8(z), 0))
				if !ok {
					continue
				}
				switch b.(type) {
				case block.Log:
					return cube.Pos{x, y, z}, true
				}
				name, _ := b.EncodeBlock()
				if strings.HasSuffix(strings.TrimPrefix(name, "minecraft:"), "_log") {
					return cube.Pos{x, y, z}, true
				}
			}
		}
	}
	return cube.Pos{}, false
}

func isLeafBlockAt(c *chunk.Chunk, pos cube.Pos) bool {
	if pos[1] <= c.Range().Min() || pos[1] > c.Range().Max() {
		return false
	}
	b, ok := world.BlockByRuntimeID(c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0))
	if !ok {
		return false
	}
	switch b.(type) {
	case block.Leaves:
		return true
	}
	name, _ := b.EncodeBlock()
	return strings.HasSuffix(strings.TrimPrefix(name, "minecraft:"), "_leaves")
}

func filledTestBiomeVolume(minY, maxY int, biome gen.Biome) sourceBiomeVolume {
	volume := newSourceBiomeVolume(minY, maxY)
	for x := 0; x < 16; x += biomeCellSize {
		for z := 0; z < 16; z += biomeCellSize {
			for y := alignDown(minY, biomeCellSize); y <= maxY; y += biomeCellSize {
				volume.set(x, y, z, biome)
			}
		}
	}
	return volume
}

var finaliseBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()
