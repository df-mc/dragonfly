package vanilla

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func (g Generator) decorateFeatures(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int) {
	if g.features == nil {
		return
	}

	surfaceBiomes := g.collectChunkBiomes(c, biomes, minY, maxY, true)
	chunkBiomes := g.collectChunkBiomes(c, biomes, minY, maxY, false)

	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, surfaceBiomes, gen.GenerationStepLakes)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepRawGeneration)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepLocalModifications)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepUndergroundStructures)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, surfaceBiomes, gen.GenerationStepSurfaceStructures)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepStrongholds)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepUndergroundOres)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepUndergroundDecoration)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, chunkBiomes, gen.GenerationStepFluidSprings)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, surfaceBiomes, gen.GenerationStepVegetalDecoration)
	g.decorateStep(c, biomes, chunkX, chunkZ, minY, maxY, surfaceBiomes, gen.GenerationStepTopLayerModification)
}

func (g Generator) decorateStep(c *chunk.Chunk, sourceBiomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int, biomes []gen.Biome, step gen.GenerationStep) {
	if len(biomes) == 0 {
		return
	}

	for _, biome := range biomes {
		features := g.biomeGeneration.featureSteps[biome][int(step)]
		if len(features) == 0 {
			continue
		}
		biomeKey := biomeKey(biome)
		for _, featureName := range features {
			g.runPlacedFeature(c, sourceBiomes, chunkX, chunkZ, minY, maxY, biomeKey, featureName)
		}
	}
}

func (g Generator) collectChunkBiomes(c *chunk.Chunk, biomes sourceBiomeVolume, minY, maxY int, surfaceOnly bool) []gen.Biome {
	var seen [256]bool
	if surfaceOnly {
		for localX := 0; localX < 16; localX++ {
			for localZ := 0; localZ < 16; localZ++ {
				surfaceY := g.heightmapPlacementY(c, localX, localZ, "WORLD_SURFACE", minY, maxY) - 1
				if surfaceY < minY {
					surfaceY = minY
				}
				if surfaceY > maxY {
					surfaceY = maxY
				}
				seen[biomes.biomeAt(localX, surfaceY, localZ)] = true
			}
		}
	} else {
		for localX := 0; localX < 16; localX += 4 {
			for localZ := 0; localZ < 16; localZ += 4 {
				for y := minY; y <= maxY; y += 4 {
					seen[biomes.biomeAt(localX, y, localZ)] = true
				}
			}
		}
	}

	out := make([]gen.Biome, 0, 8)
	for _, biome := range sortedBiomesByKey {
		if seen[biome] {
			out = append(out, biome)
		}
	}
	return out
}

func (g Generator) runPlacedFeature(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int, biomeKey, featureName string) {
	placed, err := g.features.Placed(featureName)
	if err != nil {
		return
	}

	origin := cube.Pos{chunkX * 16, minY, chunkZ * 16}
	rng := g.featureRNG(chunkX, chunkZ, biomeKey, featureName)
	positions, ok := g.applyPlacementModifiers(c, biomes, []cube.Pos{origin}, placed.Placement, biomeKey, chunkX, chunkZ, minY, maxY, &rng)
	if !ok {
		return
	}

	for _, pos := range positions {
		g.executeConfiguredFeature(c, biomes, pos, placed.Feature, biomeKey, chunkX, chunkZ, minY, maxY, &rng, 0)
	}
}

func (g Generator) executeConfiguredFeature(c *chunk.Chunk, biomes sourceBiomeVolume, pos cube.Pos, featureRef gen.ConfiguredFeatureRef, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, depth int) bool {
	if depth > 8 {
		return false
	}

	feature, err := g.features.ResolveConfigured(featureRef)
	if err != nil {
		return false
	}

	switch feature.Type {
	case "random_patch":
		cfg, err := feature.RandomPatch()
		if err != nil {
			return false
		}
		return g.executeRandomPatch(c, biomes, pos, cfg, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "flower":
		cfg, err := feature.Flower()
		if err != nil {
			return false
		}
		return g.executeRandomPatch(c, biomes, pos, cfg, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "simple_block":
		cfg, err := feature.SimpleBlock()
		if err != nil {
			return false
		}
		return g.placeStateProviderBlock(c, pos, cfg.ToPlace, rng, minY, maxY)
	case "block_column":
		cfg, err := feature.BlockColumn()
		if err != nil {
			return false
		}
		return g.executeBlockColumn(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "random_selector":
		cfg, err := feature.RandomSelector()
		if err != nil {
			return false
		}
		for _, entry := range cfg.Features {
			if rng.NextDouble() < entry.Chance {
				return g.executePlacedFeatureRef(c, biomes, pos, entry.Feature, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
			}
		}
		return g.executePlacedFeatureRef(c, biomes, pos, cfg.Default, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "simple_random_selector":
		cfg, err := feature.SimpleRandomSelector()
		if err != nil || len(cfg.Features) == 0 {
			return false
		}
		ref := cfg.Features[int(rng.NextInt(uint32(len(cfg.Features))))]
		return g.executePlacedFeatureRef(c, biomes, pos, ref, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "random_boolean_selector":
		cfg, err := feature.RandomBooleanSelector()
		if err != nil {
			return false
		}
		if rng.NextDouble() < 0.5 {
			return g.executePlacedFeatureRef(c, biomes, pos, cfg.FeatureTrue, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
		}
		return g.executePlacedFeatureRef(c, biomes, pos, cfg.FeatureFalse, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "seagrass":
		cfg, err := feature.Seagrass()
		if err != nil {
			return false
		}
		return g.executeSeagrass(c, pos, cfg, minY, maxY, rng)
	case "kelp":
		return g.executeKelp(c, pos, minY, maxY, rng)
	case "multiface_growth":
		cfg, err := feature.MultifaceGrowth()
		if err != nil {
			return false
		}
		return g.executeMultifaceGrowth(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "ore":
		cfg, err := feature.Ore()
		if err != nil {
			return false
		}
		return g.executeOre(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng, false)
	case "scattered_ore":
		cfg, err := feature.ScatteredOre()
		if err != nil {
			return false
		}
		return g.executeOre(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng, true)
	case "disk":
		cfg, err := feature.Disk()
		if err != nil {
			return false
		}
		return g.executeDisk(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "spring_feature":
		cfg, err := feature.SpringFeature()
		if err != nil {
			return false
		}
		return g.executeSpringFeature(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "underwater_magma":
		cfg, err := feature.UnderwaterMagma()
		if err != nil {
			return false
		}
		return g.executeUnderwaterMagma(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "pointed_dripstone":
		cfg, err := feature.PointedDripstone()
		if err != nil {
			return false
		}
		return g.executePointedDripstone(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "dripstone_cluster":
		cfg, err := feature.DripstoneCluster()
		if err != nil {
			return false
		}
		return g.executeDripstoneCluster(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "sculk_patch":
		cfg, err := feature.SculkPatch()
		if err != nil {
			return false
		}
		return g.executeSculkPatch(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "vines":
		return g.executeVines(c, pos, chunkX, chunkZ, minY, maxY, rng)
	case "sea_pickle":
		cfg, err := feature.SeaPickle()
		if err != nil {
			return false
		}
		return g.executeSeaPickle(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "lake":
		cfg, err := feature.Lake()
		if err != nil {
			return false
		}
		return g.executeLake(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "freeze_top_layer":
		cfg, err := feature.FreezeTopLayer()
		if err != nil {
			return false
		}
		return g.executeFreezeTopLayer(c, biomes, biomeKey, cfg, chunkX, chunkZ, minY, maxY)
	case "bamboo":
		cfg, err := feature.Bamboo()
		if err != nil {
			return false
		}
		return g.executeBamboo(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "vegetation_patch":
		cfg, err := feature.VegetationPatch()
		if err != nil {
			return false
		}
		return g.executeVegetationPatch(c, biomes, pos, cfg, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1, false)
	case "waterlogged_vegetation_patch":
		cfg, err := feature.WaterloggedVegetationPatch()
		if err != nil {
			return false
		}
		return g.executeVegetationPatch(c, biomes, pos, cfg, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1, true)
	case "root_system":
		cfg, err := feature.RootSystem()
		if err != nil {
			return false
		}
		return g.executeRootSystem(c, biomes, pos, cfg, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1)
	case "fallen_tree":
		cfg, err := feature.FallenTree()
		if err != nil {
			return false
		}
		return g.executeFallenTree(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "tree":
		cfg, err := feature.Tree()
		if err != nil {
			return false
		}
		return g.executeTree(c, pos, cfg, minY, maxY, rng)
	case "huge_fungus":
		cfg, err := feature.HugeFungus()
		if err != nil {
			return false
		}
		return g.executeHugeFungus(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "nether_forest_vegetation":
		cfg, err := feature.NetherForestVegetation()
		if err != nil {
			return false
		}
		return g.executeNetherForestVegetation(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "twisting_vines":
		cfg, err := feature.TwistingVines()
		if err != nil {
			return false
		}
		return g.executeTwistingVines(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "weeping_vines":
		cfg, err := feature.WeepingVines()
		if err != nil {
			return false
		}
		return g.executeWeepingVines(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "netherrack_replace_blobs":
		cfg, err := feature.NetherrackReplaceBlobs()
		if err != nil {
			return false
		}
		return g.executeNetherrackReplaceBlobs(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "glowstone_blob":
		cfg, err := feature.GlowstoneBlob()
		if err != nil {
			return false
		}
		return g.executeGlowstoneBlob(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "basalt_pillar":
		cfg, err := feature.BasaltPillar()
		if err != nil {
			return false
		}
		return g.executeBasaltPillar(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "basalt_columns":
		cfg, err := feature.BasaltColumns()
		if err != nil {
			return false
		}
		return g.executeBasaltColumns(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "delta_feature":
		cfg, err := feature.DeltaFeature()
		if err != nil {
			return false
		}
		return g.executeDeltaFeature(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "chorus_plant":
		cfg, err := feature.ChorusPlant()
		if err != nil {
			return false
		}
		return g.executeChorusPlant(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "end_island":
		cfg, err := feature.EndIsland()
		if err != nil {
			return false
		}
		return g.executeEndIsland(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "end_spike":
		cfg, err := feature.EndSpike()
		if err != nil {
			return false
		}
		return g.executeEndSpike(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng)
	case "end_platform":
		cfg, err := feature.EndPlatform()
		if err != nil {
			return false
		}
		return g.executeEndPlatform(c, pos, cfg, chunkX, chunkZ, minY, maxY)
	case "end_gateway":
		cfg, err := feature.EndGateway()
		if err != nil {
			return false
		}
		return g.executeEndGateway(c, pos, cfg, chunkX, chunkZ, minY, maxY)
	default:
		return false
	}
}

func (g Generator) executePlacedFeatureRef(c *chunk.Chunk, biomes sourceBiomeVolume, pos cube.Pos, placedRef gen.PlacedFeatureRef, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, depth int) bool {
	if depth > 8 {
		return false
	}

	placed, err := g.features.ResolvePlaced(placedRef)
	if err != nil {
		return false
	}
	positions, ok := g.applyPlacementModifiers(c, biomes, []cube.Pos{pos}, placed.Placement, biomeKey, chunkX, chunkZ, minY, maxY, rng)
	if !ok {
		return false
	}

	var placedAny bool
	for _, candidate := range positions {
		if g.executeConfiguredFeature(c, biomes, candidate, placed.Feature, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1) {
			placedAny = true
		}
	}
	return placedAny
}

func (g Generator) executeRandomPatch(c *chunk.Chunk, biomes sourceBiomeVolume, origin cube.Pos, cfg gen.RandomPatchConfig, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, depth int) bool {
	var placedAny bool
	for attempt := 0; attempt < cfg.Tries; attempt++ {
		pos := origin.Add(cube.Pos{
			g.signedSpread(rng, cfg.XZSpread),
			g.signedSpread(rng, cfg.YSpread),
			g.signedSpread(rng, cfg.XZSpread),
		})
		if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if g.executePlacedFeatureRef(c, biomes, pos, cfg.Feature, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1) {
			placedAny = true
		}
	}
	return placedAny
}

func (g Generator) executeBlockColumn(c *chunk.Chunk, origin cube.Pos, cfg gen.BlockColumnConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	dir := blockColumnDirection(cfg.Direction)
	if dir == (cube.Pos{}) {
		return false
	}

	current := origin
	var placedAny bool
	for _, layer := range cfg.Layers {
		height := max(0, g.sampleIntProvider(layer.Height, rng))
		for i := 0; i < height; i++ {
			if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
				return placedAny
			}
			if !g.testBlockPredicate(c, current, cfg.AllowedPlacement, chunkX, chunkZ, minY, maxY, rng) {
				return placedAny
			}
			if !g.placeStateProviderBlock(c, current, layer.Provider, rng, minY, maxY) {
				return placedAny
			}
			placedAny = true
			current = current.Add(dir)
		}
	}
	return placedAny
}

func (g Generator) executeSeagrass(c *chunk.Chunk, pos cube.Pos, cfg gen.SeagrassConfig, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if pos[1] <= minY || pos[1] >= maxY {
		return false
	}
	if c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0) != g.waterRID {
		return false
	}
	belowRID := c.Block(uint8(pos[0]&15), int16(pos[1]-1), uint8(pos[2]&15), 0)
	if !g.isSolidRID(belowRID) {
		return false
	}

	if cfg.Probability > 0 && rng.NextDouble() < cfg.Probability {
		upper := pos.Side(cube.FaceUp)
		if upper[1] <= maxY && c.Block(uint8(upper[0]&15), int16(upper[1]), uint8(upper[2]&15), 0) == g.waterRID {
			return g.setBlockStateDirect(c, pos, gen.BlockState{Name: "tall_seagrass", Properties: map[string]string{"half": "lower"}}) &&
				g.setBlockStateDirect(c, upper, gen.BlockState{Name: "tall_seagrass", Properties: map[string]string{"half": "upper"}})
		}
	}
	return g.setBlockStateDirect(c, pos, gen.BlockState{Name: "seagrass"})
}

func (g Generator) executeKelp(c *chunk.Chunk, pos cube.Pos, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if pos[1] <= minY || pos[1] >= maxY {
		return false
	}
	if c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0) != g.waterRID {
		return false
	}

	height := 1 + int(rng.NextInt(10))
	var placedAny bool
	for i := 0; i < height && pos[1]+i <= maxY; i++ {
		current := pos.Add(cube.Pos{0, i, 0})
		if c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.waterRID {
			break
		}
		if !g.setBlockStateDirect(c, current, gen.BlockState{Name: "kelp", Properties: map[string]string{"age": strconv.Itoa(int(rng.NextInt(25)))}}) {
			break
		}
		placedAny = true
	}
	return placedAny
}

func (g Generator) executeMultifaceGrowth(c *chunk.Chunk, pos cube.Pos, cfg gen.MultifaceGrowthConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	for attempt := 0; attempt <= max(1, cfg.SearchRange); attempt++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(uint32(max(1, cfg.SearchRange*2+1)))) - cfg.SearchRange,
			int(rng.NextInt(uint32(max(1, cfg.SearchRange*2+1)))) - cfg.SearchRange,
			int(rng.NextInt(uint32(max(1, cfg.SearchRange*2+1)))) - cfg.SearchRange,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		rid := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
		if rid != g.airRID && rid != g.waterRID {
			continue
		}
		if state, ok := g.multifaceStateAt(c, candidate, cfg, chunkX, chunkZ, minY, maxY); ok {
			return g.setBlockStateDirect(c, candidate, state)
		}
	}
	return false
}

func (g Generator) executeOre(c *chunk.Chunk, pos cube.Pos, cfg gen.OreConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, scattered bool) bool {
	if cfg.Size <= 0 {
		return false
	}
	if scattered {
		var placedAny bool
		for i := 0; i < cfg.Size; i++ {
			candidate := pos.Add(cube.Pos{
				int(rng.NextInt(5)) - 2,
				int(rng.NextInt(5)) - 2,
				int(rng.NextInt(5)) - 2,
			})
			if g.tryPlaceOreAt(c, candidate, cfg, chunkX, chunkZ, minY, maxY, rng) {
				placedAny = true
			}
		}
		return placedAny
	}

	angle := rng.NextDouble() * math.Pi
	spread := float64(cfg.Size) / 8.0
	x1 := float64(pos[0]) + math.Sin(angle)*spread
	x2 := float64(pos[0]) - math.Sin(angle)*spread
	z1 := float64(pos[2]) + math.Cos(angle)*spread
	z2 := float64(pos[2]) - math.Cos(angle)*spread
	y1 := float64(pos[1] + int(rng.NextInt(3)) - 1)
	y2 := float64(pos[1] + int(rng.NextInt(3)) - 1)

	var placedAny bool
	for i := 0; i < cfg.Size; i++ {
		t := float64(i) / float64(cfg.Size)
		cx := lerp(t, x1, x2)
		cy := lerp(t, y1, y2)
		cz := lerp(t, z1, z2)
		radius := ((1.0-math.Abs(2.0*t-1.0))*float64(cfg.Size)/16.0 + 1.0) / 2.0
		minX, maxX := int(math.Floor(cx-radius)), int(math.Ceil(cx+radius))
		minZ, maxZ := int(math.Floor(cz-radius)), int(math.Ceil(cz+radius))
		minBlockY, maxBlockY := int(math.Floor(cy-radius)), int(math.Ceil(cy+radius))
		for x := minX; x <= maxX; x++ {
			for y := minBlockY; y <= maxBlockY; y++ {
				for z := minZ; z <= maxZ; z++ {
					candidate := cube.Pos{x, y, z}
					if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
						continue
					}
					dx, dy, dz := float64(x)-cx, float64(y)-cy, float64(z)-cz
					if dx*dx+dy*dy+dz*dz > radius*radius {
						continue
					}
					if g.tryPlaceOreAt(c, candidate, cfg, chunkX, chunkZ, minY, maxY, rng) {
						placedAny = true
					}
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeDisk(c *chunk.Chunk, pos cube.Pos, cfg gen.DiskConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	radius := max(1, g.sampleIntProvider(cfg.Radius, rng))
	var placedAny bool
	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			if dx*dx+dz*dz > radius*radius {
				continue
			}
			for dy := -cfg.HalfHeight; dy <= cfg.HalfHeight; dy++ {
				candidate := pos.Add(cube.Pos{dx, dy, dz})
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}
				if !g.testBlockPredicate(c, candidate, cfg.Target, chunkX, chunkZ, minY, maxY, rng) {
					continue
				}
				state, ok := g.selectState(c, cfg.StateProvider, candidate, rng, minY, maxY)
				if !ok || !g.setBlockStateDirect(c, candidate, state) {
					continue
				}
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeSpringFeature(c *chunk.Chunk, pos cube.Pos, cfg gen.SpringFeatureConfig, chunkX, chunkZ, minY, maxY int, _ *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	if g.blockNameAt(c, pos) != "air" {
		return false
	}

	valid := func(target cube.Pos) bool {
		if !g.positionInChunk(target, chunkX, chunkZ, minY, maxY) {
			return false
		}
		return slices.Contains(cfg.ValidBlocks.Values, g.blockNameAt(c, target))
	}

	if cfg.RequiresBlockBelow && !valid(pos.Side(cube.FaceDown)) {
		return false
	}
	if !valid(pos.Side(cube.FaceUp)) {
		return false
	}

	rocks, holes := 0, 0
	for _, face := range append(cube.HorizontalFaces(), cube.FaceDown) {
		neighbor := pos.Side(face)
		if valid(neighbor) {
			rocks++
			continue
		}
		if g.positionInChunk(neighbor, chunkX, chunkZ, minY, maxY) && g.blockNameAt(c, neighbor) == "air" {
			holes++
		}
	}
	if rocks != cfg.RockCount || holes != cfg.HoleCount {
		return false
	}
	return g.setBlockStateDirect(c, pos, cfg.State)
}

func (g Generator) executeUnderwaterMagma(c *chunk.Chunk, pos cube.Pos, cfg gen.UnderwaterMagmaConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	floorY := -1
	for y := pos[1]; y >= max(minY, pos[1]-cfg.FloorSearchRange); y-- {
		candidate := cube.Pos{pos[0], y, pos[2]}
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) == g.waterRID &&
			g.isSolidRID(c.Block(uint8(candidate[0]&15), int16(candidate[1]-1), uint8(candidate[2]&15), 0)) {
			floorY = y - 1
			break
		}
	}
	if floorY < minY {
		return false
	}

	var placedAny bool
	for dx := -cfg.PlacementRadiusAroundFloor; dx <= cfg.PlacementRadiusAroundFloor; dx++ {
		for dz := -cfg.PlacementRadiusAroundFloor; dz <= cfg.PlacementRadiusAroundFloor; dz++ {
			candidate := cube.Pos{pos[0] + dx, floorY, pos[2] + dz}
			if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) || rng.NextDouble() > cfg.PlacementProbabilityPerValidPosition {
				continue
			}
			above := candidate.Side(cube.FaceUp)
			if c.Block(uint8(above[0]&15), int16(above[1]), uint8(above[2]&15), 0) != g.waterRID {
				continue
			}
			if g.setBlockStateDirect(c, candidate, gen.BlockState{Name: "magma"}) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executePointedDripstone(c *chunk.Chunk, pos cube.Pos, cfg gen.PointedDripstoneConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	currentRID := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
	if currentRID != g.airRID && currentRID != g.waterRID {
		return false
	}

	upSolid := g.isSolidInChunk(c, pos.Side(cube.FaceUp), chunkX, chunkZ, minY, maxY)
	downSolid := g.isSolidInChunk(c, pos.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY)
	var direction string
	switch {
	case upSolid && !downSolid:
		direction = "down"
	case downSolid && !upSolid:
		direction = "up"
	case upSolid:
		direction = "down"
	default:
		return false
	}

	if !g.setBlockStateDirect(c, pos, pointedDripstoneState(direction, "tip")) {
		return false
	}
	if rng.NextDouble() < cfg.ChanceOfTallerDripstone {
		var next cube.Pos
		if direction == "down" {
			next = pos.Side(cube.FaceDown)
		} else {
			next = pos.Side(cube.FaceUp)
		}
		if g.positionInChunk(next, chunkX, chunkZ, minY, maxY) {
			rid := c.Block(uint8(next[0]&15), int16(next[1]), uint8(next[2]&15), 0)
			if rid == g.airRID || rid == g.waterRID {
				_ = g.setBlockStateDirect(c, next, pointedDripstoneState(direction, "base"))
			}
		}
	}
	return true
}

func (g Generator) executeDripstoneCluster(c *chunk.Chunk, pos cube.Pos, cfg gen.DripstoneClusterConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	floor, ceiling, ok := g.findFloorAndCeiling(c, pos, cfg.FloorToCeilingSearchRange, chunkX, chunkZ, minY, maxY)
	if !ok {
		return false
	}
	radius := max(1, g.sampleIntProvider(cfg.Radius, rng))
	thickness := max(1, g.sampleIntProvider(cfg.DripstoneBlockLayerThickness, rng))
	height := max(1, g.sampleIntProvider(cfg.Height, rng))
	var placedAny bool

	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			if dx*dx+dz*dz > radius*radius {
				continue
			}
			for t := 0; t < thickness; t++ {
				if g.setBlockStateDirect(c, floor.Add(cube.Pos{dx, t, dz}), gen.BlockState{Name: "dripstone_block"}) {
					placedAny = true
				}
				if g.setBlockStateDirect(c, ceiling.Add(cube.Pos{dx, -t, dz}), gen.BlockState{Name: "dripstone_block"}) {
					placedAny = true
				}
			}
			base := floor.Add(cube.Pos{dx, thickness, dz})
			top := ceiling.Add(cube.Pos{dx, -thickness, dz})
			for i := 0; i < height && base[1]+i < top[1]; i++ {
				if g.setBlockStateDirect(c, base.Add(cube.Pos{0, i, 0}), pointedDripstoneState("up", "tip")) {
					placedAny = true
				}
				if g.setBlockStateDirect(c, top.Add(cube.Pos{0, -i, 0}), pointedDripstoneState("down", "tip")) {
					placedAny = true
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeSculkPatch(c *chunk.Chunk, pos cube.Pos, cfg gen.SculkPatchConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	attempts := max(1, min(cfg.SpreadAttempts, cfg.ChargeCount*cfg.SpreadRounds*4))
	var placedAny bool
	for i := 0; i < attempts; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(9)) - 4,
			int(rng.NextInt(5)) - 2,
			int(rng.NextInt(9)) - 4,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		floor := candidate.Side(cube.FaceDown)
		if !g.isSolidInChunk(c, floor, chunkX, chunkZ, minY, maxY) {
			continue
		}
		rid := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
		if rid != g.airRID && rid != g.waterRID {
			continue
		}
		if g.setBlockStateDirect(c, floor, gen.BlockState{Name: "sculk"}) {
			placedAny = true
		}
		if rng.NextDouble() < 0.35 {
			_, _ = cfg, rng
			_ = g.executeMultifaceGrowth(c, candidate, gen.MultifaceGrowthConfig{
				Block:             "sculk_vein",
				CanBePlacedOn:     []string{"stone", "andesite", "diorite", "granite", "dripstone_block", "calcite", "tuff", "deepslate", "sculk"},
				CanPlaceOnCeiling: true,
				CanPlaceOnFloor:   true,
				CanPlaceOnWall:    true,
				ChanceOfSpreading: 1.0,
				SearchRange:       4,
			}, chunkX, chunkZ, minY, maxY, rng)
		}
	}
	return placedAny
}

func (g Generator) executeVines(c *chunk.Chunk, pos cube.Pos, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	attachments := []struct {
		face cube.Face
		dir  cube.Direction
	}{
		{cube.FaceNorth, cube.North},
		{cube.FaceEast, cube.East},
		{cube.FaceSouth, cube.South},
		{cube.FaceWest, cube.West},
	}
	var vine block.Vines
	for _, attachment := range attachments {
		support := pos.Side(attachment.face)
		if g.isSolidInChunk(c, support, chunkX, chunkZ, minY, maxY) {
			vine = vine.WithAttachment(attachment.dir.Opposite(), true)
		}
	}
	if len(vine.Attachments()) == 0 {
		return false
	}
	height := 1 + int(rng.NextInt(4))
	var placedAny bool
	for i := 0; i < height && pos[1]-i > minY; i++ {
		current := pos.Add(cube.Pos{0, -i, 0})
		rid := c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0)
		if rid != g.airRID {
			break
		}
		c.SetBlock(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0, world.BlockRuntimeID(vine))
		placedAny = true
	}
	return placedAny
}

func (g Generator) executeSeaPickle(c *chunk.Chunk, pos cube.Pos, cfg gen.SeaPickleConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	if c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0) != g.waterRID {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if !g.isSolidInChunk(c, below, chunkX, chunkZ, minY, maxY) {
		return false
	}
	additional := 0
	if cfg.Count > 1 {
		additional = int(rng.NextInt(uint32(min(cfg.Count, 4))))
	}
	return g.setFeatureBlock(c, pos, block.SeaPickle{AdditionalCount: additional})
}

func (g Generator) executeLake(c *chunk.Chunk, pos cube.Pos, cfg gen.LakeConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || pos[1] <= minY+4 || pos[1] >= maxY-4 {
		return false
	}

	fluid, ok := g.selectState(c, cfg.Fluid, pos, rng, minY, maxY)
	if !ok {
		return false
	}
	barrier, barrierOK := g.selectState(c, cfg.Barrier, pos, rng, minY, maxY)

	radiusX := 2 + int(rng.NextInt(3))
	radiusZ := 2 + int(rng.NextInt(3))
	depth := 2 + int(rng.NextInt(2))
	var placedAny bool

	for dx := -radiusX; dx <= radiusX; dx++ {
		for dz := -radiusZ; dz <= radiusZ; dz++ {
			for dy := -depth; dy <= 1; dy++ {
				nx := float64(dx) / float64(radiusX)
				ny := float64(dy) / float64(depth)
				nz := float64(dz) / float64(radiusZ)
				if nx*nx+ny*ny+nz*nz > 1.0 {
					continue
				}

				candidate := pos.Add(cube.Pos{dx, dy, dz})
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}

				if dy <= 0 {
					if g.setBlockStateDirect(c, candidate, fluid) {
						placedAny = true
					}
					continue
				}
				c.SetBlock(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0, g.airRID)
				placedAny = true
			}
		}
	}

	if !placedAny || !barrierOK {
		return placedAny
	}

	for dx := -radiusX - 1; dx <= radiusX+1; dx++ {
		for dz := -radiusZ - 1; dz <= radiusZ+1; dz++ {
			for dy := -depth - 1; dy <= 0; dy++ {
				candidate := pos.Add(cube.Pos{dx, dy, dz})
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}
				nx := float64(dx) / float64(radiusX+1)
				ny := float64(dy) / float64(depth+1)
				nz := float64(dz) / float64(radiusZ+1)
				outer := nx*nx+ny*ny+nz*nz <= 1.0
				inner := float64(dx*dx)/float64(radiusX*radiusX)+float64(dy*dy)/float64(depth*depth)+float64(dz*dz)/float64(radiusZ*radiusZ) <= 1.0
				if !outer || inner {
					continue
				}
				rid := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
				if rid == g.airRID || rid == g.waterRID || rid == g.lavaRID {
					_ = g.setBlockStateDirect(c, candidate, barrier)
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeFreezeTopLayer(c *chunk.Chunk, biomes sourceBiomeVolume, biomeKey string, _ gen.FreezeTopLayerConfig, chunkX, chunkZ, minY, maxY int) bool {
	if !isFreezingBiomeKey(biomeKey) {
		return false
	}

	var placedAny bool
	for localX := 0; localX < 16; localX++ {
		for localZ := 0; localZ < 16; localZ++ {
			surfaceY := g.heightmapPlacementY(c, localX, localZ, "WORLD_SURFACE", minY, maxY) - 1
			if surfaceY < minY || surfaceY > maxY {
				continue
			}
			if g.sourceBiomeKeyAt(biomes, localX, surfaceY, localZ) != biomeKey {
				continue
			}

			top := cube.Pos{chunkX*16 + localX, surfaceY, chunkZ*16 + localZ}
			topRID := c.Block(uint8(localX), int16(surfaceY), uint8(localZ), 0)
			if topRID == g.waterRID {
				if g.setBlockStateDirect(c, top, gen.BlockState{Name: "ice"}) {
					placedAny = true
				}
				continue
			}
			if !g.isSolidRID(topRID) {
				continue
			}

			above := top.Side(cube.FaceUp)
			if above[1] > maxY {
				continue
			}
			if c.Block(uint8(above[0]&15), int16(above[1]), uint8(above[2]&15), 0) != g.airRID {
				continue
			}
			if g.setBlockStateDirect(c, above, gen.BlockState{Name: "snow"}) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeFallenTree(c *chunk.Chunk, pos cube.Pos, cfg gen.FallenTreeConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	trunkState, ok := g.selectState(c, cfg.TrunkProvider, pos, rng, minY, maxY)
	if !ok {
		return false
	}
	length := max(1, g.sampleIntProvider(cfg.LogLength, rng))

	validDirs := make([]cube.Pos, 0, 4)
	for _, dir := range []cube.Pos{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {
		end := pos.Add(cube.Pos{dir[0] * (length - 1), 0, dir[2] * (length - 1)})
		if g.positionInChunk(end, chunkX, chunkZ, minY, maxY) {
			validDirs = append(validDirs, dir)
		}
	}
	if len(validDirs) == 0 {
		return false
	}
	dir := validDirs[int(rng.NextInt(uint32(len(validDirs))))]
	if trunkState.Properties == nil {
		trunkState.Properties = make(map[string]string, 1)
	}
	if dir[0] != 0 {
		trunkState.Properties["axis"] = "x"
	} else {
		trunkState.Properties["axis"] = "z"
	}

	logBlock, ok := g.featureBlockFromState(trunkState, nil)
	if !ok {
		return false
	}

	var placedAny bool
	trunkPositions := make([]cube.Pos, 0, length)
	for i := 0; i < length; i++ {
		candidate := pos.Add(cube.Pos{dir[0] * i, 0, dir[2] * i})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			break
		}
		if !g.isSolidInChunk(c, candidate.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY) {
			break
		}
		currentRID := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
		currentBlock, _ := world.BlockByRuntimeID(currentRID)
		if !g.canReplaceFeatureBlock(currentBlock, logBlock) {
			break
		}
		if g.setBlockStateDirect(c, candidate, trunkState) {
			placedAny = true
			trunkPositions = append(trunkPositions, candidate)
		}
	}
	if placedAny {
		g.applyAttachedLogDecorators(c, trunkPositions, cfg.LogDecorators, rng, minY, maxY)
	}
	return placedAny
}

func (g Generator) executeTree(c *chunk.Chunk, pos cube.Pos, cfg gen.TreeConfig, minY, maxY int, rng *gen.Xoroshiro128) bool {
	trunkState, ok := g.selectState(c, cfg.TrunkProvider, pos, rng, minY, maxY)
	if !ok {
		return false
	}
	leafState, ok := g.selectState(c, cfg.FoliageProvider, pos, rng, minY, maxY)
	if !ok {
		return false
	}

	height, trunkType := sampleTreeHeight(cfg.TrunkPlacer, rng)
	if height <= 0 {
		return false
	}
	if !g.prepareTreeSoil(c, pos, cfg, rng, minY, maxY) {
		return false
	}

	var trunkTop cube.Pos
	doubleTrunk := false
	switch trunkType {
	case "straight_trunk_placer", "fancy_trunk_placer", "bending_trunk_placer", "cherry_trunk_placer", "upwards_branching_trunk_placer":
		top, ok := g.placeVerticalTrunk(c, pos, trunkState, height, minY, maxY)
		if !ok {
			return false
		}
		trunkTop = top
	case "forking_trunk_placer":
		top, ok := g.placeForkingAcaciaTrunk(c, pos, trunkState, height, rng, minY, maxY)
		if !ok {
			return false
		}
		trunkTop = top
	case "dark_oak_trunk_placer", "giant_trunk_placer", "mega_jungle_trunk_placer":
		top, ok := g.placeWideTrunk(c, pos, trunkState, height, minY, maxY)
		if !ok {
			return false
		}
		trunkTop = top
		doubleTrunk = true
	default:
		return false
	}

	if !g.placeTreeFoliage(c, trunkTop, leafState, cfg.FoliagePlacer, height, doubleTrunk, rng, minY, maxY) {
		return false
	}

	trunkPositions, leafPositions := g.collectTreeStructure(c, pos, trunkTop, height)
	g.applyTreeDecorators(c, pos, trunkPositions, leafPositions, cfg.Decorators, rng, minY, maxY)
	return true
}

func (g Generator) executeBamboo(c *chunk.Chunk, pos cube.Pos, cfg gen.BambooConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || pos[1] <= minY || pos[1] >= maxY {
		return false
	}
	if !g.isSolidInChunk(c, pos.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY) {
		return false
	}

	height := 5 + int(rng.NextInt(8))
	state := gen.BlockState{Name: "sugar_cane", Properties: map[string]string{"age": strconv.Itoa(int(rng.NextInt(16)))}}
	var placedAny bool
	for i := 0; i < height && pos[1]+i <= maxY; i++ {
		candidate := pos.Add(cube.Pos{0, i, 0})
		rid := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
		if rid != g.airRID {
			break
		}
		if g.setBlockStateDirect(c, candidate, state) {
			placedAny = true
		}
	}
	if !placedAny || cfg.Probability <= 0 || rng.NextDouble() >= cfg.Probability {
		return placedAny
	}

	for dx := -2; dx <= 2; dx++ {
		for dz := -2; dz <= 2; dz++ {
			if dx == 0 && dz == 0 {
				continue
			}
			candidate := pos.Add(cube.Pos{dx, -1, dz})
			if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
				continue
			}
			name := g.blockNameAt(c, candidate)
			if name != "dirt" && name != "grass" {
				continue
			}
			_ = g.setBlockStateDirect(c, candidate, gen.BlockState{Name: "podzol", Properties: map[string]string{"snowy": "false"}})
		}
	}
	return true
}

func (g Generator) executeVegetationPatch(c *chunk.Chunk, biomes sourceBiomeVolume, pos cube.Pos, cfg gen.VegetationPatchConfig, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, depth int, waterlogged bool) bool {
	radius := max(1, g.sampleIntProvider(cfg.XZRadius, rng))
	patchDepth := max(1, g.sampleIntProvider(cfg.Depth, rng))
	var placedAny bool

	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			dist2 := dx*dx + dz*dz
			if dist2 > radius*radius {
				continue
			}
			if dist2 >= (radius-1)*(radius-1) && cfg.ExtraEdgeColumnChance > 0 && rng.NextDouble() > cfg.ExtraEdgeColumnChance {
				continue
			}

			basePos, plantPos, ok := g.findVegetationPatchSurface(c, pos.Add(cube.Pos{dx, 0, dz}), cfg.Surface, cfg.VerticalRange, chunkX, chunkZ, minY, maxY)
			if !ok || !g.matchesFeatureBlockTag(g.blockNameAt(c, basePos), cfg.Replaceable) {
				continue
			}

			groundState, ok := g.selectState(c, cfg.GroundState, basePos, rng, minY, maxY)
			if !ok {
				continue
			}
			for d := 0; d < patchDepth; d++ {
				target := basePos
				if strings.EqualFold(cfg.Surface, "ceiling") {
					target = target.Add(cube.Pos{0, d, 0})
				} else {
					target = target.Add(cube.Pos{0, -d, 0})
				}
				if !g.positionInChunk(target, chunkX, chunkZ, minY, maxY) || !g.matchesFeatureBlockTag(g.blockNameAt(c, target), cfg.Replaceable) {
					continue
				}
				if g.setBlockStateDirect(c, target, groundState) {
					placedAny = true
				}
			}

			if cfg.VegetationChance <= 0 || rng.NextDouble() >= cfg.VegetationChance {
				continue
			}
			if waterlogged && c.Block(uint8(plantPos[0]&15), int16(plantPos[1]), uint8(plantPos[2]&15), 0) == g.airRID {
				c.SetBlock(uint8(plantPos[0]&15), int16(plantPos[1]), uint8(plantPos[2]&15), 0, g.waterRID)
			}
			if g.executePlacedFeatureRef(c, biomes, plantPos, cfg.VegetationFeature, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeRootSystem(c *chunk.Chunk, biomes sourceBiomeVolume, pos cube.Pos, cfg gen.RootSystemConfig, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128, depth int) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || !g.testBlockPredicate(c, pos, cfg.AllowedTreePosition, chunkX, chunkZ, minY, maxY, rng) {
		return false
	}

	waterBlocks := 0
	for y := 0; y < max(1, cfg.RequiredVerticalSpaceForTree); y++ {
		current := pos.Add(cube.Pos{0, y, 0})
		if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			return false
		}
		rid := c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0)
		if rid == g.waterRID {
			waterBlocks++
			continue
		}
		if rid != g.airRID {
			return false
		}
	}
	if waterBlocks > cfg.AllowedVerticalWaterForTree {
		return false
	}
	if !g.executePlacedFeatureRef(c, biomes, pos, cfg.Feature, biomeKey, chunkX, chunkZ, minY, maxY, rng, depth+1) {
		return false
	}

	rootState, rootStateOK := g.selectState(c, cfg.RootStateProvider, pos, rng, minY, maxY)
	hangingState, hangingStateOK := g.selectState(c, cfg.HangingRootStateProvider, pos, rng, minY, maxY)
	for i := 0; i < cfg.RootPlacementAttempts; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(uint32(max(1, cfg.RootRadius*2+1)))) - cfg.RootRadius,
			-int(rng.NextInt(uint32(max(1, min(4, cfg.RootColumnMaxHeight)+1)))),
			int(rng.NextInt(uint32(max(1, cfg.RootRadius*2+1)))) - cfg.RootRadius,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) || !rootStateOK {
			continue
		}
		if g.matchesFeatureBlockTag(g.blockNameAt(c, candidate), cfg.RootReplaceable) {
			_ = g.setBlockStateDirect(c, candidate, rootState)
		}
	}
	for i := 0; i < cfg.HangingRootPlacementAttempts; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(uint32(max(1, cfg.HangingRootRadius*2+1)))) - cfg.HangingRootRadius,
			-int(rng.NextInt(uint32(max(1, cfg.HangingRootsVerticalSpan+1)))),
			int(rng.NextInt(uint32(max(1, cfg.HangingRootRadius*2+1)))) - cfg.HangingRootRadius,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) || !hangingStateOK {
			continue
		}
		rid := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
		if rid != g.airRID || !g.isSolidInChunk(c, candidate.Side(cube.FaceUp), chunkX, chunkZ, minY, maxY) {
			continue
		}
		_ = g.setBlockStateDirect(c, candidate, hangingState)
	}
	return true
}

func (g Generator) executeHugeFungus(c *chunk.Chunk, pos cube.Pos, cfg gen.HugeFungusConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || pos[1] <= minY+1 {
		return false
	}

	basePos := pos.Side(cube.FaceDown)
	if !g.positionInChunk(basePos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	if g.blockNameAt(c, basePos) != normalizeFeatureStateName(cfg.ValidBaseBlock.Name) {
		return false
	}
	if !g.testBlockPredicate(c, pos, cfg.ReplaceableBlocks, chunkX, chunkZ, minY, maxY, rng) {
		return false
	}

	height := 5 + int(rng.NextInt(7))
	if cfg.Planted {
		height = max(4, height-2)
	}
	if pos[1]+height > maxY {
		height = maxY - pos[1]
	}
	if height <= 0 {
		return false
	}

	var placedAny bool
	for dy := 0; dy < height; dy++ {
		current := pos.Add(cube.Pos{0, dy, 0})
		if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if dy != 0 && !g.testBlockPredicate(c, current, cfg.ReplaceableBlocks, chunkX, chunkZ, minY, maxY, rng) {
			break
		}
		if g.setBlockStateDirect(c, current, cfg.StemState) {
			placedAny = true
		}
	}

	topY := pos[1] + height - 1
	for y := topY - 3; y <= topY; y++ {
		if y < minY || y > maxY {
			continue
		}
		layer := topY - y
		radius := 2
		if layer == 3 {
			radius = 1
		}
		for dx := -radius; dx <= radius; dx++ {
			for dz := -radius; dz <= radius; dz++ {
				candidate := cube.Pos{pos[0] + dx, y, pos[2] + dz}
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}
				if dx*dx+dz*dz > radius*radius+1 {
					continue
				}
				if !g.testBlockPredicate(c, candidate, cfg.ReplaceableBlocks, chunkX, chunkZ, minY, maxY, rng) {
					continue
				}
				state := cfg.HatState
				if layer <= 1 && (abs(dx) == radius || abs(dz) == radius) && rng.NextDouble() < 0.2 {
					state = cfg.DecorState
				}
				if g.setBlockStateDirect(c, candidate, state) {
					placedAny = true
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeNetherForestVegetation(c *chunk.Chunk, pos cube.Pos, cfg gen.NetherForestVegetationConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	attempts := max(16, cfg.SpreadWidth*cfg.SpreadWidth)
	var placedAny bool
	for i := 0; i < attempts; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(uint32(max(1, cfg.SpreadWidth*2+1)))) - cfg.SpreadWidth,
			int(rng.NextInt(uint32(max(1, cfg.SpreadHeight*2+1)))) - cfg.SpreadHeight,
			int(rng.NextInt(uint32(max(1, cfg.SpreadWidth*2+1)))) - cfg.SpreadWidth,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
			continue
		}
		if !supportsNetherFloraBlock(g.worldBlockAtChunkSafe(c, candidate.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY)) {
			continue
		}
		if g.placeStateProviderBlock(c, candidate, cfg.StateProvider, rng, minY, maxY) {
			placedAny = true
		}
	}
	return placedAny
}

func (g Generator) executeTwistingVines(c *chunk.Chunk, pos cube.Pos, cfg gen.TwistingVinesConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	attempts := max(16, cfg.SpreadWidth*cfg.SpreadWidth)
	var placedAny bool
	for i := 0; i < attempts; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(uint32(max(1, cfg.SpreadWidth*2+1)))) - cfg.SpreadWidth,
			int(rng.NextInt(uint32(max(1, cfg.SpreadHeight*2+1)))) - cfg.SpreadHeight,
			int(rng.NextInt(uint32(max(1, cfg.SpreadWidth*2+1)))) - cfg.SpreadWidth,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
			continue
		}
		if !supportsTwistingVinesBlock(g.worldBlockAtChunkSafe(c, candidate.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY)) {
			continue
		}
		height := 1 + int(rng.NextInt(uint32(max(1, cfg.MaxHeight))))
		for dy := 0; dy < height && candidate[1]+dy <= maxY; dy++ {
			current := candidate.Add(cube.Pos{0, dy, 0})
			if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) || c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
				break
			}
			if g.setBlockStateDirect(c, current, gen.BlockState{Name: "twisting_vines", Properties: map[string]string{"age": strconv.Itoa(int(rng.NextInt(26)))}}) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeWeepingVines(c *chunk.Chunk, pos cube.Pos, _ gen.WeepingVinesConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	var placedAny bool
	for i := 0; i < 32; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(17)) - 8,
			int(rng.NextInt(9)) - 4,
			int(rng.NextInt(17)) - 8,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
			continue
		}
		if !supportsWeepingVinesBlock(g.worldBlockAtChunkSafe(c, candidate.Side(cube.FaceUp), chunkX, chunkZ, minY, maxY)) {
			continue
		}
		height := 1 + int(rng.NextInt(8))
		for dy := 0; dy < height && candidate[1]-dy > minY; dy++ {
			current := candidate.Add(cube.Pos{0, -dy, 0})
			if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) || c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
				break
			}
			if g.setBlockStateDirect(c, current, gen.BlockState{Name: "weeping_vines", Properties: map[string]string{"age": strconv.Itoa(int(rng.NextInt(26)))}}) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeNetherrackReplaceBlobs(c *chunk.Chunk, pos cube.Pos, cfg gen.NetherrackReplaceBlobsConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	radius := max(1, g.sampleIntProvider(cfg.Radius, rng))
	targetName := normalizeFeatureStateName(cfg.Target.Name)
	var placedAny bool
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			for dz := -radius; dz <= radius; dz++ {
				if dx*dx+dy*dy+dz*dz > radius*radius {
					continue
				}
				candidate := pos.Add(cube.Pos{dx, dy, dz})
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}
				if g.blockNameAt(c, candidate) != targetName {
					continue
				}
				if g.setBlockStateDirect(c, candidate, cfg.State) {
					placedAny = true
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeGlowstoneBlob(c *chunk.Chunk, pos cube.Pos, _ gen.GlowstoneBlobConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0) != g.airRID {
		return false
	}
	aboveName := g.blockNameAtSafe(c, pos.Side(cube.FaceUp), chunkX, chunkZ, minY, maxY)
	if aboveName != "netherrack" && aboveName != "basalt" && aboveName != "blackstone" {
		return false
	}

	var placedAny bool
	for i := 0; i < 40; i++ {
		candidate := pos.Add(cube.Pos{
			int(rng.NextInt(9)) - 4,
			-int(rng.NextInt(6)),
			int(rng.NextInt(9)) - 4,
		})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
			continue
		}
		neighbors := 0
		for _, face := range cube.Faces() {
			if g.blockNameAtSafe(c, candidate.Side(face), chunkX, chunkZ, minY, maxY) == "glowstone" {
				neighbors++
			}
		}
		if candidate == pos || neighbors > 0 {
			if g.setBlockStateDirect(c, candidate, gen.BlockState{Name: "glowstone"}) {
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeBasaltPillar(c *chunk.Chunk, pos cube.Pos, _ gen.BasaltPillarConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	start := pos
	for start[1] < maxY && c.Block(uint8(start[0]&15), int16(start[1]), uint8(start[2]&15), 0) == g.airRID {
		start = start.Side(cube.FaceUp)
	}
	if !supportsBasaltAnchorBlock(g.worldBlockAtChunkSafe(c, start, chunkX, chunkZ, minY, maxY)) {
		return false
	}
	start = start.Side(cube.FaceDown)

	var placedAny bool
	height := 2 + int(rng.NextInt(8))
	for dy := 0; dy < height && start[1]-dy >= minY; dy++ {
		current := start.Add(cube.Pos{0, -dy, 0})
		if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
			break
		}
		if g.setBlockStateDirect(c, current, gen.BlockState{Name: "basalt", Properties: map[string]string{"axis": "y"}}) {
			placedAny = true
		}
	}
	return placedAny
}

func (g Generator) executeBasaltColumns(c *chunk.Chunk, pos cube.Pos, cfg gen.BasaltColumnsConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	reach := max(1, g.sampleIntProvider(cfg.Reach, rng))
	height := max(1, g.sampleIntProvider(cfg.Height, rng))
	var placedAny bool
	for dx := -reach; dx <= reach; dx++ {
		for dz := -reach; dz <= reach; dz++ {
			if dx*dx+dz*dz > reach*reach {
				continue
			}
			base := pos.Add(cube.Pos{dx, 0, dz})
			for base[1] > minY && c.Block(uint8(base[0]&15), int16(base[1]), uint8(base[2]&15), 0) == g.airRID {
				base = base.Side(cube.FaceDown)
			}
			if !supportsBasaltAnchorBlock(g.worldBlockAtChunkSafe(c, base, chunkX, chunkZ, minY, maxY)) {
				continue
			}
			for dy := 1; dy <= height && base[1]+dy <= maxY; dy++ {
				current := base.Add(cube.Pos{0, dy, 0})
				if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) || c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
					break
				}
				if g.setBlockStateDirect(c, current, gen.BlockState{Name: "basalt", Properties: map[string]string{"axis": "y"}}) {
					placedAny = true
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeDeltaFeature(c *chunk.Chunk, pos cube.Pos, cfg gen.DeltaFeatureConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	size := max(2, g.sampleIntProvider(cfg.Size, rng))
	rim := max(0, g.sampleIntProvider(cfg.RimSize, rng))
	var placedAny bool
	for dx := -size - rim; dx <= size+rim; dx++ {
		for dz := -size - rim; dz <= size+rim; dz++ {
			candidate := pos.Add(cube.Pos{dx, 0, dz})
			if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
				continue
			}
			dist2 := dx*dx + dz*dz
			if dist2 > (size+rim)*(size+rim) {
				continue
			}
			state := cfg.Rim
			if dist2 <= size*size {
				state = cfg.Contents
			}
			if g.blockNameAt(c, candidate) == "air" || g.blockNameAt(c, candidate) == "lava" || g.blockNameAt(c, candidate) == "netherrack" || g.blockNameAt(c, candidate) == "magma" {
				if g.setBlockStateDirect(c, candidate, state) {
					placedAny = true
				}
			}
		}
	}
	return placedAny
}

func (g Generator) executeChorusPlant(c *chunk.Chunk, pos cube.Pos, _ gen.ChorusPlantConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) || !supportsChorusBlock(g.worldBlockAtChunkSafe(c, pos.Side(cube.FaceDown), chunkX, chunkZ, minY, maxY)) {
		return false
	}
	height := 2 + int(rng.NextInt(4))
	var placedAny bool
	top := pos
	for dy := 0; dy < height && pos[1]+dy <= maxY; dy++ {
		current := pos.Add(cube.Pos{0, dy, 0})
		if c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
			break
		}
		if g.setBlockStateDirect(c, current, gen.BlockState{Name: "chorus_plant"}) {
			placedAny = true
			top = current
		}
	}
	branchCount := 1 + int(rng.NextInt(4))
	for i := 0; i < branchCount; i++ {
		dir := []cube.Pos{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}}[int(rng.NextInt(4))]
		length := 1 + int(rng.NextInt(3))
		current := top
		for step := 0; step < length; step++ {
			current = current.Add(dir)
			if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) || c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
				break
			}
			if g.setBlockStateDirect(c, current, gen.BlockState{Name: "chorus_plant"}) {
				placedAny = true
			}
		}
		flower := current.Side(cube.FaceUp)
		if g.positionInChunk(flower, chunkX, chunkZ, minY, maxY) && c.Block(uint8(flower[0]&15), int16(flower[1]), uint8(flower[2]&15), 0) == g.airRID {
			if g.setBlockStateDirect(c, flower, gen.BlockState{Name: "chorus_flower", Properties: map[string]string{"age": "0"}}) {
				placedAny = true
			}
		}
	}
	if !placedAny {
		return g.setBlockStateDirect(c, top.Side(cube.FaceUp), gen.BlockState{Name: "chorus_flower", Properties: map[string]string{"age": "0"}})
	}
	return true
}

func (g Generator) executeEndIsland(c *chunk.Chunk, pos cube.Pos, _ gen.EndIslandConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	radius := 3 + int(rng.NextInt(4))
	var placedAny bool
	for layer := 0; radius > 0; layer++ {
		y := pos[1] - layer
		if y < minY || y > maxY {
			break
		}
		for dx := -radius; dx <= radius; dx++ {
			for dz := -radius; dz <= radius; dz++ {
				if dx*dx+dz*dz > radius*radius {
					continue
				}
				candidate := cube.Pos{pos[0] + dx, y, pos[2] + dz}
				if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
					continue
				}
				if g.setBlockStateDirect(c, candidate, gen.BlockState{Name: "end_stone"}) {
					placedAny = true
				}
			}
		}
		radius--
	}
	return placedAny
}

func (g Generator) executeEndSpike(c *chunk.Chunk, _ cube.Pos, _ gen.EndSpikeConfig, chunkX, chunkZ, minY, maxY int, _ *gen.Xoroshiro128) bool {
	var placedAny bool
	for _, spike := range endSpikesForSeed(g.seed) {
		if !spikeIntersectsChunk(spike, chunkX, chunkZ) {
			continue
		}
		for x := spike.X - spike.Radius; x <= spike.X+spike.Radius; x++ {
			for z := spike.Z - spike.Radius; z <= spike.Z+spike.Radius; z++ {
				if x>>4 != chunkX || z>>4 != chunkZ {
					continue
				}
				dx, dz := x-spike.X, z-spike.Z
				if dx*dx+dz*dz > spike.Radius*spike.Radius {
					continue
				}
				for y := max(minY, 45); y <= min(maxY, spike.Height); y++ {
					if g.setBlockStateDirect(c, cube.Pos{x, y, z}, gen.BlockState{Name: "obsidian"}) {
						placedAny = true
					}
				}
			}
		}
		top := cube.Pos{spike.X, spike.Height + 1, spike.Z}
		if g.positionInChunk(top, chunkX, chunkZ, minY, maxY) {
			_ = g.setBlockStateDirect(c, top.Side(cube.FaceDown), plainBedrockFeatureState())
			_ = g.setBlockStateDirect(c, top, gen.BlockState{Name: "fire", Properties: map[string]string{"age": "0"}})
			placedAny = true
		}
	}
	return placedAny
}

func (g Generator) executeEndPlatform(c *chunk.Chunk, pos cube.Pos, _ gen.EndPlatformConfig, chunkX, chunkZ, minY, maxY int) bool {
	var placedAny bool
	for x := pos[0] - 2; x <= pos[0]+2; x++ {
		for z := pos[2] - 2; z <= pos[2]+2; z++ {
			floor := cube.Pos{x, pos[1] - 1, z}
			if g.positionInChunk(floor, chunkX, chunkZ, minY, maxY) && g.setBlockStateDirect(c, floor, gen.BlockState{Name: "obsidian"}) {
				placedAny = true
			}
			for y := pos[1]; y <= min(maxY, pos[1]+3); y++ {
				current := cube.Pos{x, y, z}
				if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
					continue
				}
				c.SetBlock(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0, g.airRID)
				placedAny = true
			}
		}
	}
	return placedAny
}

func (g Generator) executeEndGateway(c *chunk.Chunk, pos cube.Pos, _ gen.EndGatewayConfig, chunkX, chunkZ, minY, maxY int) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	var placedAny bool
	for _, offset := range []cube.Pos{{0, 0, 0}, {0, -1, 0}, {0, 1, 0}, {1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {
		current := pos.Add(offset)
		if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			continue
		}
		state := plainBedrockFeatureState()
		if offset == (cube.Pos{}) {
			state = gen.BlockState{Name: "end_gateway"}
		}
		if g.setBlockStateDirect(c, current, state) {
			placedAny = true
		}
	}
	return placedAny
}

type endSpike struct {
	X      int
	Z      int
	Radius int
	Height int
}

func endSpikesForSeed(seed int64) []endSpike {
	rng := gen.NewXoroshiro128FromSeed(seed ^ 0x4f9939f508)
	out := make([]endSpike, 0, 10)
	for i := 0; i < 10; i++ {
		angle := float64(i) * (2 * math.Pi / 10)
		heightClass := int(rng.NextInt(10))
		out = append(out, endSpike{
			X:      int(math.Round(math.Cos(angle) * 42)),
			Z:      int(math.Round(math.Sin(angle) * 42)),
			Radius: 2 + heightClass/3,
			Height: 76 + heightClass*3,
		})
	}
	return out
}

func spikeIntersectsChunk(spike endSpike, chunkX, chunkZ int) bool {
	minX, maxX := chunkX*16, chunkX*16+15
	minZ, maxZ := chunkZ*16, chunkZ*16+15
	return spike.X+spike.Radius >= minX && spike.X-spike.Radius <= maxX &&
		spike.Z+spike.Radius >= minZ && spike.Z-spike.Radius <= maxZ
}

func (g Generator) findVegetationPatchSurface(c *chunk.Chunk, origin cube.Pos, surface string, verticalRange, chunkX, chunkZ, minY, maxY int) (cube.Pos, cube.Pos, bool) {
	for dy := verticalRange; dy >= -verticalRange; dy-- {
		candidate := origin.Add(cube.Pos{0, dy, 0})
		if !g.positionInChunk(candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if strings.EqualFold(surface, "ceiling") {
			if !g.isSolidInChunk(c, candidate, chunkX, chunkZ, minY, maxY) {
				continue
			}
			plantPos := candidate.Side(cube.FaceDown)
			rid := c.Block(uint8(plantPos[0]&15), int16(plantPos[1]), uint8(plantPos[2]&15), 0)
			if rid == g.airRID || rid == g.waterRID {
				return candidate, plantPos, true
			}
			continue
		}
		if !g.isSolidInChunk(c, candidate, chunkX, chunkZ, minY, maxY) {
			continue
		}
		plantPos := candidate.Side(cube.FaceUp)
		rid := c.Block(uint8(plantPos[0]&15), int16(plantPos[1]), uint8(plantPos[2]&15), 0)
		if rid == g.airRID || rid == g.waterRID {
			return candidate, plantPos, true
		}
	}
	return cube.Pos{}, cube.Pos{}, false
}

func (g Generator) collectTreeStructure(c *chunk.Chunk, origin, top cube.Pos, height int) ([]cube.Pos, []cube.Pos) {
	radius := max(6, height/2+4)
	minY := max(c.Range().Min(), origin[1]-2)
	maxY := min(c.Range().Max(), top[1]+4)
	minX := max(origin[0]-radius, origin[0]&^15)
	maxX := min(origin[0]+radius, (origin[0]&^15)+15)
	minZ := max(origin[2]-radius, origin[2]&^15)
	maxZ := min(origin[2]+radius, (origin[2]&^15)+15)

	trunks := make([]cube.Pos, 0, height+8)
	leaves := make([]cube.Pos, 0, height*6)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			for y := minY; y <= maxY; y++ {
				pos := cube.Pos{x, y, z}
				name := g.blockNameAt(c, pos)
				switch {
				case strings.HasSuffix(name, "_log"), strings.HasSuffix(name, "_wood"), strings.HasSuffix(name, "_stem"):
					trunks = append(trunks, pos)
				case strings.HasSuffix(name, "_leaves"):
					leaves = append(leaves, pos)
				}
			}
		}
	}
	return trunks, leaves
}

func (g Generator) applyTreeDecorators(c *chunk.Chunk, origin cube.Pos, trunkPositions, leafPositions []cube.Pos, decorators []gen.FeatureDecorator, rng *gen.Xoroshiro128, minY, maxY int) {
	if len(decorators) == 0 {
		return
	}

	for _, decorator := range decorators {
		switch decorator.Type {
		case "beehive":
			var cfg struct {
				Probability float64 `json:"probability"`
			}
			if err := json.Unmarshal(decorator.Data, &cfg); err == nil {
				g.placeBeeNestDecorator(c, trunkPositions, rng, cfg.Probability)
			}
		case "place_on_ground":
			var cfg struct {
				BlockStateProvider gen.StateProvider `json:"block_state_provider"`
				Height             int               `json:"height"`
				Radius             int               `json:"radius"`
				Tries              int               `json:"tries"`
			}
			if err := json.Unmarshal(decorator.Data, &cfg); err == nil {
				g.placeGroundDecorator(c, origin, cfg.BlockStateProvider, max(1, cfg.Height), max(1, cfg.Radius), max(1, cfg.Tries), rng, minY, maxY)
			}
		case "leave_vine":
			var cfg struct {
				Probability float64 `json:"probability"`
			}
			if err := json.Unmarshal(decorator.Data, &cfg); err == nil {
				g.placeLeafVines(c, leafPositions, rng, cfg.Probability, minY)
			}
		case "trunk_vine":
			g.placeTrunkVines(c, trunkPositions, rng, minY)
		case "attached_to_leaves":
			var cfg struct {
				BlockProvider       gen.StateProvider `json:"block_provider"`
				Directions          []string          `json:"directions"`
				Probability         float64           `json:"probability"`
				RequiredEmptyBlocks int               `json:"required_empty_blocks"`
			}
			if err := json.Unmarshal(decorator.Data, &cfg); err == nil {
				g.placeAttachedToLeaves(c, leafPositions, cfg.BlockProvider, cfg.Directions, cfg.Probability, max(1, cfg.RequiredEmptyBlocks), rng, minY, maxY)
			}
		case "alter_ground":
			var cfg struct {
				Provider gen.StateProvider `json:"provider"`
			}
			if err := json.Unmarshal(decorator.Data, &cfg); err == nil {
				g.alterGroundAroundTree(c, origin, cfg.Provider, rng, minY, maxY)
			}
		}
	}
}

func (g Generator) applyAttachedLogDecorators(c *chunk.Chunk, logPositions []cube.Pos, decorators []gen.FeatureDecorator, rng *gen.Xoroshiro128, minY, maxY int) {
	for _, decorator := range decorators {
		if decorator.Type != "attached_to_logs" {
			continue
		}
		var cfg struct {
			BlockProvider gen.StateProvider `json:"block_provider"`
			Directions    []string          `json:"directions"`
			Probability   float64           `json:"probability"`
		}
		if err := json.Unmarshal(decorator.Data, &cfg); err != nil {
			continue
		}
		for _, logPos := range logPositions {
			for _, direction := range cfg.Directions {
				if cfg.Probability > 0 && rng.NextDouble() >= cfg.Probability {
					continue
				}
				offset := blockColumnDirection(direction)
				candidate := logPos.Add(offset)
				if candidate[1] <= minY || candidate[1] > maxY {
					continue
				}
				if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
					continue
				}
				_ = g.placeStateProviderBlock(c, candidate, cfg.BlockProvider, rng, minY, maxY)
			}
		}
	}
}

func (g Generator) placeBeeNestDecorator(c *chunk.Chunk, trunkPositions []cube.Pos, rng *gen.Xoroshiro128, probability float64) {
	if len(trunkPositions) == 0 || probability <= 0 || rng.NextDouble() >= probability {
		return
	}
	target := trunkPositions[len(trunkPositions)*2/3]
	for _, dir := range []cube.Pos{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {
		candidate := target.Add(dir)
		if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
			continue
		}
		beeNest, ok := world.BlockByName("minecraft:bee_nest", map[string]any{"direction": int32(2), "honey_level": int32(0)})
		if !ok {
			return
		}
		c.SetBlock(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0, world.BlockRuntimeID(beeNest))
		return
	}
}

func (g Generator) placeGroundDecorator(c *chunk.Chunk, origin cube.Pos, provider gen.StateProvider, height, radius, tries int, rng *gen.Xoroshiro128, minY, maxY int) {
	for i := 0; i < tries; i++ {
		x := origin[0] + int(rng.NextInt(uint32(radius*2+1))) - radius
		z := origin[2] + int(rng.NextInt(uint32(radius*2+1))) - radius
		for y := min(maxY, origin[1]+height); y >= max(minY, origin[1]-height); y-- {
			ground := cube.Pos{x, y, z}
			above := ground.Side(cube.FaceUp)
			if !g.isSolidRID(c.Block(uint8(ground[0]&15), int16(ground[1]), uint8(ground[2]&15), 0)) {
				continue
			}
			if c.Block(uint8(above[0]&15), int16(above[1]), uint8(above[2]&15), 0) != g.airRID {
				break
			}
			_ = g.placeStateProviderBlock(c, above, provider, rng, minY, maxY)
			break
		}
	}
}

func (g Generator) placeLeafVines(c *chunk.Chunk, leafPositions []cube.Pos, rng *gen.Xoroshiro128, probability float64, minY int) {
	if probability <= 0 {
		return
	}
	for _, leafPos := range leafPositions {
		for _, face := range cube.HorizontalFaces() {
			if rng.NextDouble() >= probability {
				continue
			}
			candidate := leafPos.Side(face)
			if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
				continue
			}
			g.placeVineColumn(c, candidate, face, minY, 3+int(rng.NextInt(3)))
		}
	}
}

func (g Generator) placeTrunkVines(c *chunk.Chunk, trunkPositions []cube.Pos, rng *gen.Xoroshiro128, minY int) {
	for _, trunkPos := range trunkPositions {
		for _, face := range cube.HorizontalFaces() {
			if rng.NextDouble() >= 0.15 {
				continue
			}
			candidate := trunkPos.Side(face)
			if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID {
				continue
			}
			g.placeVineColumn(c, candidate, face, minY, 1+int(rng.NextInt(2)))
		}
	}
}

func (g Generator) placeVineColumn(c *chunk.Chunk, pos cube.Pos, supportFace cube.Face, minY, length int) {
	vine := block.Vines{}
	switch supportFace {
	case cube.FaceNorth:
		vine = vine.WithAttachment(cube.South, true)
	case cube.FaceSouth:
		vine = vine.WithAttachment(cube.North, true)
	case cube.FaceEast:
		vine = vine.WithAttachment(cube.West, true)
	case cube.FaceWest:
		vine = vine.WithAttachment(cube.East, true)
	default:
		return
	}
	for i := 0; i < length && pos[1]-i > minY; i++ {
		current := pos.Add(cube.Pos{0, -i, 0})
		if c.Block(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0) != g.airRID {
			break
		}
		c.SetBlock(uint8(current[0]&15), int16(current[1]), uint8(current[2]&15), 0, world.BlockRuntimeID(vine))
	}
}

func (g Generator) placeAttachedToLeaves(c *chunk.Chunk, leafPositions []cube.Pos, provider gen.StateProvider, directions []string, probability float64, requiredEmptyBlocks int, rng *gen.Xoroshiro128, minY, maxY int) {
	for _, leafPos := range leafPositions {
		for _, direction := range directions {
			if probability > 0 && rng.NextDouble() >= probability {
				continue
			}
			offset := blockColumnDirection(direction)
			candidate := leafPos.Add(offset)
			if candidate[1] <= minY || candidate[1] > maxY {
				continue
			}
			empty := true
			for i := 0; i < requiredEmptyBlocks; i++ {
				check := candidate.Add(cube.Pos{0, -i, 0})
				rid := c.Block(uint8(check[0]&15), int16(check[1]), uint8(check[2]&15), 0)
				if rid != g.airRID {
					empty = false
					break
				}
			}
			if !empty {
				continue
			}
			state, ok := g.selectState(c, provider, candidate, rng, minY, maxY)
			if !ok {
				continue
			}
			_ = g.setBlockStateDirect(c, candidate, state)
		}
	}
}

func (g Generator) alterGroundAroundTree(c *chunk.Chunk, origin cube.Pos, provider gen.StateProvider, rng *gen.Xoroshiro128, minY, maxY int) {
	for dx := -2; dx <= 2; dx++ {
		for dz := -2; dz <= 2; dz++ {
			x := origin[0] + dx
			z := origin[2] + dz
			localX := x & 15
			localZ := z & 15
			surfaceY := g.heightmapPlacementY(c, localX, localZ, "WORLD_SURFACE", minY, maxY) - 1
			if surfaceY < minY || surfaceY > maxY {
				continue
			}
			surface := cube.Pos{x, surfaceY, z}
			name := g.blockNameAt(c, surface)
			if name != "grass" && name != "dirt" && name != "podzol" && name != "coarse_dirt" {
				continue
			}
			_ = g.placeStateProviderBlock(c, surface, provider, rng, minY, maxY)
		}
	}
}

func (g Generator) applyPlacementModifiers(c *chunk.Chunk, biomes sourceBiomeVolume, positions []cube.Pos, modifiers []gen.PlacementModifier, biomeKey string, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) ([]cube.Pos, bool) {
	out := slices.Clone(positions)

	for _, modifier := range modifiers {
		switch modifier.Type {
		case "count", "count_on_every_layer":
			cfg, err := modifier.Count()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				for i := 0; i < g.sampleIntProvider(cfg.Count, rng); i++ {
					next = append(next, pos)
				}
			}
			out = next
		case "noise_threshold_count":
			cfg, err := modifier.NoiseThresholdCount()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				count := cfg.BelowNoise
				if g.featureCountNoise(pos[0], pos[2]) > cfg.NoiseLevel {
					count = cfg.AboveNoise
				}
				for i := 0; i < count; i++ {
					next = append(next, pos)
				}
			}
			out = next
		case "noise_based_count":
			cfg, err := modifier.NoiseBasedCount()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				count := g.sampleNoiseBasedCount(cfg, pos)
				for i := 0; i < count; i++ {
					next = append(next, pos)
				}
			}
			out = next
		case "rarity_filter":
			cfg, err := modifier.RarityFilter()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				if cfg.Chance <= 1 || rng.NextInt(uint32(cfg.Chance)) == 0 {
					next = append(next, pos)
				}
			}
			out = next
		case "in_square":
			for i, pos := range out {
				pos[0] = chunkX*16 + int(rng.NextInt(16))
				pos[2] = chunkZ*16 + int(rng.NextInt(16))
				out[i] = pos
			}
		case "height_range":
			cfg, err := modifier.HeightRange()
			if err != nil {
				return nil, false
			}
			for i, pos := range out {
				pos[1] = g.sampleHeightProvider(cfg.Height, minY, maxY, rng)
				out[i] = pos
			}
		case "heightmap":
			cfg, err := modifier.Heightmap()
			if err != nil {
				return nil, false
			}
			for i, pos := range out {
				localX := pos[0] - chunkX*16
				localZ := pos[2] - chunkZ*16
				pos[1] = g.heightmapPlacementY(c, localX, localZ, cfg.Heightmap, minY, maxY)
				out[i] = pos
			}
		case "surface_water_depth_filter":
			cfg, err := modifier.SurfaceWaterDepthFilter()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				if g.surfaceWaterDepthAt(c, pos[0]-chunkX*16, pos[2]-chunkZ*16, minY) <= cfg.MaxWaterDepth {
					next = append(next, pos)
				}
			}
			out = next
		case "biome":
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				localX := pos[0] - chunkX*16
				localZ := pos[2] - chunkZ*16
				if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
					continue
				}
				if g.sourceBiomeKeyAt(biomes, localX, pos[1], localZ) == biomeKey {
					next = append(next, pos)
				}
			}
			out = next
		case "random_offset":
			cfg, err := modifier.RandomOffset()
			if err != nil {
				return nil, false
			}
			for i, pos := range out {
				pos[0] += g.sampleIntProvider(cfg.XZSpread, rng)
				pos[1] += g.sampleIntProvider(cfg.YSpread, rng)
				pos[2] += g.sampleIntProvider(cfg.XZSpread, rng)
				out[i] = pos
			}
		case "fixed_placement":
			cfg, err := modifier.FixedPlacement()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out)*max(1, len(cfg.Positions)))
			for range out {
				for _, fixed := range cfg.Positions {
					next = append(next, cube.Pos(fixed))
				}
			}
			out = next
		case "environment_scan":
			cfg, err := modifier.EnvironmentScan()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				if scanned, ok := g.scanEnvironment(c, pos, cfg, chunkX, chunkZ, minY, maxY, rng); ok {
					next = append(next, scanned)
				}
			}
			out = next
		case "surface_relative_threshold_filter":
			cfg, err := modifier.SurfaceRelativeThresholdFilter()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				surfaceY := g.heightmapPlacementY(c, pos[0]-chunkX*16, pos[2]-chunkZ*16, cfg.Heightmap, minY, maxY)
				delta := pos[1] - surfaceY
				if cfg.MinInclusive != nil && delta < *cfg.MinInclusive {
					continue
				}
				if cfg.MaxInclusive != nil && delta > *cfg.MaxInclusive {
					continue
				}
				next = append(next, pos)
			}
			out = next
		case "block_predicate_filter":
			cfg, err := modifier.BlockPredicateFilter()
			if err != nil {
				return nil, false
			}
			next := make([]cube.Pos, 0, len(out))
			for _, pos := range out {
				if g.testBlockPredicate(c, pos, cfg.Predicate, chunkX, chunkZ, minY, maxY, rng) {
					next = append(next, pos)
				}
			}
			out = next
		default:
			return nil, false
		}

		if len(out) == 0 {
			return nil, true
		}
	}
	return out, true
}

func (g Generator) placeStateProviderBlock(c *chunk.Chunk, pos cube.Pos, provider gen.StateProvider, rng *gen.Xoroshiro128, minY, maxY int) bool {
	state, ok := g.selectState(c, provider, pos, rng, minY, maxY)
	if !ok {
		return false
	}
	return g.placeFeatureState(c, pos, state, rng, minY, maxY)
}

func (g Generator) selectState(c *chunk.Chunk, provider gen.StateProvider, pos cube.Pos, rng *gen.Xoroshiro128, minY, maxY int) (gen.BlockState, bool) {
	switch provider.Type {
	case "simple_state_provider":
		cfg, err := provider.SimpleState()
		if err != nil {
			return gen.BlockState{}, false
		}
		return cfg.State, true
	case "weighted_state_provider":
		cfg, err := provider.WeightedState()
		if err != nil || len(cfg.Entries) == 0 {
			return gen.BlockState{}, false
		}
		total := 0
		for _, entry := range cfg.Entries {
			total += entry.Weight
		}
		if total <= 0 {
			return gen.BlockState{}, false
		}
		pick := int(rng.NextInt(uint32(total)))
		for _, entry := range cfg.Entries {
			pick -= entry.Weight
			if pick < 0 {
				return entry.Data, true
			}
		}
		return cfg.Entries[len(cfg.Entries)-1].Data, true
	case "randomized_int_state_provider":
		cfg, err := provider.RandomizedIntState()
		if err != nil {
			return gen.BlockState{}, false
		}
		state, ok := g.selectState(c, cfg.Source, pos, rng, minY, maxY)
		if !ok {
			return gen.BlockState{}, false
		}
		if state.Properties == nil {
			state.Properties = make(map[string]string, 1)
		}
		state.Properties[cfg.Property] = strconv.Itoa(g.sampleIntProvider(cfg.Values, rng))
		return state, true
	case "rule_based_state_provider":
		cfg, err := provider.RuleBasedState()
		if err != nil {
			return gen.BlockState{}, false
		}
		for _, rule := range cfg.Rules {
			if g.testBlockPredicate(c, pos, rule.IfTrue, pos[0]>>4, pos[2]>>4, minY, maxY, rng) {
				return g.selectState(c, rule.Then, pos, rng, minY, maxY)
			}
		}
		return g.selectState(c, cfg.Fallback, pos, rng, minY, maxY)
	case "noise_threshold_provider":
		cfg, err := provider.NoiseThreshold()
		if err != nil {
			return gen.BlockState{}, false
		}
		value := g.noiseThresholdProviderValue(provider, cfg, pos)
		if value < cfg.Threshold && len(cfg.LowStates) > 0 {
			return cfg.LowStates[int(rng.NextInt(uint32(len(cfg.LowStates))))], true
		}
		if len(cfg.HighStates) > 0 && rng.NextDouble() < cfg.HighChance {
			return cfg.HighStates[int(rng.NextInt(uint32(len(cfg.HighStates))))], true
		}
		return cfg.DefaultState, true
	default:
		return gen.BlockState{}, false
	}
}

func (g Generator) placeFeatureState(c *chunk.Chunk, pos cube.Pos, state gen.BlockState, rng *gen.Xoroshiro128, minY, maxY int) bool {
	featureBlock, ok := g.featureBlockFromState(state, rng)
	if !ok || pos[1] <= minY || pos[1] > maxY {
		return false
	}

	localX := uint8(pos[0] & 15)
	localZ := uint8(pos[2] & 15)
	currentRID := c.Block(localX, int16(pos[1]), localZ, 0)
	currentBlock, _ := world.BlockByRuntimeID(currentRID)
	if !g.canReplaceFeatureBlock(currentBlock, featureBlock) {
		return false
	}

	if !g.canBlockStateSurvive(c, pos, state, rng, minY, maxY) {
		return false
	}

	g.setFeatureBlock(c, pos, featureBlock)

	if tall, ok := featureBlock.(block.DoubleTallGrass); ok && !tall.UpperPart {
		upperPos := pos.Side(cube.FaceUp)
		if upperPos[1] > maxY {
			return false
		}
		upperRID := c.Block(uint8(upperPos[0]&15), int16(upperPos[1]), uint8(upperPos[2]&15), 0)
		upperBlock, _ := world.BlockByRuntimeID(upperRID)
		if !g.canReplaceFeatureBlock(upperBlock, block.DoubleTallGrass{Type: tall.Type, UpperPart: true}) {
			return false
		}
		g.setFeatureBlock(c, upperPos, block.DoubleTallGrass{Type: tall.Type, UpperPart: true})
	}

	return true
}

func (g Generator) featureBlockFromState(state gen.BlockState, rng *gen.Xoroshiro128) (world.Block, bool) {
	state = normalizeFeatureState(state)

	switch state.Name {
	case "tall_grass":
		upper := state.Properties["half"] == "upper"
		return block.DoubleTallGrass{Type: block.NormalDoubleTallGrass(), UpperPart: upper}, true
	case "large_fern":
		upper := state.Properties["half"] == "upper"
		return block.DoubleTallGrass{Type: block.FernDoubleTallGrass(), UpperPart: upper}, true
	case "sugar_cane":
		return block.SugarCane{Age: parseStateInt(state.Properties, "age")}, true
	case "cactus":
		return block.Cactus{Age: parseStateInt(state.Properties, "age")}, true
	case "kelp":
		return block.Kelp{Age: parseStateInt(state.Properties, "age")}, true
	case "water":
		return block.Water{Depth: 8, Falling: state.Properties["falling"] == "true"}, true
	case "lava":
		return block.Lava{Depth: 8, Falling: state.Properties["falling"] == "true"}, true
	case "pumpkin":
		facing := cube.Direction(0)
		if rng != nil {
			facing = cube.Direction(rng.NextInt(4))
		}
		return block.Pumpkin{Facing: facing}, true
	}

	props := featureBlockProperties(state.Properties)
	name := state.Name
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	featureBlock, ok := world.BlockByName(name, props)
	if ok {
		return featureBlock, true
	}

	fallbackState, ok := dragonflyFallbackFeatureState(state)
	if !ok {
		return nil, false
	}
	return g.featureBlockFromState(fallbackState, rng)
}

func featureBlockProperties(properties map[string]string) map[string]any {
	if len(properties) == 0 {
		return nil
	}

	out := make(map[string]any, len(properties))
	for key, value := range properties {
		switch value {
		case "true":
			out[key] = true
		case "false":
			out[key] = false
		default:
			if n, err := strconv.ParseInt(value, 10, 32); err == nil {
				out[key] = int32(n)
			} else {
				out[key] = value
			}
		}
	}
	return out
}

func parseStateInt(properties map[string]string, key string) int {
	if properties == nil {
		return 0
	}
	value, ok := properties[key]
	if !ok {
		return 0
	}
	n, _ := strconv.Atoi(value)
	return n
}

func normalizeFeatureState(state gen.BlockState) gen.BlockState {
	state.Name = normalizeFeatureStateName(state.Name)
	if len(state.Properties) == 0 {
		return state
	}

	props := make(map[string]string, len(state.Properties))
	for key, value := range state.Properties {
		props[key] = value
	}

	switch {
	case strings.HasSuffix(state.Name, "_log"),
		strings.HasSuffix(state.Name, "_wood"),
		strings.HasSuffix(state.Name, "_stem"),
		strings.HasSuffix(state.Name, "_hyphae"),
		state.Name == "muddy_mangrove_roots",
		state.Name == "basalt",
		state.Name == "deepslate":
		renameFeatureProperty(props, "axis", "pillar_axis")
	case strings.HasSuffix(state.Name, "_leaves"),
		state.Name == "azalea_leaves",
		state.Name == "azalea_leaves_flowered":
		renameFeatureProperty(props, "persistent", "persistent_bit")
		delete(props, "distance")
		delete(props, "waterlogged")
		if _, ok := props["update_bit"]; !ok {
			props["update_bit"] = "false"
		}
	}
	if state.Name == "hanging_roots" {
		delete(props, "waterlogged")
	}

	if len(props) == 0 {
		state.Properties = nil
	} else {
		state.Properties = props
	}
	return state
}

func normalizeFeatureStateName(name string) string {
	name = strings.TrimPrefix(name, "minecraft:")
	switch name {
	case "lily_pad":
		return "waterlily"
	case "snow_block":
		return "snow"
	case "nether_quartz_ore":
		return "quartz_ore"
	case "flowering_azalea_leaves":
		return "azalea_leaves_flowered"
	default:
		return name
	}
}

func dragonflyFallbackFeatureState(state gen.BlockState) (gen.BlockState, bool) {
	name := strings.TrimPrefix(state.Name, "minecraft:")
	switch name {
	case "bamboo":
		// TODO: Place real bamboo when Dragonfly exposes minecraft:bamboo.
		return gen.BlockState{Name: "sugar_cane"}, true
	case "rooted_dirt":
		// TODO: Place real rooted dirt when Dragonfly exposes minecraft:rooted_dirt.
		return gen.BlockState{Name: "dirt"}, true
	case "leaf_litter":
		// TODO: Place real leaf litter when Dragonfly exposes minecraft:leaf_litter.
		return gen.BlockState{Name: "short_grass"}, true
	case "mangrove_propagule":
		// TODO: Place real mangrove propagules when Dragonfly exposes minecraft:mangrove_propagule.
		return gen.BlockState{Name: "oak_sapling"}, true
	case "azalea":
		// TODO: Place real azalea shrubs when Dragonfly exposes minecraft:azalea.
		return gen.BlockState{Name: "oak_sapling"}, true
	case "flowering_azalea":
		// TODO: Place real flowering azalea shrubs when Dragonfly exposes minecraft:flowering_azalea.
		return gen.BlockState{Name: "oak_sapling"}, true
	case "big_dripleaf", "big_dripleaf_stem":
		// TODO: Place real big dripleaf blocks when Dragonfly exposes minecraft:big_dripleaf.
		return gen.BlockState{Name: "sugar_cane"}, true
	case "small_dripleaf":
		// TODO: Place real small dripleaf blocks when Dragonfly exposes minecraft:small_dripleaf.
		return gen.BlockState{Name: "short_grass"}, true
	default:
		return state, false
	}
}

func renameFeatureProperty(properties map[string]string, from, to string) {
	value, ok := properties[from]
	if !ok {
		return
	}
	delete(properties, from)
	if _, exists := properties[to]; !exists {
		properties[to] = value
	}
}

func (g Generator) blockEncodedName(b world.Block) string {
	name, _ := b.EncodeBlock()
	return strings.TrimPrefix(name, "minecraft:")
}

func (g Generator) canSaplingSurviveOn(belowBlock world.Block, stateName string) bool {
	switch belowBlock.(type) {
	case block.Dirt, block.Grass, block.Podzol, block.Farmland:
		return true
	case block.Mud, block.MuddyMangroveRoots:
		return stateName == "mangrove_propagule"
	default:
		return false
	}
}

func isFreezingBiomeKey(biomeKey string) bool {
	switch biomeKey {
	case "frozen_ocean", "deep_frozen_ocean", "frozen_river", "snowy_beach", "snowy_plains", "snowy_taiga", "ice_spikes", "grove", "snowy_slopes", "frozen_peaks", "jagged_peaks":
		return true
	default:
		return strings.Contains(biomeKey, "snowy") || strings.Contains(biomeKey, "frozen")
	}
}

func (g Generator) canReplaceFeatureBlock(current, with world.Block) bool {
	if current == nil {
		return true
	}
	if _, ok := current.(block.Air); ok {
		return true
	}
	if _, ok := current.(block.Water); ok {
		_, submerged := with.(block.Kelp)
		return submerged || strings.Contains(g.blockEncodedName(with), "seagrass") || strings.Contains(g.blockEncodedName(with), "sea_pickle")
	}
	replaceable, ok := current.(block.Replaceable)
	return ok && replaceable.ReplaceableBy(with)
}

func (g Generator) canBlockStateSurvive(c *chunk.Chunk, pos cube.Pos, state gen.BlockState, rng *gen.Xoroshiro128, minY, maxY int) bool {
	stateName := normalizeFeatureStateName(state.Name)
	if g.canNamedFeatureStateSurvive(c, pos, stateName, minY, maxY) {
		return true
	}

	featureBlock, ok := g.featureBlockFromState(state, rng)
	if !ok {
		return false
	}
	return g.canFeatureBlockSurvive(c, pos, featureBlock, state.Name, minY, maxY)
}

func (g Generator) canNamedFeatureStateSurvive(c *chunk.Chunk, pos cube.Pos, stateName string, minY, maxY int) bool {
	if pos[1] <= minY {
		return false
	}

	belowRID := c.Block(uint8(pos[0]&15), int16(pos[1]-1), uint8(pos[2]&15), 0)
	belowBlock, _ := world.BlockByRuntimeID(belowRID)

	switch stateName {
	case "oak_sapling", "spruce_sapling", "birch_sapling", "jungle_sapling", "acacia_sapling", "dark_oak_sapling", "cherry_sapling", "pale_oak_sapling", "mangrove_propagule", "azalea", "flowering_azalea":
		return g.canSaplingSurviveOn(belowBlock, stateName)
	case "sweet_berry_bush", "brown_mushroom", "red_mushroom", "firefly_bush":
		return belowRID != g.airRID && belowRID != g.waterRID && belowRID != g.lavaRID
	case "warped_fungus", "crimson_fungus":
		return supportsNetherFloraBlock(belowBlock)
	case "warped_roots", "crimson_roots":
		return supportsNetherRootsBlock(belowBlock)
	case "twisting_vines":
		return supportsTwistingVinesBlock(belowBlock)
	case "weeping_vines":
		aboveRID := c.Block(uint8(pos[0]&15), int16(pos[1]+1), uint8(pos[2]&15), 0)
		aboveBlock, _ := world.BlockByRuntimeID(aboveRID)
		return supportsWeepingVinesBlock(aboveBlock)
	case "chorus_plant", "chorus_flower":
		return supportsChorusBlock(belowBlock)
	case "seagrass", "tall_seagrass", "sea_pickle":
		currentRID := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
		return currentRID == g.waterRID && g.isSolidRID(belowRID)
	case "lily_pad":
		return belowRID == g.waterRID
	default:
		return false
	}
}

func (g Generator) canFeatureBlockSurvive(c *chunk.Chunk, pos cube.Pos, featureBlock world.Block, stateName string, minY, maxY int) bool {
	if pos[1] <= minY {
		return false
	}

	stateName = normalizeFeatureStateName(stateName)
	if g.canNamedFeatureStateSurvive(c, pos, stateName, minY, maxY) {
		return true
	}

	belowRID := c.Block(uint8(pos[0]&15), int16(pos[1]-1), uint8(pos[2]&15), 0)
	belowBlock, _ := world.BlockByRuntimeID(belowRID)
	switch featureBlock := featureBlock.(type) {
	case block.ShortGrass, block.DoubleTallGrass, block.Flower:
		soil, ok := belowBlock.(block.Soil)
		return ok && soil.SoilFor(featureBlock)
	case block.Fungus:
		return supportsNetherFloraBlock(belowBlock)
	case block.Roots:
		return supportsNetherRootsBlock(belowBlock)
	case block.NetherVines:
		if featureBlock.Twisting {
			return supportsTwistingVinesBlock(belowBlock)
		}
		aboveRID := c.Block(uint8(pos[0]&15), int16(pos[1]+1), uint8(pos[2]&15), 0)
		aboveBlock, _ := world.BlockByRuntimeID(aboveRID)
		return supportsWeepingVinesBlock(aboveBlock)
	case block.ChorusPlant, block.ChorusFlower:
		return supportsChorusBlock(belowBlock)
	case block.SugarCane:
		if !g.positionInChunk(pos, pos[0]>>4, pos[2]>>4, minY, maxY) {
			return false
		}
		if _, ok := belowBlock.(block.SugarCane); ok {
			return true
		}
		for _, face := range cube.HorizontalFaces() {
			side := pos.Side(face).Side(cube.FaceDown)
			if !g.positionInChunk(side, pos[0]>>4, pos[2]>>4, minY, maxY) {
				continue
			}
			if rid := c.Block(uint8(side[0]&15), int16(side[1]), uint8(side[2]&15), 0); rid == g.waterRID {
				soil, ok := belowBlock.(block.Soil)
				return ok && soil.SoilFor(featureBlock)
			}
		}
		return false
	case block.Cactus:
		for _, face := range cube.HorizontalFaces() {
			side := pos.Side(face)
			if !g.positionInChunk(side, pos[0]>>4, pos[2]>>4, minY, maxY) {
				continue
			}
			sideRID := c.Block(uint8(side[0]&15), int16(side[1]), uint8(side[2]&15), 0)
			if sideRID != g.airRID {
				return false
			}
		}
		soil, ok := belowBlock.(block.Soil)
		return ok && soil.SoilFor(featureBlock)
	case block.Kelp:
		currentRID := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
		if currentRID != g.waterRID {
			return false
		}
		if _, ok := belowBlock.(block.Kelp); ok {
			return true
		}
		return g.isSolidRID(belowRID)
	case block.Pumpkin:
		return belowRID != g.airRID && belowRID != g.waterRID && belowRID != g.lavaRID
	default:
		return false
	}
}

func (g Generator) testBlockPredicate(c *chunk.Chunk, pos cube.Pos, predicate gen.BlockPredicate, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	switch predicate.Type {
	case "matching_blocks":
		cfg, err := predicate.MatchingBlocks()
		if err != nil {
			return false
		}
		target := pos.Add(cube.Pos(cfg.Offset))
		if !g.positionInChunk(target, chunkX, chunkZ, minY, maxY) {
			return false
		}
		name := g.blockNameAt(c, target)
		return slices.Contains(cfg.Blocks.Values, name)
	case "matching_fluids":
		cfg, err := predicate.MatchingFluids()
		if err != nil {
			return false
		}
		target := pos.Add(cube.Pos(cfg.Offset))
		if !g.positionInChunk(target, chunkX, chunkZ, minY, maxY) {
			return false
		}
		rid := c.Block(uint8(target[0]&15), int16(target[1]), uint8(target[2]&15), 0)
		fluid := g.blockNameAt(c, target)
		if rid == g.waterRID && slices.Contains(cfg.Fluids.Values, "flowing_water") {
			return true
		}
		return slices.Contains(cfg.Fluids.Values, fluid)
	case "matching_block_tag":
		cfg, err := predicate.MatchingBlockTag()
		if err != nil {
			return false
		}
		target := pos.Add(cube.Pos(cfg.Offset))
		if !g.positionInChunk(target, chunkX, chunkZ, minY, maxY) {
			return false
		}
		return g.matchesFeatureBlockTag(g.blockNameAt(c, target), cfg.Tag)
	case "solid":
		rid := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
		return g.isSolidRID(rid)
	case "all_of":
		var raw struct {
			Predicates []gen.BlockPredicate `json:"predicates"`
		}
		if err := json.Unmarshal(predicate.Data, &raw); err != nil {
			return false
		}
		for _, child := range raw.Predicates {
			if !g.testBlockPredicate(c, pos, child, chunkX, chunkZ, minY, maxY, rng) {
				return false
			}
		}
		return true
	case "any_of":
		var raw struct {
			Predicates []gen.BlockPredicate `json:"predicates"`
		}
		if err := json.Unmarshal(predicate.Data, &raw); err != nil {
			return false
		}
		for _, child := range raw.Predicates {
			if g.testBlockPredicate(c, pos, child, chunkX, chunkZ, minY, maxY, rng) {
				return true
			}
		}
		return false
	case "not":
		cfg, err := predicate.Not()
		if err != nil {
			return false
		}
		return !g.testBlockPredicate(c, pos, cfg.Predicate, chunkX, chunkZ, minY, maxY, rng)
	case "would_survive":
		cfg, err := predicate.WouldSurvive()
		if err != nil {
			return false
		}
		return g.canBlockStateSurvive(c, pos, cfg.State, rng, minY, maxY)
	case "inside_world_bounds":
		var raw struct {
			Offset gen.BlockPos `json:"offset"`
		}
		if err := json.Unmarshal(predicate.Data, &raw); err != nil {
			return false
		}
		target := pos.Add(cube.Pos(raw.Offset))
		return target[1] >= minY && target[1] <= maxY
	default:
		return false
	}
}

func (g Generator) blockNameAt(c *chunk.Chunk, pos cube.Pos) string {
	rid := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
	featureBlock, ok := world.BlockByRuntimeID(rid)
	if !ok {
		return "air"
	}
	name, _ := featureBlock.EncodeBlock()
	return strings.TrimPrefix(name, "minecraft:")
}

func (g Generator) blockNameAtSafe(c *chunk.Chunk, pos cube.Pos, chunkX, chunkZ, minY, maxY int) string {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return "air"
	}
	return g.blockNameAt(c, pos)
}

func worldBlockAtChunk(c *chunk.Chunk, pos cube.Pos) world.Block {
	rid := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
	b, _ := world.BlockByRuntimeID(rid)
	return b
}

func (g Generator) worldBlockAtChunkSafe(c *chunk.Chunk, pos cube.Pos, chunkX, chunkZ, minY, maxY int) world.Block {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return nil
	}
	return worldBlockAtChunk(c, pos)
}

func supportsNetherFloraBlock(b world.Block) bool {
	switch b.(type) {
	case block.Nylium:
		return true
	default:
		return false
	}
}

func supportsNetherRootsBlock(b world.Block) bool {
	switch b.(type) {
	case block.Nylium, block.SoulSoil:
		return true
	default:
		return false
	}
}

func supportsTwistingVinesBlock(b world.Block) bool {
	switch b.(type) {
	case block.Netherrack, block.Nylium, block.NetherWartBlock, block.Blackstone:
		return true
	default:
		return false
	}
}

func supportsWeepingVinesBlock(b world.Block) bool {
	switch b.(type) {
	case block.Netherrack, block.NetherWartBlock, block.Wood, block.Log:
		return true
	default:
		return false
	}
}

func supportsBasaltAnchorBlock(b world.Block) bool {
	switch b.(type) {
	case block.Netherrack, block.Basalt, block.Blackstone, block.SoulSoil, block.SoulSand:
		return true
	default:
		name, _ := b.EncodeBlock()
		name = strings.TrimPrefix(name, "minecraft:")
		return name == "magma" || name == "magma_block"
	}
}

func supportsChorusBlock(b world.Block) bool {
	switch b.(type) {
	case block.EndStone, block.ChorusPlant:
		return true
	default:
		return false
	}
}

func (g Generator) matchesFeatureBlockTag(blockName, tag string) bool {
	tag = normalizeFeatureTag(tag)
	switch tag {
	case "replaceable_by_trees":
		return blockName == "air" || blockName == "short_grass" || blockName == "tall_grass" || blockName == "fern" || blockName == "large_fern" || strings.HasSuffix(blockName, "_mushroom") || strings.Contains(blockName, "flower") || blockName == "waterlily" || blockName == "snow"
	case "azalea_grows_on":
		return slices.Contains([]string{"dirt", "grass", "clay", "moss_block", "podzol"}, blockName)
	case "moss_replaceable":
		return slices.Contains([]string{"stone", "granite", "diorite", "andesite", "tuff", "deepslate", "calcite", "dripstone_block", "clay", "dirt", "grass", "podzol", "mud"}, blockName)
	case "lush_ground_replaceable":
		return slices.Contains([]string{"stone", "granite", "diorite", "andesite", "tuff", "deepslate", "calcite", "dripstone_block", "clay", "dirt", "grass", "podzol", "moss_block"}, blockName)
	case "azalea_root_replaceable":
		return slices.Contains([]string{"stone", "granite", "diorite", "andesite", "tuff", "deepslate", "calcite", "dripstone_block", "clay", "dirt", "grass", "podzol", "moss_block"}, blockName)
	case "mangrove_roots_can_grow_through":
		return blockName == "air" || blockName == "water" || blockName == "mud" || blockName == "short_grass" || blockName == "tall_grass" || strings.HasSuffix(blockName, "_leaves")
	case "mangrove_logs_can_grow_through":
		return blockName == "air" || blockName == "water" || blockName == "short_grass" || blockName == "tall_grass" || strings.HasSuffix(blockName, "_leaves")
	default:
		return false
	}
}

func normalizeFeatureTag(tag string) string {
	tag = strings.TrimPrefix(tag, "#")
	return strings.TrimPrefix(tag, "minecraft:")
}

func (g Generator) biomeKeyAt(c *chunk.Chunk, localX, y, localZ int) string {
	return biomeKey(biomeFromRuntimeID(c.Biome(uint8(localX), int16(y), uint8(localZ))))
}

func (g Generator) sourceBiomeKeyAt(biomes sourceBiomeVolume, localX, y, localZ int) string {
	return biomeKey(biomes.biomeAt(localX, y, localZ))
}

func (g Generator) heightmapPlacementY(c *chunk.Chunk, localX, localZ int, kind string, minY, maxY int) int {
	switch kind {
	case "WORLD_SURFACE_WG", "WORLD_SURFACE", "MOTION_BLOCKING", "MOTION_BLOCKING_NO_LEAVES":
		return g.columnHeightmapY(c, localX, localZ, kind, minY, maxY)
	case "OCEAN_FLOOR", "OCEAN_FLOOR_WG":
		return g.columnHeightmapY(c, localX, localZ, kind, minY, maxY)
	default:
		return g.columnHeightmapY(c, localX, localZ, "WORLD_SURFACE", minY, maxY)
	}
}

func (g Generator) surfaceWaterDepthAt(c *chunk.Chunk, localX, localZ, minY int) int {
	maxY := c.Range().Max()
	worldSurface := g.heightmapPlacementY(c, localX, localZ, "WORLD_SURFACE", minY, maxY)
	oceanFloor := g.heightmapPlacementY(c, localX, localZ, "OCEAN_FLOOR_WG", minY, maxY)
	if worldSurface <= oceanFloor {
		return 0
	}
	return worldSurface - oceanFloor
}

func (g Generator) columnHeightmapY(c *chunk.Chunk, localX, localZ int, kind string, minY, maxY int) int {
	topY := int(c.HighestBlock(uint8(localX), uint8(localZ)))
	if topY < minY {
		return minY
	}
	if topY > maxY {
		topY = maxY
	}

	switch kind {
	case "WORLD_SURFACE_WG", "WORLD_SURFACE":
		return min(topY+1, maxY)
	case "MOTION_BLOCKING":
		for y := topY; y >= minY; y-- {
			if rid := g.columnScanRuntimeID(c, localX, y, localZ); g.isMotionBlockingRID(rid, false) {
				return min(y+1, maxY)
			}
		}
		return minY
	case "MOTION_BLOCKING_NO_LEAVES":
		for y := topY; y >= minY; y-- {
			if rid := g.columnScanRuntimeID(c, localX, y, localZ); g.isMotionBlockingRID(rid, true) {
				return min(y+1, maxY)
			}
		}
		return minY
	case "OCEAN_FLOOR", "OCEAN_FLOOR_WG":
		for y := topY; y >= minY; y-- {
			rid := g.columnScanRuntimeID(c, localX, y, localZ)
			if rid == g.airRID || rid == g.waterRID || rid == g.lavaRID {
				continue
			}
			if g.isSolidRID(rid) {
				return min(y+1, maxY)
			}
		}
		return minY
	default:
		return min(topY+1, maxY)
	}
}

func (g Generator) columnScanRuntimeID(c *chunk.Chunk, localX, y, localZ int) uint32 {
	rid := c.Block(uint8(localX), int16(y), uint8(localZ), 0)
	if rid != g.airRID {
		return rid
	}
	return c.Block(uint8(localX), int16(y), uint8(localZ), 1)
}

func (g Generator) isMotionBlockingRID(rid uint32, ignoreLeaves bool) bool {
	if rid == g.airRID {
		return false
	}
	if rid == g.waterRID || rid == g.lavaRID {
		return true
	}
	if g.isLeafRID(rid) {
		return !ignoreLeaves
	}
	return g.isSolidRID(rid)
}

func (g Generator) isLeafRID(rid uint32) bool {
	if rid == g.airRID {
		return false
	}
	b, ok := world.BlockByRuntimeID(rid)
	if !ok {
		return false
	}
	name, _ := b.EncodeBlock()
	return strings.HasSuffix(strings.TrimPrefix(name, "minecraft:"), "_leaves")
}

func (g Generator) sampleNoiseBasedCount(cfg gen.NoiseBasedCountPlacement, pos cube.Pos) int {
	noise := g.surface.SurfaceSecondary(int(float64(pos[0])/cfg.NoiseFactor), int(float64(pos[2])/cfg.NoiseFactor))*2.0 - 1.0
	count := int(math.Ceil((noise + cfg.NoiseOffset) * float64(cfg.NoiseToCountRatio)))
	if count < 0 {
		return 0
	}
	return count
}

func (g Generator) sampleHeightProvider(provider gen.HeightProvider, minY, maxY int, rng *gen.Xoroshiro128) int {
	low := clamp(g.anchorY(provider.MinInclusive, minY, maxY), minY, maxY)
	high := clamp(g.anchorY(provider.MaxInclusive, minY, maxY), minY, maxY)
	if high < low {
		low, high = high, low
	}
	switch provider.Kind {
	case "uniform":
		if high <= low {
			return low
		}
		return low + int(rng.NextInt(uint32(high-low+1)))
	case "trapezoid":
		if high <= low {
			return low
		}
		span := high - low
		return low + int(math.Round((rng.NextDouble()+rng.NextDouble())*float64(span)/2.0))
	case "biased_to_bottom":
		if high <= low {
			return low
		}
		width := high - low + 1
		return low + int(rng.NextInt(uint32(max(1, int(rng.NextInt(uint32(width))+1)))))
	case "very_biased_to_bottom":
		if high <= low {
			return low
		}
		width := high - low + 1
		return low + int(rng.NextInt(uint32(max(1, int(rng.NextInt(uint32(max(1, int(rng.NextInt(uint32(width))+1))))+1)))))
	case "clamped_normal":
		return clamp(int(math.Round(g.normalFloat64(rng, provider.Mean, provider.Deviation))), low, high)
	default:
		return low
	}
}

func (g Generator) anchorY(anchor gen.VerticalAnchor, minY, maxY int) int {
	switch anchor.Kind {
	case "absolute":
		return anchor.Value
	case "above_bottom":
		return minY + anchor.Value
	case "below_top":
		return maxY - anchor.Value
	default:
		return minY
	}
}

func (g Generator) scanEnvironment(c *chunk.Chunk, pos cube.Pos, cfg gen.EnvironmentScanPlacement, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) (cube.Pos, bool) {
	dir := blockColumnDirection(cfg.DirectionOfSearch)
	if dir == (cube.Pos{}) {
		return cube.Pos{}, false
	}
	current := pos
	for step := 0; step <= cfg.MaxSteps; step++ {
		if !g.positionInChunk(current, chunkX, chunkZ, minY, maxY) {
			return cube.Pos{}, false
		}
		if cfg.AllowedSearchCondition != nil && !g.testBlockPredicate(c, current, *cfg.AllowedSearchCondition, chunkX, chunkZ, minY, maxY, rng) {
			return cube.Pos{}, false
		}
		if g.testBlockPredicate(c, current, cfg.TargetCondition, chunkX, chunkZ, minY, maxY, rng) {
			return current, true
		}
		current = current.Add(dir)
	}
	return cube.Pos{}, false
}

func blockColumnDirection(direction string) cube.Pos {
	switch strings.ToLower(direction) {
	case "up":
		return cube.Pos{0, 1, 0}
	case "down":
		return cube.Pos{0, -1, 0}
	case "north":
		return cube.Pos{0, 0, -1}
	case "south":
		return cube.Pos{0, 0, 1}
	case "east":
		return cube.Pos{1, 0, 0}
	case "west":
		return cube.Pos{-1, 0, 0}
	default:
		return cube.Pos{}
	}
}

func (g Generator) setBlockStateDirect(c *chunk.Chunk, pos cube.Pos, state gen.BlockState) bool {
	featureBlock, ok := g.featureBlockFromState(state, nil)
	if !ok {
		return false
	}
	return g.setFeatureBlock(c, pos, featureBlock)
}

func (g Generator) setFeatureBlock(c *chunk.Chunk, pos cube.Pos, featureBlock world.Block) bool {
	localX := uint8(pos[0] & 15)
	localZ := uint8(pos[2] & 15)
	y := int16(pos[1])

	liquidRID, displaced := g.displacedLiquidRuntimeID(c, pos, featureBlock)

	c.SetBlock(localX, y, localZ, 0, world.BlockRuntimeID(featureBlock))
	if displaced {
		c.SetBlock(localX, y, localZ, 1, liquidRID)
	} else {
		c.SetBlock(localX, y, localZ, 1, g.airRID)
	}
	return true
}

func (g Generator) displacedLiquidRuntimeID(c *chunk.Chunk, pos cube.Pos, featureBlock world.Block) (uint32, bool) {
	displacer, ok := featureBlock.(world.LiquidDisplacer)
	if !ok {
		return 0, false
	}

	localX := uint8(pos[0] & 15)
	localZ := uint8(pos[2] & 15)
	y := int16(pos[1])

	for _, layer := range [...]uint8{1, 0} {
		rid := c.Block(localX, y, localZ, layer)
		if rid == g.airRID {
			continue
		}
		placed, ok := world.BlockByRuntimeID(rid)
		if !ok {
			continue
		}
		liquid, ok := placed.(world.Liquid)
		if ok && displacer.CanDisplace(liquid) {
			return rid, true
		}
	}
	return 0, false
}

func (g Generator) tryPlaceOreAt(c *chunk.Chunk, pos cube.Pos, cfg gen.OreConfig, chunkX, chunkZ, minY, maxY int, rng *gen.Xoroshiro128) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	currentName := g.blockNameAt(c, pos)
	targetState, ok := g.matchOreTarget(currentName, cfg.Targets)
	if !ok {
		return false
	}
	if cfg.DiscardChanceOnAirExposure > 0 && g.isExposedToAir(c, pos, chunkX, chunkZ, minY, maxY) && rng.NextDouble() < cfg.DiscardChanceOnAirExposure {
		return false
	}
	return g.setBlockStateDirect(c, pos, targetState)
}

func (g Generator) matchOreTarget(blockName string, targets []gen.OreTargetConfig) (gen.BlockState, bool) {
	for _, target := range targets {
		switch target.Target.PredicateType {
		case "tag_match":
			if g.matchesOreTag(blockName, target.Target.Tag) {
				return target.State, true
			}
		case "block_match":
			if blockName == target.Target.Block {
				return target.State, true
			}
		default:
			if g.matchesOreTag(blockName, target.Target.Tag) {
				return target.State, true
			}
		}
		if target.Target.Block != "" && blockName == target.Target.Block {
			return target.State, true
		}
	}
	return gen.BlockState{}, false
}

func (g Generator) matchesOreTag(blockName, tag string) bool {
	switch tag {
	case "stone_ore_replaceables":
		return slices.Contains([]string{"stone", "granite", "diorite", "andesite", "tuff"}, blockName)
	case "deepslate_ore_replaceables":
		return blockName == "deepslate"
	case "base_stone_overworld":
		return slices.Contains([]string{"stone", "granite", "diorite", "andesite", "tuff", "deepslate"}, blockName)
	case "base_stone_nether":
		return slices.Contains([]string{"netherrack", "basalt", "blackstone"}, blockName)
	default:
		return false
	}
}

func (g Generator) isExposedToAir(c *chunk.Chunk, pos cube.Pos, chunkX, chunkZ, minY, maxY int) bool {
	for _, face := range cube.Faces() {
		neighbor := pos.Side(face)
		if !g.positionInChunk(neighbor, chunkX, chunkZ, minY, maxY) {
			continue
		}
		rid := c.Block(uint8(neighbor[0]&15), int16(neighbor[1]), uint8(neighbor[2]&15), 0)
		if rid == g.airRID {
			return true
		}
	}
	return false
}

func sampleTreeHeight(placer gen.TypedJSONValue, rng *gen.Xoroshiro128) (int, string) {
	var raw struct {
		BaseHeight  int `json:"base_height"`
		HeightRandA int `json:"height_rand_a"`
		HeightRandB int `json:"height_rand_b"`
	}
	if err := json.Unmarshal(placer.Data, &raw); err != nil {
		return 0, placer.Type
	}
	height := raw.BaseHeight
	if raw.HeightRandA > 0 {
		height += int(rng.NextInt(uint32(raw.HeightRandA + 1)))
	}
	if raw.HeightRandB > 0 {
		height += int(rng.NextInt(uint32(raw.HeightRandB + 1)))
	}
	return height, placer.Type
}

func (g Generator) prepareTreeSoil(c *chunk.Chunk, pos cube.Pos, cfg gen.TreeConfig, rng *gen.Xoroshiro128, minY, maxY int) bool {
	if pos[1] <= minY || pos[1] > maxY {
		return false
	}
	below := pos.Side(cube.FaceDown)
	belowRID := c.Block(uint8(below[0]&15), int16(below[1]), uint8(below[2]&15), 0)
	if belowRID != g.airRID && belowRID != g.waterRID && belowRID != g.lavaRID && !cfg.ForceDirt {
		return true
	}
	dirt, ok := g.selectState(c, cfg.DirtProvider, below, rng, minY, maxY)
	if !ok {
		return false
	}
	return g.setBlockStateDirect(c, below, dirt)
}

func (g Generator) placeVerticalTrunk(c *chunk.Chunk, pos cube.Pos, trunk gen.BlockState, height, minY, maxY int) (cube.Pos, bool) {
	current := pos
	for i := 0; i < height; i++ {
		if current[1] <= minY || current[1] > maxY {
			return cube.Pos{}, false
		}
		if !g.setBlockStateDirect(c, current, trunk) {
			return cube.Pos{}, false
		}
		current = current.Side(cube.FaceUp)
	}
	return current.Side(cube.FaceDown), true
}

func (g Generator) placeWideTrunk(c *chunk.Chunk, pos cube.Pos, trunk gen.BlockState, height, minY, maxY int) (cube.Pos, bool) {
	if pos[0]&15 == 15 || pos[2]&15 == 15 {
		return cube.Pos{}, false
	}
	currentY := pos[1]
	for i := 0; i < height; i++ {
		if currentY <= minY || currentY > maxY {
			return cube.Pos{}, false
		}
		for dx := 0; dx < 2; dx++ {
			for dz := 0; dz < 2; dz++ {
				if !g.setBlockStateDirect(c, cube.Pos{pos[0] + dx, currentY, pos[2] + dz}, trunk) {
					return cube.Pos{}, false
				}
			}
		}
		currentY++
	}
	return cube.Pos{pos[0], currentY - 1, pos[2]}, true
}

func (g Generator) placeForkingAcaciaTrunk(c *chunk.Chunk, pos cube.Pos, trunk gen.BlockState, height int, rng *gen.Xoroshiro128, minY, maxY int) (cube.Pos, bool) {
	top, ok := g.placeVerticalTrunk(c, pos, trunk, max(2, height-1), minY, maxY)
	if !ok {
		return cube.Pos{}, false
	}
	branchDir := []cube.Pos{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}}[rng.NextInt(4)]
	branch := top
	for i := 0; i < 2; i++ {
		branch = branch.Add(branchDir).Side(cube.FaceUp)
		if branch[1] > maxY || !g.setBlockStateDirect(c, branch, trunk) {
			break
		}
	}
	return branch, true
}

func (g Generator) placeTreeFoliage(c *chunk.Chunk, top cube.Pos, leaf gen.BlockState, placer gen.TypedJSONValue, height int, doubleTrunk bool, rng *gen.Xoroshiro128, minY, maxY int) bool {
	switch placer.Type {
	case "blob_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
			Height int             `json:"height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		for yo := offset; yo >= offset-raw.Height; yo-- {
			currentRadius := max(leafRadius-1-yo/2, 0)
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, blobFoliageSkip, nil)
		}
		return true
	case "fancy_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
			Height int             `json:"height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		for yo := offset; yo >= offset-raw.Height; yo-- {
			currentRadius := leafRadius
			if yo != offset && yo != offset-raw.Height {
				currentRadius++
			}
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, fancyFoliageSkip, nil)
		}
		return true
	case "bush_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
			Height int             `json:"height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		for yo := offset; yo >= offset-raw.Height; yo-- {
			currentRadius := max(leafRadius-1-yo, 0)
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, bushFoliageSkip, nil)
		}
		return true
	case "spruce_foliage_placer":
		var raw struct {
			Radius      gen.IntProvider `json:"radius"`
			Offset      gen.IntProvider `json:"offset"`
			TrunkHeight gen.IntProvider `json:"trunk_height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		foliageHeight := max(4, height-g.sampleIntProvider(raw.TrunkHeight, rng))
		currentRadius := int(rng.NextInt(2))
		maxRadius := 1
		minRadius := 0
		for yo := offset; yo >= -foliageHeight; yo-- {
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, coniferFoliageSkip, nil)
			if currentRadius >= maxRadius {
				currentRadius = minRadius
				minRadius = 1
				maxRadius = min(maxRadius+1, leafRadius)
			} else {
				currentRadius++
			}
		}
		return true
	case "pine_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
			Height gen.IntProvider `json:"height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		if span := max(height+1, 1); span > 1 {
			leafRadius += int(rng.NextInt(uint32(span)))
		}
		offset := g.sampleIntProvider(raw.Offset, rng)
		foliageHeight := max(0, g.sampleIntProvider(raw.Height, rng))
		currentRadius := 0
		for yo := offset; yo >= offset-foliageHeight; yo-- {
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, coniferFoliageSkip, nil)
			if currentRadius >= 1 && yo == offset-foliageHeight+1 {
				currentRadius--
			} else if currentRadius < leafRadius {
				currentRadius++
			}
		}
		return true
	case "acacia_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		foliagePos := top.Add(cube.Pos{0, g.sampleIntProvider(raw.Offset, rng), 0})
		g.placeTreeLeafRow(c, foliagePos, leafRadius, -1, doubleTrunk, leaf, minY, maxY, rng, acaciaFoliageSkip, nil)
		g.placeTreeLeafRow(c, foliagePos, max(leafRadius-1, 0), 0, doubleTrunk, leaf, minY, maxY, rng, acaciaFoliageSkip, nil)
		g.placeTreeLeafRow(c, foliagePos, max(leafRadius-1, 0), 0, doubleTrunk, leaf, minY, maxY, rng, acaciaFoliageSkip, nil)
		return true
	case "dark_oak_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		foliagePos := top.Add(cube.Pos{0, g.sampleIntProvider(raw.Offset, rng), 0})
		if doubleTrunk {
			g.placeTreeLeafRow(c, foliagePos, leafRadius+2, -1, true, leaf, minY, maxY, rng, darkOakFoliageSkip, darkOakSignedSkip)
			g.placeTreeLeafRow(c, foliagePos, leafRadius+3, 0, true, leaf, minY, maxY, rng, darkOakFoliageSkip, darkOakSignedSkip)
			g.placeTreeLeafRow(c, foliagePos, leafRadius+2, 1, true, leaf, minY, maxY, rng, darkOakFoliageSkip, darkOakSignedSkip)
			if rng.NextInt(2) == 0 {
				g.placeTreeLeafRow(c, foliagePos, leafRadius, 2, true, leaf, minY, maxY, rng, darkOakFoliageSkip, darkOakSignedSkip)
			}
		} else {
			g.placeTreeLeafRow(c, foliagePos, leafRadius+2, -1, false, leaf, minY, maxY, rng, darkOakFoliageSkip, nil)
			g.placeTreeLeafRow(c, foliagePos, leafRadius+1, 0, false, leaf, minY, maxY, rng, darkOakFoliageSkip, nil)
		}
		return true
	case "random_spread_foliage_placer":
		var raw struct {
			Radius                gen.IntProvider `json:"radius"`
			Offset                gen.IntProvider `json:"offset"`
			FoliageHeight         gen.IntProvider `json:"foliage_height"`
			LeafPlacementAttempts int             `json:"leaf_placement_attempts"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(1, g.sampleIntProvider(raw.Radius, rng))
		foliageHeight := max(1, g.sampleIntProvider(raw.FoliageHeight, rng))
		attempts := raw.LeafPlacementAttempts
		if attempts <= 0 {
			attempts = max(24, height*8)
		}
		origin := top.Add(cube.Pos{0, g.sampleIntProvider(raw.Offset, rng), 0})
		for i := 0; i < attempts; i++ {
			candidate := origin.Add(cube.Pos{
				int(rng.NextInt(uint32(leafRadius))) - int(rng.NextInt(uint32(leafRadius))),
				int(rng.NextInt(uint32(foliageHeight))) - int(rng.NextInt(uint32(foliageHeight))),
				int(rng.NextInt(uint32(leafRadius))) - int(rng.NextInt(uint32(leafRadius))),
			})
			if candidate[1] <= minY || candidate[1] > maxY {
				continue
			}
			currentRID := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
			if currentRID == g.airRID {
				_ = g.setBlockStateDirect(c, candidate, leaf)
			}
		}
		return true
	case "cherry_foliage_placer":
		var raw struct {
			Radius                       gen.IntProvider `json:"radius"`
			Offset                       gen.IntProvider `json:"offset"`
			Height                       gen.IntProvider `json:"height"`
			WideBottomLayerHoleChance    float64         `json:"wide_bottom_layer_hole_chance"`
			CornerHoleChance             float64         `json:"corner_hole_chance"`
			HangingLeavesChance          float64         `json:"hanging_leaves_chance"`
			HangingLeavesExtensionChance float64         `json:"hanging_leaves_extension_chance"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		foliageHeight := max(0, g.sampleIntProvider(raw.Height, rng))
		foliagePos := top.Add(cube.Pos{0, g.sampleIntProvider(raw.Offset, rng), 0})
		currentRadius := max(leafRadius-1, 0)
		g.placeTreeLeafRow(c, foliagePos, max(currentRadius-2, 0), foliageHeight-3, doubleTrunk, leaf, minY, maxY, rng, cherryFoliageSkip(raw.WideBottomLayerHoleChance, raw.CornerHoleChance), nil)
		g.placeTreeLeafRow(c, foliagePos, max(currentRadius-1, 0), foliageHeight-4, doubleTrunk, leaf, minY, maxY, rng, cherryFoliageSkip(raw.WideBottomLayerHoleChance, raw.CornerHoleChance), nil)
		for y := foliageHeight - 5; y >= 0; y-- {
			g.placeTreeLeafRow(c, foliagePos, currentRadius, y, doubleTrunk, leaf, minY, maxY, rng, cherryFoliageSkip(raw.WideBottomLayerHoleChance, raw.CornerHoleChance), nil)
		}
		g.placeTreeLeafRowWithHangingBelow(c, foliagePos, currentRadius, -1, doubleTrunk, leaf, minY, maxY, rng, cherryFoliageSkip(raw.WideBottomLayerHoleChance, raw.CornerHoleChance), raw.HangingLeavesChance, raw.HangingLeavesExtensionChance)
		g.placeTreeLeafRowWithHangingBelow(c, foliagePos, max(currentRadius-1, 0), -2, doubleTrunk, leaf, minY, maxY, rng, cherryFoliageSkip(raw.WideBottomLayerHoleChance, raw.CornerHoleChance), raw.HangingLeavesChance, raw.HangingLeavesExtensionChance)
		return true
	case "jungle_foliage_placer":
		var raw struct {
			Radius gen.IntProvider `json:"radius"`
			Offset gen.IntProvider `json:"offset"`
			Height int             `json:"height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		leafHeight := 1 + int(rng.NextInt(2))
		if doubleTrunk {
			leafHeight = raw.Height
		}
		for yo := offset; yo >= offset-leafHeight; yo-- {
			currentRadius := max(leafRadius+1-yo, 0)
			g.placeTreeLeafRow(c, top, currentRadius, yo, doubleTrunk, leaf, minY, maxY, rng, megaFoliageSkip, nil)
		}
		return true
	case "mega_pine_foliage_placer":
		var raw struct {
			Radius      gen.IntProvider `json:"radius"`
			Offset      gen.IntProvider `json:"offset"`
			CrownHeight gen.IntProvider `json:"crown_height"`
		}
		if err := json.Unmarshal(placer.Data, &raw); err != nil {
			return false
		}
		leafRadius := max(0, g.sampleIntProvider(raw.Radius, rng))
		offset := g.sampleIntProvider(raw.Offset, rng)
		foliageHeight := max(1, g.sampleIntProvider(raw.CrownHeight, rng))
		prevRadius := 0
		for yy := top[1] - foliageHeight + offset; yy <= top[1]+offset; yy++ {
			yo := top[1] - yy
			smoothRadius := leafRadius + int(math.Floor(float64(yo)/float64(foliageHeight)*3.5))
			jaggedRadius := smoothRadius
			if yo > 0 && smoothRadius == prevRadius && (yy&1) == 0 {
				jaggedRadius++
			}
			g.placeTreeLeafRow(c, cube.Pos{top[0], yy, top[2]}, jaggedRadius, 0, doubleTrunk, leaf, minY, maxY, rng, megaFoliageSkip, nil)
			prevRadius = smoothRadius
		}
		return true
	default:
		_ = height
		_ = rng
		return false
	}
}

type treeFoliageSkip func(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool

func (g Generator) placeTreeLeafRow(c *chunk.Chunk, center cube.Pos, currentRadius, y int, doubleTrunk bool, leaf gen.BlockState, minY, maxY int, rng *gen.Xoroshiro128, skip, signedSkip treeFoliageSkip) {
	if currentRadius < 0 {
		return
	}
	offset := 0
	if doubleTrunk {
		offset = 1
	}
	for dx := -currentRadius; dx <= currentRadius+offset; dx++ {
		for dz := -currentRadius; dz <= currentRadius+offset; dz++ {
			if signedSkip != nil && signedSkip(rng, dx, y, dz, currentRadius, doubleTrunk) {
				continue
			}
			minDx, minDz := abs(dx), abs(dz)
			if doubleTrunk {
				minDx = min(abs(dx), abs(dx-1))
				minDz = min(abs(dz), abs(dz-1))
			}
			if skip != nil && skip(rng, minDx, y, minDz, currentRadius, doubleTrunk) {
				continue
			}
			candidate := center.Add(cube.Pos{dx, y, dz})
			if candidate[1] <= minY || candidate[1] > maxY {
				continue
			}
			currentRID := c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0)
			if currentRID != g.airRID {
				continue
			}
			_ = g.setBlockStateDirect(c, candidate, leaf)
		}
	}
}

func (g Generator) placeTreeLeafRowWithHangingBelow(c *chunk.Chunk, center cube.Pos, currentRadius, y int, doubleTrunk bool, leaf gen.BlockState, minY, maxY int, rng *gen.Xoroshiro128, skip treeFoliageSkip, hangingChance, extensionChance float64) {
	g.placeTreeLeafRow(c, center, currentRadius, y, doubleTrunk, leaf, minY, maxY, rng, skip, nil)
	offset := 0
	if doubleTrunk {
		offset = 1
	}
	for dx := -currentRadius; dx <= currentRadius+offset; dx++ {
		for dz := -currentRadius; dz <= currentRadius+offset; dz++ {
			if abs(dx) != currentRadius && abs(dz) != currentRadius && (!doubleTrunk || (dx != currentRadius+offset && dz != currentRadius+offset)) {
				continue
			}
			candidate := center.Add(cube.Pos{dx, y - 1, dz})
			if candidate[1] <= minY || candidate[1] > maxY {
				continue
			}
			above := candidate.Side(cube.FaceUp)
			if !g.isSameTreeLeaf(c, above, leaf) || !g.positionInChunk(candidate, int(center[0])>>4, int(center[2])>>4, minY, maxY) {
				continue
			}
			if c.Block(uint8(candidate[0]&15), int16(candidate[1]), uint8(candidate[2]&15), 0) != g.airRID || rng.NextDouble() > hangingChance {
				continue
			}
			_ = g.setBlockStateDirect(c, candidate, leaf)
			extension := candidate.Side(cube.FaceDown)
			if extension[1] > minY && extension[1] <= maxY && c.Block(uint8(extension[0]&15), int16(extension[1]), uint8(extension[2]&15), 0) == g.airRID && rng.NextDouble() <= extensionChance {
				_ = g.setBlockStateDirect(c, extension, leaf)
			}
		}
	}
}

func (g Generator) isSameTreeLeaf(c *chunk.Chunk, pos cube.Pos, leaf gen.BlockState) bool {
	if pos[1] <= c.Range().Min() || pos[1] > c.Range().Max() {
		return false
	}
	rid := c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0)
	if rid == g.airRID {
		return false
	}
	b, ok := world.BlockByRuntimeID(rid)
	if !ok {
		return false
	}
	name, _ := b.EncodeBlock()
	return strings.TrimPrefix(name, "minecraft:") == strings.TrimPrefix(leaf.Name, "minecraft:")
}

func blobFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	return dx == currentRadius && dz == currentRadius && (rng.NextInt(2) == 0 || y == 0)
}

func fancyFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	return (float64(dx)+0.5)*(float64(dx)+0.5)+(float64(dz)+0.5)*(float64(dz)+0.5) > float64(currentRadius*currentRadius)
}

func bushFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	return dx == currentRadius && dz == currentRadius && rng.NextInt(2) == 0
}

func coniferFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	return dx == currentRadius && dz == currentRadius && currentRadius > 0
}

func acaciaFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	if y == 0 {
		return (dx > 1 || dz > 1) && dx != 0 && dz != 0
	}
	return dx == currentRadius && dz == currentRadius && currentRadius > 0
}

func darkOakSignedSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	if y != 0 || !doubleTrunk {
		return false
	}
	return (dx == -currentRadius || dx >= currentRadius) && (dz == -currentRadius || dz >= currentRadius)
}

func darkOakFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	if y == -1 && !doubleTrunk {
		return dx == currentRadius && dz == currentRadius
	}
	if y == 1 {
		return dx+dz > currentRadius*2-2
	}
	return false
}

func megaFoliageSkip(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
	return dx+dz >= 7 || dx*dx+dz*dz > currentRadius*currentRadius
}

func cherryFoliageSkip(wideBottomLayerHoleChance, cornerHoleChance float64) treeFoliageSkip {
	return func(rng *gen.Xoroshiro128, dx, y, dz, currentRadius int, doubleTrunk bool) bool {
		if y == -1 && (dx == currentRadius || dz == currentRadius) && rng.NextDouble() < wideBottomLayerHoleChance {
			return true
		}
		corner := dx == currentRadius && dz == currentRadius
		wideLayer := currentRadius > 2
		if wideLayer {
			return corner || (dx+dz > currentRadius*2-2 && rng.NextDouble() < cornerHoleChance)
		}
		return corner && rng.NextDouble() < cornerHoleChance
	}
}

func (g Generator) multifaceStateAt(c *chunk.Chunk, pos cube.Pos, cfg gen.MultifaceGrowthConfig, chunkX, chunkZ, minY, maxY int) (gen.BlockState, bool) {
	type faceProp struct {
		face cube.Face
		key  string
	}
	faces := []faceProp{
		{cube.FaceUp, "down"},
		{cube.FaceDown, "up"},
		{cube.FaceNorth, "south"},
		{cube.FaceSouth, "north"},
		{cube.FaceEast, "west"},
		{cube.FaceWest, "east"},
	}
	props := map[string]string{
		"down":        "false",
		"up":          "false",
		"north":       "false",
		"south":       "false",
		"east":        "false",
		"west":        "false",
		"waterlogged": "false",
	}
	for _, entry := range faces {
		switch entry.face {
		case cube.FaceUp:
			if !cfg.CanPlaceOnFloor {
				continue
			}
		case cube.FaceDown:
			if !cfg.CanPlaceOnCeiling {
				continue
			}
		default:
			if !cfg.CanPlaceOnWall {
				continue
			}
		}
		support := pos.Side(entry.face)
		if !g.positionInChunk(support, chunkX, chunkZ, minY, maxY) {
			continue
		}
		if slices.Contains(cfg.CanBePlacedOn, g.blockNameAt(c, support)) {
			props[entry.key] = "true"
			return gen.BlockState{Name: strings.TrimPrefix(cfg.Block, "minecraft:"), Properties: props}, true
		}
	}
	return gen.BlockState{}, false
}

func (g Generator) isSolidInChunk(c *chunk.Chunk, pos cube.Pos, chunkX, chunkZ, minY, maxY int) bool {
	if !g.positionInChunk(pos, chunkX, chunkZ, minY, maxY) {
		return false
	}
	return g.isSolidRID(c.Block(uint8(pos[0]&15), int16(pos[1]), uint8(pos[2]&15), 0))
}

func (g Generator) findFloorAndCeiling(c *chunk.Chunk, pos cube.Pos, searchRange, chunkX, chunkZ, minY, maxY int) (cube.Pos, cube.Pos, bool) {
	floor, ceiling := cube.Pos{}, cube.Pos{}
	foundFloor, foundCeiling := false, false
	for y := pos[1]; y >= max(minY, pos[1]-searchRange); y-- {
		candidate := cube.Pos{pos[0], y, pos[2]}
		if g.isSolidInChunk(c, candidate, chunkX, chunkZ, minY, maxY) {
			floor = candidate
			foundFloor = true
			break
		}
	}
	for y := pos[1]; y <= min(maxY, pos[1]+searchRange); y++ {
		candidate := cube.Pos{pos[0], y, pos[2]}
		if g.isSolidInChunk(c, candidate, chunkX, chunkZ, minY, maxY) {
			ceiling = candidate
			foundCeiling = true
			break
		}
	}
	return floor, ceiling, foundFloor && foundCeiling && ceiling[1]-floor[1] > 2
}

func pointedDripstoneState(direction, thickness string) gen.BlockState {
	return gen.BlockState{
		Name: "pointed_dripstone",
		Properties: map[string]string{
			"thickness":          thickness,
			"vertical_direction": direction,
			"waterlogged":        "false",
		},
	}
}

func (g Generator) featureCountNoise(x, z int) float64 {
	return g.surface.SurfaceSecondary(x, z)*2.0 - 1.0
}

func (g Generator) noiseThresholdProviderValue(provider gen.StateProvider, cfg gen.NoiseThresholdStateProviderConfig, pos cube.Pos) float64 {
	key := string(provider.Data)
	noise, ok := g.featureNoiseCache.Lookup(key)
	if !ok {
		rng := gen.NewXoroshiro128FromSeed(cfg.Seed)
		noise = gen.NewDoublePerlinNoise(&rng, cfg.Noise.Amplitudes, cfg.Noise.FirstOctave)
		g.featureNoiseCache.Store(key, noise)
	}
	return noise.Sample(float64(pos[0])*cfg.Scale, 0.0, float64(pos[2])*cfg.Scale)
}

func (g Generator) sampleIntProvider(provider gen.IntProvider, rng *gen.Xoroshiro128) int {
	switch provider.Kind {
	case "constant":
		if provider.Constant != nil {
			return *provider.Constant
		}
	case "uniform":
		if provider.MaxInclusive <= provider.MinInclusive {
			return provider.MinInclusive
		}
		return provider.MinInclusive + int(rng.NextInt(uint32(provider.MaxInclusive-provider.MinInclusive+1)))
	case "biased_to_bottom":
		if provider.MaxInclusive <= provider.MinInclusive {
			return provider.MinInclusive
		}
		span := provider.MaxInclusive - provider.MinInclusive + 1
		return provider.MinInclusive + int(rng.NextInt(uint32(max(1, int(rng.NextInt(uint32(span))+1)))))
	case "weighted_list":
		total := 0
		for _, entry := range provider.Distribution {
			total += entry.Weight
		}
		if total <= 0 {
			return 0
		}
		pick := int(rng.NextInt(uint32(total)))
		for _, entry := range provider.Distribution {
			pick -= entry.Weight
			if pick < 0 {
				return entry.Data
			}
		}
	case "clamped":
		if provider.Source == nil {
			return provider.MinInclusive
		}
		return clamp(g.sampleIntProvider(*provider.Source, rng), provider.MinInclusive, provider.MaxInclusive)
	case "clamped_normal":
		return clamp(int(math.Round(g.normalFloat64(rng, provider.Mean, provider.Deviation))), provider.MinInclusive, provider.MaxInclusive)
	}
	return 0
}

func (g Generator) normalFloat64(rng *gen.Xoroshiro128, mean, deviation float64) float64 {
	u1 := rng.NextDouble()
	if u1 <= 0 {
		u1 = math.SmallestNonzeroFloat64
	}
	u2 := rng.NextDouble()
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return mean + z0*deviation
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func lerp(t, a, b float64) float64 {
	return a + (b-a)*t
}

func (g Generator) signedSpread(rng *gen.Xoroshiro128, spread int) int {
	if spread <= 0 {
		return 0
	}
	return int(rng.NextInt(uint32(spread*2+1))) - spread
}

func (g Generator) positionInChunk(pos cube.Pos, chunkX, chunkZ, minY, maxY int) bool {
	return pos[0] >= chunkX*16 && pos[0] < chunkX*16+16 &&
		pos[2] >= chunkZ*16 && pos[2] < chunkZ*16+16 &&
		pos[1] >= minY && pos[1] <= maxY
}

func (g Generator) featureRNG(chunkX, chunkZ int, biomeKey, featureName string) gen.Xoroshiro128 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(biomeKey))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(featureName))
	seed := int64(h.Sum64()) ^ g.seed ^ int64(chunkX)*341873128712 ^ int64(chunkZ)*132897987541
	return gen.NewXoroshiro128FromSeed(seed)
}
