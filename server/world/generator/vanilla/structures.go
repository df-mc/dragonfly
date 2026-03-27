package vanilla

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

type structureStartCache struct {
	mu    sync.RWMutex
	byKey map[structureStartKey]cachedStructureStart
}

type cachedStructureStart struct {
	loaded bool
	exists bool
	start  plannedStructureStart
}

type structureStartKey struct {
	setName string
	chunkX  int32
	chunkZ  int32
}

type plannedStructureStart struct {
	setName       string
	structureName string
	templateName  string
	startChunk    world.ChunkPos
	origin        cube.Pos
	size          [3]int
	rootOrigin    cube.Pos
	rootSize      [3]int
	pieces        []plannedStructurePiece
}

type PlannedStructureInfo struct {
	StructureSet string
	Structure    string
	Template     string
	StartChunk   world.ChunkPos
	Origin       cube.Pos
	Size         [3]int
	PaletteNames []string
}

type structurePlanner struct {
	setName         string
	placement       gen.RandomSpreadPlacement
	candidates      []structurePlannerCandidate
	candidateByName map[string]int
	totalWeight     int
	maxBackreachX   int
	maxBackreachZ   int
}

type structurePlannerCandidate struct {
	structureName       string
	structureType       string
	biomeTag            string
	weight              int
	generic             gen.GenericStructureDef
	netherFossil        gen.NetherFossilStructureDef
	jigsaw              gen.JigsawStructureDef
	shipwreck           gen.ShipwreckStructureDef
	oceanRuin           gen.OceanRuinStructureDef
	ruinedPortal        gen.RuinedPortalStructureDef
	startTemplates      []weightedStartTemplate
	totalTemplateWeight int
	maxBackreachX       int
	maxBackreachZ       int
}

type weightedStartTemplate struct {
	name       string
	weight     int
	size       [3]int
	ignoreAir  bool
	projection string
	jigsaws    []structureJigsaw
	processors []structureProcessor
}

func newStructureStartCache() *structureStartCache {
	return &structureStartCache{byKey: make(map[structureStartKey]cachedStructureStart)}
}

func (c *structureStartCache) Lookup(key structureStartKey) (plannedStructureStart, bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.byKey[key]
	if !ok || !entry.loaded {
		return plannedStructureStart{}, false, false
	}
	return entry.start, entry.exists, true
}

func (c *structureStartCache) Store(key structureStartKey, start plannedStructureStart, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byKey[key] = cachedStructureStart{loaded: true, exists: exists, start: start}
}

func buildStructurePlanners(worldgen *gen.WorldgenRegistry, templates *gen.StructureTemplateRegistry, dim world.Dimension) []structurePlanner {
	if worldgen == nil || templates == nil {
		return nil
	}

	names := worldgen.StructureSetNames()
	planners := make([]structurePlanner, 0, len(names))
	for _, setName := range names {
		set, err := worldgen.StructureSet(setName)
		if err != nil || set.Placement.Type != "random_spread" {
			continue
		}
		placement, err := set.Placement.RandomSpread()
		if err != nil {
			continue
		}

		planner := structurePlanner{
			setName:         setName,
			placement:       placement,
			candidateByName: make(map[string]int),
		}
		for _, entry := range set.Structures {
			if entry.Weight <= 0 {
				continue
			}
			structureName := normalizeStructureName(entry.Structure)
			def, err := worldgen.Structure(structureName)
			if err != nil {
				continue
			}
			if !structureSupportedInDimension(structureName, def.Type, dim) {
				continue
			}
			candidate := structurePlannerCandidate{
				structureName: structureName,
				structureType: def.Type,
				weight:        entry.Weight,
			}

			switch def.Type {
			case "jigsaw":
				jigsaw, err := def.Jigsaw()
				if err != nil {
					continue
				}
				startTemplates, totalTemplateWeight, maxBackreachX, maxBackreachZ := buildStartTemplates(worldgen, templates, jigsaw.StartPool)
				if len(startTemplates) == 0 || totalTemplateWeight <= 0 {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(jigsaw.Biomes)
				candidate.jigsaw = jigsaw
				candidate.startTemplates = startTemplates
				candidate.totalTemplateWeight = totalTemplateWeight
				candidate.maxBackreachX = maxBackreachX
				candidate.maxBackreachZ = maxBackreachZ
				reachChunks := templateBackreachChunks(jigsaw.MaxDistanceFromCenter*2 + 1)
				if reachChunks > candidate.maxBackreachX {
					candidate.maxBackreachX = reachChunks
				}
				if reachChunks > candidate.maxBackreachZ {
					candidate.maxBackreachZ = reachChunks
				}
			case "igloo", "buried_treasure", "swamp_hut":
				generic, err := def.Generic()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(generic.Biomes)
				candidate.generic = generic
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			case "end_city":
				generic, err := def.Generic()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(generic.Biomes)
				candidate.generic = generic
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			case "nether_fossil":
				netherFossil, err := def.NetherFossil()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(netherFossil.Biomes)
				candidate.netherFossil = netherFossil
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			case "shipwreck":
				shipwreck, err := def.Shipwreck()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(shipwreck.Biomes)
				candidate.shipwreck = shipwreck
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			case "ocean_ruin":
				oceanRuin, err := def.OceanRuin()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(oceanRuin.Biomes)
				candidate.oceanRuin = oceanRuin
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			case "ruined_portal":
				ruinedPortal, err := def.RuinedPortal()
				if err != nil {
					continue
				}
				candidate.biomeTag = normalizeStructureTag(ruinedPortal.Biomes)
				candidate.ruinedPortal = ruinedPortal
				candidate.maxBackreachX, candidate.maxBackreachZ = estimateDirectStructureBackreach(structureName, def.Type)
			default:
				continue
			}
			planner.candidateByName[structureName] = len(planner.candidates)
			planner.candidates = append(planner.candidates, candidate)
			planner.totalWeight += entry.Weight
			if candidate.maxBackreachX > planner.maxBackreachX {
				planner.maxBackreachX = candidate.maxBackreachX
			}
			if candidate.maxBackreachZ > planner.maxBackreachZ {
				planner.maxBackreachZ = candidate.maxBackreachZ
			}
		}
		if len(planner.candidates) == 0 {
			continue
		}
		planners = append(planners, planner)
	}
	return planners
}

func structureSupportedInDimension(structureName, structureType string, dim world.Dimension) bool {
	switch dim {
	case world.Nether:
		switch structureName {
		case "bastion_remnant", "nether_fossil", "ruined_portal_nether":
			return true
		default:
			return false
		}
	case world.End:
		return structureName == "end_city"
	default:
		switch structureType {
		case "end_city", "nether_fossil":
			return false
		default:
			return structureName != "bastion_remnant" && structureName != "ruined_portal_nether"
		}
	}
}

func buildStartTemplates(worldgen *gen.WorldgenRegistry, templates *gen.StructureTemplateRegistry, poolName string) ([]weightedStartTemplate, int, int, int) {
	pool, err := worldgen.TemplatePool(poolName)
	if err != nil {
		return nil, 0, 0, 0
	}

	startTemplates := make([]weightedStartTemplate, 0, len(pool.Elements))
	totalWeight := 0
	maxBackreachX := 0
	maxBackreachZ := 0
	for _, entry := range pool.Elements {
		single, err := entry.Element.Single()
		if err != nil || single.Location == "" || entry.Weight <= 0 {
			continue
		}
		template, err := templates.Template(single.Location)
		if err != nil {
			continue
		}
		startTemplates = append(startTemplates, weightedStartTemplate{
			name:       single.Location,
			weight:     entry.Weight,
			size:       template.Size,
			ignoreAir:  entry.Element.ElementType == "legacy_single_pool_element",
			projection: normalizeIdentifierName(single.Projection),
			jigsaws:    extractTemplateJigsaws(template),
			processors: compileStructureProcessors(worldgen, single.Processors),
		})
		totalWeight += entry.Weight
		if backreach := templateBackreachChunks(template.Size[0]); backreach > maxBackreachX {
			maxBackreachX = backreach
		}
		if backreach := templateBackreachChunks(template.Size[2]); backreach > maxBackreachZ {
			maxBackreachZ = backreach
		}
	}
	return startTemplates, totalWeight, maxBackreachX, maxBackreachZ
}

func templateBackreachChunks(size int) int {
	if size <= 1 {
		return 0
	}
	return (size - 1) / 16
}

func (g Generator) findStructurePlanner(setName string) (structurePlanner, bool) {
	normalized := normalizeStructureName(setName)
	for _, planner := range g.structurePlanners {
		if planner.setName == normalized {
			return planner, true
		}
	}
	return structurePlanner{}, false
}

func (g Generator) placeStructures(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int) {
	if g.structureTemplates == nil || g.structureStarts == nil || len(g.structurePlanners) == 0 {
		return
	}

	for _, planner := range g.structurePlanners {
		g.placeRandomSpreadStructureSet(c, biomes, chunkX, chunkZ, minY, maxY, planner)
	}
}

func (g Generator) placeRandomSpreadStructureSet(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int, planner structurePlanner) {
	startMinChunkX := chunkX - planner.maxBackreachX
	startMaxChunkX := chunkX + planner.maxBackreachX
	startMinChunkZ := chunkZ - planner.maxBackreachZ
	startMaxChunkZ := chunkZ + planner.maxBackreachZ

	minGridX := randomSpreadMinGrid(startMinChunkX, planner.placement.Spacing, planner.placement.Separation)
	maxGridX := floorDiv(startMaxChunkX, planner.placement.Spacing)
	minGridZ := randomSpreadMinGrid(startMinChunkZ, planner.placement.Spacing, planner.placement.Separation)
	maxGridZ := floorDiv(startMaxChunkZ, planner.placement.Spacing)

	for gridX := minGridX; gridX <= maxGridX; gridX++ {
		for gridZ := minGridZ; gridZ <= maxGridZ; gridZ++ {
			startChunk := randomSpreadPotentialChunk(g.seed, planner.placement, gridX, gridZ)
			if int(startChunk[0]) < startMinChunkX || int(startChunk[0]) > startMaxChunkX || int(startChunk[1]) < startMinChunkZ || int(startChunk[1]) > startMaxChunkZ {
				continue
			}
			start, ok := g.planStructureStart(planner, startChunk, minY, maxY)
			if !ok || !structureIntersectsChunk(start, chunkX, chunkZ, minY, maxY) {
				continue
			}
			g.placePlannedStructure(c, biomes, chunkX, chunkZ, minY, maxY, start)
		}
	}
}

func randomSpreadMinGrid(startMinChunk, spacing, separation int) int {
	if spacing <= 0 {
		return 0
	}
	maxOffset := spacing - separation - 1
	if maxOffset < 0 {
		maxOffset = 0
	}
	return ceilDiv(startMinChunk-maxOffset, spacing)
}

func ceilDiv(value, divisor int) int {
	return -floorDiv(-value, divisor)
}

func (g Generator) planStructureStart(planner structurePlanner, startChunk world.ChunkPos, minY, maxY int) (plannedStructureStart, bool) {
	cacheKey := structureStartKey{setName: planner.setName, chunkX: startChunk[0], chunkZ: startChunk[1]}
	if start, exists, ok := g.structureStarts.Lookup(cacheKey); ok {
		return start, exists
	}
	if !structurePlacementAllows(g.seed, planner.placement, int(startChunk[0]), int(startChunk[1])) {
		g.structureStarts.Store(cacheKey, plannedStructureStart{}, false)
		return plannedStructureStart{}, false
	}

	startX := int(startChunk[0]) * 16
	startZ := int(startChunk[1]) * 16
	surfaceY := g.preliminarySurfaceLevelAt(startX+8, startZ+8, minY, maxY)
	if surfaceY < minY {
		surfaceY = minY
	}
	if surfaceY > maxY {
		surfaceY = maxY
	}
	surfaceBiome := g.biomeSource.GetBiome(startX+8, surfaceY, startZ+8)

	candidate, ok := g.chooseStructureForPlanner(planner, surfaceBiome, startChunk)
	if !ok {
		g.structureStarts.Store(cacheKey, plannedStructureStart{}, false)
		return plannedStructureStart{}, false
	}

	rng := g.structureRNG(planner.setName, startChunk)
	var (
		templateName  string
		pieces        []plannedStructurePiece
		overallBounds structureBox
		rootOrigin    cube.Pos
		rootSize      [3]int
		okBuild       bool
	)
	if candidate.structureType == "jigsaw" {
		startTemplate, ok := chooseStartTemplate(candidate, &rng)
		if !ok {
			g.structureStarts.Store(cacheKey, plannedStructureStart{}, false)
			return plannedStructureStart{}, false
		}
		templateName = startTemplate.name
		pieces, overallBounds, rootOrigin, rootSize, okBuild = g.buildPlannedStructure(candidate, startTemplate, startX, startZ, minY, maxY, &rng)
	} else {
		templateName, pieces, overallBounds, rootOrigin, rootSize, okBuild = g.buildPlannedDirectStructure(candidate, planner.placement, startChunk, startX, startZ, surfaceY, minY, maxY, &rng)
	}
	if !okBuild || len(pieces) == 0 {
		g.structureStarts.Store(cacheKey, plannedStructureStart{}, false)
		return plannedStructureStart{}, false
	}
	overallOrigin, overallSize := overallBounds.originAndSize()
	start := plannedStructureStart{
		setName:       planner.setName,
		structureName: candidate.structureName,
		templateName:  templateName,
		startChunk:    startChunk,
		origin:        overallOrigin,
		size:          overallSize,
		rootOrigin:    rootOrigin,
		rootSize:      rootSize,
		pieces:        pieces,
	}
	g.structureStarts.Store(cacheKey, start, true)
	return start, true
}

func (g Generator) chooseStructureForPlanner(planner structurePlanner, biome gen.Biome, startChunk world.ChunkPos) (structurePlannerCandidate, bool) {
	if len(planner.candidates) == 0 {
		return structurePlannerCandidate{}, false
	}

	if planner.setName == "villages" {
		name, ok := villageStructureNameForBiome(biome)
		if !ok {
			return structurePlannerCandidate{}, false
		}
		if index, ok := planner.candidateByName[name]; ok {
			return planner.candidates[index], true
		}
		return structurePlannerCandidate{}, false
	}

	if len(planner.candidates) == 1 {
		if g.structureCandidateAllowed(planner.candidates[0], biome) {
			return planner.candidates[0], true
		}
		return structurePlannerCandidate{}, false
	}

	totalWeight := 0
	for _, candidate := range planner.candidates {
		if candidate.weight <= 0 || !g.structureCandidateAllowed(candidate, biome) {
			continue
		}
		totalWeight += candidate.weight
	}
	if totalWeight <= 0 {
		return structurePlannerCandidate{}, false
	}

	rng := g.structureRNG(planner.setName+":structure", startChunk)
	pick := int(rng.NextInt(uint32(totalWeight)))
	for _, candidate := range planner.candidates {
		if candidate.weight <= 0 {
			continue
		}
		if !g.structureCandidateAllowed(candidate, biome) {
			continue
		}
		if pick < candidate.weight {
			return candidate, true
		}
		pick -= candidate.weight
	}
	return structurePlannerCandidate{}, false
}

func villageStructureNameForBiome(biome gen.Biome) (string, bool) {
	switch {
	case biome == gen.BiomePlains || biome == gen.BiomeSunflowerPlains:
		return "village_plains", true
	case biome == gen.BiomeDesert:
		return "village_desert", true
	case biome == gen.BiomeSavanna || biome == gen.BiomeSavannaPlateau || biome == gen.BiomeWindsweptSavanna:
		return "village_savanna", true
	case biome == gen.BiomeSnowyPlains || biome == gen.BiomeIceSpikes || biome == gen.BiomeGrove || biome == gen.BiomeSnowySlopes || biome == gen.BiomeFrozenPeaks:
		return "village_snowy", true
	case biome == gen.BiomeTaiga || biome == gen.BiomeSnowyTaiga || biome == gen.BiomeOldGrowthPineTaiga || biome == gen.BiomeOldGrowthSpruceTaiga:
		return "village_taiga", true
	default:
		return "", false
	}
}

func chooseStartTemplate(candidate structurePlannerCandidate, rng *gen.Xoroshiro128) (weightedStartTemplate, bool) {
	if candidate.totalTemplateWeight <= 0 {
		return weightedStartTemplate{}, false
	}

	pick := int(rng.NextInt(uint32(candidate.totalTemplateWeight)))
	for _, startTemplate := range candidate.startTemplates {
		if startTemplate.weight <= 0 {
			continue
		}
		if pick < startTemplate.weight {
			return startTemplate, true
		}
		pick -= startTemplate.weight
	}
	return weightedStartTemplate{}, false
}

func (g Generator) resolveJigsawStartY(def gen.JigsawStructureDef, blockX, blockZ, minY, maxY int, rng *gen.Xoroshiro128) int {
	base := g.sampleStructureHeightProvider(def.StartHeight, minY, maxY, rng)
	if def.ProjectStartToHeight != "" {
		return g.preliminarySurfaceLevelAt(blockX, blockZ, minY, maxY) + base
	}
	return base
}

func (g Generator) sampleStructureHeightProvider(provider gen.StructureHeightProvider, minY, maxY int, rng *gen.Xoroshiro128) int {
	switch provider.Kind {
	case "constant":
		return resolveVerticalAnchor(provider.Anchor, minY, maxY)
	case "uniform", "trapezoid", "biased_to_bottom", "very_biased_to_bottom":
		minValue := resolveVerticalAnchor(provider.MinInclusive, minY, maxY)
		maxValue := resolveVerticalAnchor(provider.MaxInclusive, minY, maxY)
		if maxValue <= minValue {
			return minValue
		}
		return minValue + int(rng.NextInt(uint32(maxValue-minValue+1)))
	default:
		return minY
	}
}

func resolveVerticalAnchor(anchor gen.VerticalAnchor, minY, maxY int) int {
	switch anchor.Kind {
	case "above_bottom":
		return minY + anchor.Value
	case "below_top":
		return maxY - anchor.Value
	default:
		return anchor.Value
	}
}

func (g Generator) preliminarySurfaceLevelAt(blockX, blockZ, minY, maxY int) int {
	chunkX := floorDiv(blockX, 16)
	chunkZ := floorDiv(blockZ, 16)
	flat := g.graph.NewFlatCacheGrid(chunkX, chunkZ, g.noises)
	col := g.graph.NewColumnContext(blockX, blockZ, g.noises, flat)
	ctx := gen.FunctionContext{BlockX: blockX, BlockY: 0, BlockZ: blockZ}
	value := 0.0
	if g.dimension == world.Overworld {
		value = gen.ComputePreliminarySurfaceLevel(ctx, g.noises, flat, col)
	} else {
		value = g.graph.Eval(g.rootIndex("preliminary_surface_level"), ctx, g.noises, flat, col, nil)
	}
	y := int(math.Floor(value))
	if y < minY {
		return minY
	}
	if y > maxY {
		return maxY
	}
	return y
}

func structureIntersectsChunk(start plannedStructureStart, chunkX, chunkZ, minY, maxY int) bool {
	return structureBox{
		minX: start.origin[0],
		minY: start.origin[1],
		minZ: start.origin[2],
		maxX: start.origin[0] + start.size[0] - 1,
		maxY: start.origin[1] + start.size[1] - 1,
		maxZ: start.origin[2] + start.size[2] - 1,
	}.intersectsChunk(chunkX, chunkZ, minY, maxY)
}

func (g Generator) placePlannedStructure(c *chunk.Chunk, biomes sourceBiomeVolume, chunkX, chunkZ, minY, maxY int, start plannedStructureStart) {
	for _, piece := range start.pieces {
		if !piece.bounds.intersectsChunk(chunkX, chunkZ, minY, maxY) {
			continue
		}
		for _, blockInfo := range piece.manualBlocks {
			g.placeStructureBlockState(c, chunkX, chunkZ, minY, maxY, blockInfo.worldPos, blockInfo.state)
		}
		for _, placement := range piece.element.placements {
			template, err := g.structureTemplates.Template(placement.templateName)
			if err != nil {
				continue
			}
			for _, blockInfo := range g.processStructureTemplatePlacement(c, chunkX, chunkZ, piece.origin, piece.rotation, piece.mirror, piece.pivot, piece.useTemplateTransform, template, placement) {
				switch blockInfo.state.Name {
				case "structure_void", "jigsaw", "structure_block":
					continue
				}
				if blockInfo.state.Name == "air" && placement.ignoreAir && blockInfo.originalState.Name == "air" {
					continue
				}
				worldX := blockInfo.worldPos[0]
				worldY := blockInfo.worldPos[1]
				worldZ := blockInfo.worldPos[2]
				if worldX < chunkX*16 || worldX >= chunkX*16+16 || worldZ < chunkZ*16 || worldZ >= chunkZ*16+16 || worldY < minY || worldY > maxY {
					continue
				}

				placedState := applyPlacedStructureStateTransform(blockInfo.state, piece.mirror, piece.rotation)
				g.placeStructureBlockState(c, chunkX, chunkZ, minY, maxY, blockInfo.worldPos, placedState)
			}
		}

		for _, feature := range piece.element.features {
			if piece.origin[0] < chunkX*16 || piece.origin[0] >= chunkX*16+16 || piece.origin[2] < chunkZ*16 || piece.origin[2] >= chunkZ*16+16 {
				continue
			}
			if piece.origin[1] <= minY || piece.origin[1] > maxY {
				continue
			}
			biomeName := biomeKey(g.biomeSource.GetBiome(piece.origin[0], clamp(piece.origin[1], minY+1, maxY), piece.origin[2]))
			rng := g.structureFeatureRNG(start.structureName, feature.featureName, piece.origin)
			_ = g.executePlacedFeatureRef(
				c,
				biomes,
				piece.origin,
				gen.PlacedFeatureRef{Name: feature.featureName},
				biomeName,
				chunkX,
				chunkZ,
				minY,
				maxY,
				&rng,
				0,
			)
		}
	}
}

func (g Generator) placeStructureBlockState(c *chunk.Chunk, chunkX, chunkZ, minY, maxY int, worldPos cube.Pos, state gen.BlockState) {
	worldX := worldPos[0]
	worldY := worldPos[1]
	worldZ := worldPos[2]
	if worldX < chunkX*16 || worldX >= chunkX*16+16 || worldZ < chunkZ*16 || worldZ >= chunkZ*16+16 || worldY < minY || worldY > maxY {
		return
	}
	rid, ok := g.lookupTemplateBlock(structureLookupName(state.Name), structureLookupProperties(state.Name, state.Properties))
	if !ok {
		return
	}
	c.SetBlock(uint8(worldX-chunkX*16), int16(worldY), uint8(worldZ-chunkZ*16), 0, rid)
}

func (g Generator) lookupTemplateBlock(name string, properties map[string]any) (uint32, bool) {
	key := templateBlockCacheKey(name, properties)
	if rid, ok := g.surfaceBlockCache.Lookup(key); ok {
		return rid, true
	}

	blockProps := make(map[string]any, len(properties))
	for key, value := range properties {
		switch v := value.(type) {
		case bool, int32, string:
			blockProps[key] = v
		case float64:
			blockProps[key] = int32(v)
		default:
			blockProps[key] = fmt.Sprint(v)
		}
	}
	if len(blockProps) == 0 {
		blockProps = nil
	}

	rid, ok := chunk.StateToRuntimeID(name, blockProps)
	if !ok {
		return 0, false
	}
	g.surfaceBlockCache.Store(key, rid)
	return rid, true
}

func templateBlockCacheKey(name string, properties map[string]any) string {
	if len(properties) == 0 {
		return name
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	for _, key := range keys {
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(key))
		_, _ = h.Write([]byte{'='})
		_, _ = h.Write([]byte(fmt.Sprint(properties[key])))
	}
	return fmt.Sprintf("%s#%x", name, h.Sum64())
}

func (g Generator) structureRNG(name string, pos world.ChunkPos) gen.Xoroshiro128 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	seed := int64(h.Sum64()) ^ g.seed ^ int64(pos[0])*341873128712 ^ int64(pos[1])*132897987541
	return gen.NewXoroshiro128FromSeed(seed)
}

func (g Generator) structureFeatureRNG(structureName, featureName string, pos cube.Pos) gen.Xoroshiro128 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(structureName))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(featureName))
	seed := int64(h.Sum64()) ^ g.seed ^ int64(pos[0])*341873128712 ^ int64(pos[1])*132897987541 ^ int64(pos[2])*42317861
	return gen.NewXoroshiro128FromSeed(seed)
}

func randomSpreadPotentialChunk(seed int64, placement gen.RandomSpreadPlacement, gridX, gridZ int) world.ChunkPos {
	rng := newLegacyRandom(int64(gridX)*341873128712 + int64(gridZ)*132897987541 + seed + int64(placement.Salt))
	limit := placement.Spacing - placement.Separation
	if limit <= 0 {
		return world.ChunkPos{int32(gridX * placement.Spacing), int32(gridZ * placement.Spacing)}
	}

	spreadX := rng.NextInt(limit)
	spreadZ := rng.NextInt(limit)
	if placement.SpreadType == "triangular" {
		spreadX = (spreadX + rng.NextInt(limit)) / 2
		spreadZ = (spreadZ + rng.NextInt(limit)) / 2
	}
	return world.ChunkPos{
		int32(gridX*placement.Spacing + spreadX),
		int32(gridZ*placement.Spacing + spreadZ),
	}
}

type legacyRandom struct {
	seed int64
}

func newLegacyRandom(seed int64) legacyRandom {
	return legacyRandom{seed: (seed ^ 25214903917) & 281474976710655}
}

func (r *legacyRandom) next(bits int) int {
	r.seed = (r.seed*25214903917 + 11) & 281474976710655
	return int(uint64(r.seed) >> (48 - bits))
}

func (r *legacyRandom) NextInt(bound int) int {
	if bound <= 1 {
		return 0
	}
	if bound&(bound-1) == 0 {
		return int((int64(bound) * int64(r.next(31))) >> 31)
	}
	for {
		bits := r.next(31)
		value := bits % bound
		if bits-value+(bound-1) >= 0 {
			return value
		}
	}
}

func (r *legacyRandom) NextFloat64() float64 {
	return float64(r.next(24)) / (1 << 24)
}

func (r *legacyRandom) NextDouble() float64 {
	return (float64(uint64(r.next(26))<<27) + float64(r.next(27))) / (1 << 53)
}

func floorDiv(value, divisor int) int {
	quotient := value / divisor
	remainder := value % divisor
	if remainder != 0 && ((remainder < 0) != (divisor < 0)) {
		quotient--
	}
	return quotient
}

func normalizeStructureName(name string) string {
	if len(name) >= 10 && name[:10] == "minecraft:" {
		return name[10:]
	}
	return name
}

func FindPlannedStructureStart(seed int64, setName string, maxGridDistance int) (PlannedStructureInfo, bool) {
	return FindPlannedStructureStartForDimension(seed, world.Overworld, setName, maxGridDistance)
}

func FindPlannedStructureStartForDimension(seed int64, dim world.Dimension, setName string, maxGridDistance int) (PlannedStructureInfo, bool) {
	g := NewForDimension(seed, dim)

	planner, ok := g.findStructurePlanner(setName)
	if !ok {
		return PlannedStructureInfo{}, false
	}

	for gridX := -maxGridDistance; gridX <= maxGridDistance; gridX++ {
		for gridZ := -maxGridDistance; gridZ <= maxGridDistance; gridZ++ {
			startChunk := randomSpreadPotentialChunk(seed, planner.placement, gridX, gridZ)
			start, exists := g.planStructureStart(planner, startChunk, -64, 319)
			if !exists {
				continue
			}
			paletteNames := make([]string, 0, 16)
			seen := make(map[string]struct{}, 32)
			for _, piece := range start.pieces {
				for _, placement := range piece.element.placements {
					template, err := g.structureTemplates.Template(placement.templateName)
					if err != nil {
						continue
					}
					for _, state := range template.Palette {
						switch state.Name {
						case "minecraft:air", "minecraft:jigsaw", "minecraft:structure_void":
							continue
						}
						if _, ok := seen[state.Name]; ok {
							continue
						}
						seen[state.Name] = struct{}{}
						paletteNames = append(paletteNames, state.Name)
					}
				}
			}
			sort.Strings(paletteNames)
			infoOrigin := start.rootOrigin
			infoSize := start.rootSize
			if infoSize[0] <= 0 || infoSize[1] <= 0 || infoSize[2] <= 0 {
				infoOrigin = start.origin
				infoSize = start.size
			}
			return PlannedStructureInfo{
				StructureSet: setName,
				Structure:    start.structureName,
				Template:     start.templateName,
				StartChunk:   start.startChunk,
				Origin:       infoOrigin,
				Size:         infoSize,
				PaletteNames: paletteNames,
			}, true
		}
	}
	return PlannedStructureInfo{}, false
}
