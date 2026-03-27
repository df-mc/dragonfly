package vanilla

import (
	"fmt"
	"sync"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

const seaLevel = 63
const featureStepCount = int(gen.GenerationStepTopLayerModification) + 1

type Generator struct {
	dimension          world.Dimension
	dimensionName      string
	seed               int64
	graph              *gen.Graph
	graphRoots         map[string]int
	noises             *gen.NoiseRegistry
	worldgen           *gen.WorldgenRegistry
	metadata           gen.DimensionMetadata
	biomeSource        gen.BiomeSource
	carvers            *gen.CarverRegistry
	features           *gen.FeatureRegistry
	biomeGeneration    *biomeGenerationIndex
	structureTemplates *gen.StructureTemplateRegistry
	structureResolver  *structureResolver
	structurePlanners  []structurePlanner
	structureStarts    *structureStartCache
	surface            *gen.SurfaceRuntime
	surfaceBlockCache  *blockRIDCache
	featureNoiseCache  *doublePerlinNoiseCache
	finalDensityScalar gen.DensityScalarEvaluator
	finalDensityVector gen.DensityVectorEvaluator
	airRID             uint32
	defaultBlockRID    uint32
	deepRID            uint32
	bedrockRID         uint32
	defaultFluidRID    uint32
	waterRID           uint32
	lavaRID            uint32
	forceBottomBedrock bool
}

func New(seed int64) Generator {
	return NewForDimension(seed, world.Overworld)
}

func NewForDimension(seed int64, dim world.Dimension) Generator {
	noises := gen.NewNoiseRegistry(seed)
	worldgen := gen.NewWorldgenRegistry()
	dimensionName, graph, roots, surfaceRuntime, forceBottomBedrock, scalar, vector := dimensionRuntime(seed, dim, noises, worldgen)
	biomeSource, err := gen.NewBiomeSource(seed, worldgen, dimensionName)
	if err != nil {
		panic(err)
	}
	if surfaceRuntime == nil {
		surfaceRuntime = dimensionSurfaceRuntime(seed, dim, noises, biomeSource)
	}
	metadata, err := worldgen.DimensionMetadata("minecraft:" + dimensionName)
	if err != nil {
		panic(err)
	}
	structureTemplates := gen.NewStructureTemplateRegistry(worldgen)
	structureResolver := newStructureResolver(worldgen, structureTemplates)
	structurePlanners := buildStructurePlanners(worldgen, structureTemplates, dim)
	carvers := gen.NewCarverRegistry()
	features := gen.NewFeatureRegistry()
	biomeGeneration := newBiomeGenerationIndex(features, carvers)
	// Prewarm static structure pool/template data so chunk generation doesn't pay first-use decode costs.
	structureResolver.prewarmJigsawCandidates(structurePlanners)
	defaultBlockRID := runtimeIDForDimensionState(metadata.DefaultBlock)
	defaultFluidRID := runtimeIDForDimensionState(metadata.DefaultFluid)
	return Generator{
		dimension:          dim,
		dimensionName:      dimensionName,
		seed:               seed,
		graph:              graph,
		graphRoots:         roots,
		noises:             noises,
		worldgen:           worldgen,
		metadata:           metadata,
		biomeSource:        biomeSource,
		carvers:            carvers,
		features:           features,
		biomeGeneration:    biomeGeneration,
		structureTemplates: structureTemplates,
		structureResolver:  structureResolver,
		structurePlanners:  structurePlanners,
		structureStarts:    newStructureStartCache(),
		surface:            surfaceRuntime,
		surfaceBlockCache:  newBlockRIDCache(),
		featureNoiseCache:  newDoublePerlinNoiseCache(),
		finalDensityScalar: scalar,
		finalDensityVector: vector,
		airRID:             world.BlockRuntimeID(block.Air{}),
		defaultBlockRID:    defaultBlockRID,
		deepRID:            world.BlockRuntimeID(block.Deepslate{Type: block.NormalDeepslate(), Axis: cube.Y}),
		bedrockRID:         world.BlockRuntimeID(block.Bedrock{}),
		defaultFluidRID:    defaultFluidRID,
		waterRID:           world.BlockRuntimeID(block.Water{Still: true, Depth: 8}),
		lavaRID:            world.BlockRuntimeID(block.Lava{Still: true, Depth: 8}),
		forceBottomBedrock: forceBottomBedrock,
	}
}

func (g Generator) GenerateChunk(pos world.ChunkPos, c *chunk.Chunk) {
	chunkX := int(pos[0])
	chunkZ := int(pos[1])
	minY := c.Range().Min()
	maxY := c.Range().Max()
	flat := g.graph.NewFlatCacheGrid(chunkX, chunkZ, g.noises)
	finalDensityRoot := g.rootIndex("final_density")
	densityChunk := gen.NewFinalDensityChunkWithEvaluator(
		g.graph,
		finalDensityRoot,
		chunkX,
		chunkZ,
		minY,
		maxY,
		g.noises,
		flat,
		g.finalDensityScalar,
		g.finalDensityVector,
	)
	var aquifer *gen.NoiseBasedAquifer
	if g.metadata.AquifersEnabled {
		aquifer = gen.NewNoiseBasedAquifer(
			g.graph,
			chunkX,
			chunkZ,
			minY,
			maxY,
			g.noises,
			flat,
			g.seed,
			gen.OverworldFluidPicker{SeaLevel: g.metadata.SeaLevel},
		)
	}

	for localX := 0; localX < 16; localX++ {
		for localZ := 0; localZ < 16; localZ++ {
			worldX := chunkX*16 + localX
			worldZ := chunkZ*16 + localZ

			for y := minY + 1; y <= maxY; y++ {
				density := densityChunk.Density(localX, y, localZ)

				if density > 0 {
					rid := g.baseRuntimeID(y)
					c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, rid)
					continue
				}

				if aquifer != nil {
					switch aquifer.ComputeSubstance(
						gen.FunctionContext{BlockX: worldX, BlockY: y, BlockZ: worldZ},
						density,
					) {
					case gen.AquiferBarrier:
						c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, g.baseRuntimeID(y))
					case gen.AquiferWater:
						c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, g.waterRID)
					case gen.AquiferLava:
						c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, g.lavaRID)
					}
					continue
				}

				if y <= g.metadata.SeaLevel && g.defaultFluidRID != g.airRID {
					c.SetBlock(uint8(localX), int16(y), uint8(localZ), 0, g.defaultFluidRID)
				}
			}

			if g.forceBottomBedrock {
				c.SetBlock(uint8(localX), int16(minY), uint8(localZ), 0, g.bedrockRID)
			}
		}
	}

	biomes := g.populateBiomeVolume(c, chunkX, chunkZ, minY, maxY)
	g.carveTerrain(c, biomes, chunkX, chunkZ, minY, maxY, aquifer)
	g.applySurfaceAndBiomes(c, biomes, chunkX, chunkZ, minY, maxY)
	g.decorateFeatures(c, biomes, chunkX, chunkZ, minY, maxY)
	g.decorateEndMainIsland(c, chunkX, chunkZ, minY, maxY)
	g.placeStructures(c, biomes, chunkX, chunkZ, minY, maxY)
}

// ConcurrentChunkGeneration returns true because Generator guards its shared
// caches and registries internally.
func (g Generator) ConcurrentChunkGeneration() bool { return true }

func (g Generator) baseRuntimeID(y int) uint32 {
	if g.dimension == world.Overworld && y < 0 {
		return g.deepRID
	}
	return g.defaultBlockRID
}

func (g Generator) isSolidRID(rid uint32) bool {
	return rid != g.airRID && rid != g.waterRID && rid != g.lavaRID
}

func (g Generator) rootIndex(name string) int {
	if g.graphRoots == nil {
		return -1
	}
	if root, ok := g.graphRoots[name]; ok {
		return root
	}
	return -1
}

func dimensionRuntime(_ int64, dim world.Dimension, _ *gen.NoiseRegistry, _ *gen.WorldgenRegistry) (string, *gen.Graph, map[string]int, *gen.SurfaceRuntime, bool, gen.DensityScalarEvaluator, gen.DensityVectorEvaluator) {
	switch dim {
	case world.Overworld:
		return "overworld", gen.OverworldGraph, gen.OverworldRoots, nil, true, gen.ComputeFinalDensity, gen.ComputeFinalDensity4
	case world.Nether:
		return "nether", gen.NetherGraph, gen.NetherRoots, nil, true, nil, nil
	case world.End:
		return "end", gen.EndGraph, gen.EndRoots, nil, false, nil, nil
	default:
		panic(fmt.Sprintf("unsupported dimension %v", dim))
	}
}

func dimensionSurfaceRuntime(seed int64, dim world.Dimension, noises *gen.NoiseRegistry, biomeSource gen.BiomeSource) *gen.SurfaceRuntime {
	switch dim {
	case world.Overworld:
		return gen.NewOverworldSurfaceRuntime(seed, noises, biomeSource)
	case world.Nether:
		return gen.NewNetherSurfaceRuntime(seed, noises, biomeSource)
	case world.End:
		return gen.NewEndSurfaceRuntime(seed, noises, biomeSource)
	default:
		return nil
	}
}

func runtimeIDForDimensionState(state gen.DimensionBlockState) uint32 {
	switch state.Name {
	case "minecraft:air":
		return world.BlockRuntimeID(block.Air{})
	case "minecraft:stone":
		return world.BlockRuntimeID(block.Stone{})
	case "minecraft:netherrack":
		return world.BlockRuntimeID(block.Netherrack{})
	case "minecraft:end_stone":
		return world.BlockRuntimeID(block.EndStone{})
	case "minecraft:water":
		return world.BlockRuntimeID(block.Water{Still: true, Depth: 8})
	case "minecraft:lava":
		return world.BlockRuntimeID(block.Lava{Still: true, Depth: 8})
	}

	properties := make(map[string]any, len(state.Properties))
	for key, value := range state.Properties {
		properties[key] = value
	}
	if b, ok := world.BlockByName(state.Name, properties); ok {
		return world.BlockRuntimeID(b)
	}
	return world.BlockRuntimeID(block.Air{})
}

type blockRIDCache struct {
	mu    sync.RWMutex
	byKey map[string]uint32
}

func newBlockRIDCache() *blockRIDCache {
	return &blockRIDCache{byKey: make(map[string]uint32)}
}

func (c *blockRIDCache) Lookup(key string) (uint32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rid, ok := c.byKey[key]
	return rid, ok
}

func (c *blockRIDCache) Store(key string, rid uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byKey[key] = rid
}

type doublePerlinNoiseCache struct {
	mu    sync.RWMutex
	byKey map[string]gen.DoublePerlinNoise
}

func newDoublePerlinNoiseCache() *doublePerlinNoiseCache {
	return &doublePerlinNoiseCache{byKey: make(map[string]gen.DoublePerlinNoise)}
}

func (c *doublePerlinNoiseCache) Lookup(key string) (gen.DoublePerlinNoise, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	noise, ok := c.byKey[key]
	return noise, ok
}

func (c *doublePerlinNoiseCache) Store(key string, noise gen.DoublePerlinNoise) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byKey[key] = noise
}

type biomeGenerationIndex struct {
	featureSteps [256][featureStepCount][]string
	carverNames  [256][]string
}

func newBiomeGenerationIndex(features *gen.FeatureRegistry, carvers *gen.CarverRegistry) *biomeGenerationIndex {
	idx := &biomeGenerationIndex{}
	for _, biome := range sortedBiomesByKey {
		biomeID := int(biome)
		key := biomeKey(biome)
		if key == "" {
			continue
		}
		for step := 0; step < featureStepCount; step++ {
			idx.featureSteps[biomeID][step] = features.BiomePlacedFeatures(key, gen.GenerationStep(step))
		}
		idx.carverNames[biomeID] = carvers.BiomeCarvers(key)
	}
	return idx
}
