package vanilla

import (
	"math"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

func estimateDirectStructureBackreach(structureName, structureType string) (int, int) {
	switch structureType {
	case "ocean_ruin":
		return 3, 3
	case "shipwreck":
		return 2, 2
	case "ruined_portal":
		return 2, 2
	case "igloo":
		return 1, 1
	case "swamp_hut":
		return 1, 1
	case "buried_treasure":
		return 0, 0
	case "nether_fossil":
		return 2, 2
	case "end_city":
		return 16, 16
	default:
		_ = structureName
		return 0, 0
	}
}

func structurePlacementAllows(seed int64, placement gen.RandomSpreadPlacement, sourceX, sourceZ int) bool {
	if placement.Frequency <= 0 {
		return false
	}
	if placement.Frequency >= 1 {
		return true
	}
	switch placement.FrequencyReductionMethod {
	case "legacy_type_2":
		return structurePlacementProbability(seed, 10387320, sourceX, sourceZ, placement.Frequency)
	case "legacy_type_3":
		return structurePlacementLegacyDoubleProbability(seed, sourceX, sourceZ, placement.Frequency)
	case "legacy_type_1":
		return structurePlacementLegacyType1(seed, sourceX, sourceZ, placement.Frequency)
	default:
		return structurePlacementProbability(seed, placement.Salt, sourceX, sourceZ, placement.Frequency)
	}
}

func structurePlacementProbability(seed int64, salt, sourceX, sourceZ int, probability float64) bool {
	rng := newLegacyRandom(int64(sourceX)*341873128712 + int64(sourceZ)*132897987541 + seed + int64(salt))
	return rng.NextFloat64() < probability
}

func structurePlacementLegacyDoubleProbability(seed int64, sourceX, sourceZ int, probability float64) bool {
	rng := newLegacyRandom(int64(sourceX)*341873128712 + int64(sourceZ)*132897987541 + seed)
	return rng.NextDouble() < probability
}

func structurePlacementLegacyType1(seed int64, sourceX, sourceZ int, probability float64) bool {
	if probability <= 0 {
		return false
	}
	cx := sourceX >> 4
	cz := sourceZ >> 4
	rng := newLegacyRandom(int64(cx^(cz<<4)) ^ seed)
	_ = rng.next(32)
	return rng.NextInt(int(math.Max(1, 1.0/probability))) == 0
}

func (g Generator) structureCandidateAllowed(candidate structurePlannerCandidate, biome gen.Biome) bool {
	switch candidate.structureType {
	case "shipwreck":
		if candidate.shipwreck.IsBeached {
			return isBeachBiome(biome)
		}
		return isOceanBiome(biome)
	case "ocean_ruin":
		switch candidate.oceanRuin.BiomeTemp {
		case "warm":
			return biome == gen.BiomeWarmOcean || biome == gen.BiomeLukewarmOcean || biome == gen.BiomeDeepLukewarmOcean
		default:
			return biome == gen.BiomeFrozenOcean || biome == gen.BiomeColdOcean || biome == gen.BiomeOcean || biome == gen.BiomeDeepFrozenOcean || biome == gen.BiomeDeepColdOcean || biome == gen.BiomeDeepOcean
		}
	case "igloo":
		return biome == gen.BiomeSnowyTaiga || biome == gen.BiomeSnowyPlains || biome == gen.BiomeSnowySlopes
	case "buried_treasure":
		return isBeachBiome(biome)
	case "swamp_hut":
		return biome == gen.BiomeSwamp
	case "nether_fossil":
		return biome == gen.BiomeSoulSandValley
	case "end_city":
		return biome == gen.BiomeEndHighlands || biome == gen.BiomeEndMidlands || biome == gen.BiomeEndBarrens
	case "ruined_portal":
		return ruinedPortalBiomeAllowed(candidate.structureName, biome)
	default:
		if candidate.structureType == "jigsaw" {
			return true
		}
		return true
	}
}

func isBeachBiome(biome gen.Biome) bool {
	return biome == gen.BiomeBeach || biome == gen.BiomeSnowyBeach
}

func ruinedPortalBiomeAllowed(structureName string, biome gen.Biome) bool {
	switch structureName {
	case "ruined_portal_desert":
		return biome == gen.BiomeDesert
	case "ruined_portal_jungle":
		return isJungleBiome(biome)
	case "ruined_portal_nether":
		switch biome {
		case gen.BiomeNetherWastes, gen.BiomeSoulSandValley, gen.BiomeCrimsonForest, gen.BiomeWarpedForest, gen.BiomeBasaltDeltas:
			return true
		default:
			return false
		}
	case "ruined_portal_ocean":
		return isOceanBiome(biome)
	case "ruined_portal_swamp":
		return biome == gen.BiomeSwamp || biome == gen.BiomeMangroveSwamp
	case "ruined_portal_mountain":
		return isBadlandsBiome(biome) || isHillBiome(biome) || biome == gen.BiomeSavannaPlateau || biome == gen.BiomeWindsweptSavanna || biome == gen.BiomeStonyShore || isMountainBiome(biome)
	default:
		return isBeachBiome(biome) || isRiverBiome(biome) || isTaigaBiome(biome) || isForestBiome(biome) || biome == gen.BiomeMushroomFields || biome == gen.BiomeIceSpikes || biome == gen.BiomeDripstoneCaves || biome == gen.BiomeLushCaves || biome == gen.BiomeSavanna || biome == gen.BiomeSnowyPlains || biome == gen.BiomePlains || biome == gen.BiomeSunflowerPlains
	}
}

func isForestBiome(biome gen.Biome) bool {
	switch biome {
	case gen.BiomeForest, gen.BiomeFlowerForest, gen.BiomeBirchForest, gen.BiomeBirchForestHills, gen.BiomeDarkForest, gen.BiomeGrove:
		return true
	default:
		return false
	}
}

func isTaigaBiome(biome gen.Biome) bool {
	switch biome {
	case gen.BiomeTaiga, gen.BiomeSnowyTaiga, gen.BiomeOldGrowthPineTaiga, gen.BiomeOldGrowthSpruceTaiga:
		return true
	default:
		return false
	}
}

func isJungleBiome(biome gen.Biome) bool {
	switch biome {
	case gen.BiomeJungle, gen.BiomeSparseJungle, gen.BiomeBambooJungle:
		return true
	default:
		return false
	}
}

func isHillBiome(biome gen.Biome) bool {
	switch biome {
	case gen.BiomeWindsweptHills, gen.BiomeWindsweptForest, gen.BiomeTaigaHills, gen.BiomeJungleHills:
		return true
	default:
		return false
	}
}

func isMountainBiome(biome gen.Biome) bool {
	switch biome {
	case gen.BiomeMeadow, gen.BiomeFrozenPeaks, gen.BiomeJaggedPeaks, gen.BiomeStonyPeaks, gen.BiomeSnowySlopes, gen.BiomeGrove:
		return true
	default:
		return false
	}
}

func (g Generator) buildPlannedDirectStructure(
	candidate structurePlannerCandidate,
	placement gen.RandomSpreadPlacement,
	startChunk world.ChunkPos,
	startX, startZ, surfaceY, minY, maxY int,
	rng *gen.Xoroshiro128,
) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	switch candidate.structureType {
	case "igloo":
		return g.buildIglooStructure(startX, startZ, minY, maxY, rng)
	case "shipwreck":
		return g.buildShipwreckStructure(candidate.shipwreck, startX, startZ, minY, maxY, rng)
	case "ocean_ruin":
		return g.buildOceanRuinStructure(candidate.oceanRuin, startX, startZ, minY, maxY, rng)
	case "ruined_portal":
		return g.buildRuinedPortalStructure(candidate.ruinedPortal, startX, startZ, minY, maxY, rng)
	case "buried_treasure":
		return g.buildBuriedTreasureStructure(startX, startZ, minY, maxY, placement)
	case "swamp_hut":
		return g.buildSwampHutStructure(startChunk, startX, startZ, surfaceY, minY, maxY, rng)
	case "nether_fossil":
		return g.buildNetherFossilStructure(candidate.netherFossil, startX, startZ, minY, maxY, rng)
	case "end_city":
		return g.buildEndCityStructure(startChunk, startX, startZ, minY, maxY, rng)
	default:
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
}

func (g Generator) buildIglooStructure(startX, startZ, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	position := cube.Pos{startX, 90, startZ}
	rotation := randomStructureRotation(rng)

	topPiece, ok := g.newTemplateStructurePiece("igloo/top", position, rotation, structureMirrorNone, cube.Pos{3, 5, 5}, false, nil)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	entrance := structureTemplateWorldPos(position, [3]int{3, 0, 5}, rotation, structureMirrorNone, cube.Pos{3, 5, 5})
	deltaY := g.preliminarySurfaceLevelAt(entrance[0], entrance[2], minY, maxY) - 91

	translate := cube.Pos{0, deltaY, 0}
	topPiece.origin = topPiece.origin.Add(translate)
	topPiece.bounds = shiftStructureBox(topPiece.bounds, translate)
	pieces := []plannedStructurePiece{topPiece}
	overall := topPiece.bounds

	if rng.NextDouble() < 0.5 {
		depth := int(rng.NextInt(8)) + 4
		bottomPos := cube.Pos{startX, 90 - depth*3, startZ}
		bottomPos = bottomPos.Add(cube.Pos{0, -3 + deltaY, -2})
		bottom, ok := g.newTemplateStructurePiece("igloo/bottom", bottomPos, rotation, structureMirrorNone, cube.Pos{3, 6, 7}, false, nil)
		if ok {
			pieces = append(pieces, bottom)
			overall = unionStructureBoxes(overall, bottom.bounds)
		}
		for i := 0; i < depth-1; i++ {
			middlePos := cube.Pos{startX, 90 - i*3, startZ}
			middlePos = middlePos.Add(cube.Pos{2, -3 + deltaY, 4})
			middle, ok := g.newTemplateStructurePiece("igloo/middle", middlePos, rotation, structureMirrorNone, cube.Pos{1, 3, 1}, false, nil)
			if ok {
				pieces = append(pieces, middle)
				overall = unionStructureBoxes(overall, middle.bounds)
			}
		}
	}

	rootOrigin, rootSize := topPiece.bounds.originAndSize()
	return "igloo/top", pieces, overall, rootOrigin, rootSize, true
}

func (g Generator) buildShipwreckStructure(def gen.ShipwreckStructureDef, startX, startZ, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	templates := shipwreckOceanTemplates
	if def.IsBeached {
		templates = shipwreckBeachedTemplates
	}
	if len(templates) == 0 {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	templateName := templates[int(rng.NextInt(uint32(len(templates))))]
	reference := cube.Pos{startX, 90, startZ}
	rotation := randomStructureRotation(rng)
	piece, ok := g.newTemplateStructurePiece(templateName, reference, rotation, structureMirrorNone, cube.Pos{4, 0, 15}, true, nil)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}

	targetY := g.sampleTemplateMeanY(piece.bounds, minY, maxY)
	if def.IsBeached {
		targetY = g.sampleTemplateMinY(piece.bounds, minY, maxY) - piece.bounds.originAndSizeY()/2 - int(rng.NextInt(3))
	}
	translate := cube.Pos{0, targetY - piece.bounds.minY, 0}
	piece.origin = piece.origin.Add(translate)
	piece.bounds = shiftStructureBox(piece.bounds, translate)
	rootOrigin, rootSize := piece.bounds.originAndSize()
	return templateName, []plannedStructurePiece{piece}, piece.bounds, rootOrigin, rootSize, true
}

func (g Generator) buildOceanRuinStructure(def gen.OceanRuinStructureDef, startX, startZ, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	position := cube.Pos{startX + 8, 90, startZ + 8}
	rotation := randomStructureRotation(rng)
	isLarge := rng.NextDouble() <= def.LargeProbability
	baseIntegrity := 0.8
	if isLarge {
		baseIntegrity = 0.9
	}

	pieces, firstTemplate, ok := g.buildOceanRuinPieces(def, position, rotation, isLarge, baseIntegrity, rng)
	if !ok || len(pieces) == 0 {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	overall := emptyStructureBox()
	for i := range pieces {
		targetY := g.sampleTemplateFloorY(pieces[i].bounds, minY, maxY)
		translate := cube.Pos{0, targetY - pieces[i].bounds.minY, 0}
		pieces[i].origin = pieces[i].origin.Add(translate)
		pieces[i].bounds = shiftStructureBox(pieces[i].bounds, translate)
		overall = unionStructureBoxes(overall, pieces[i].bounds)
	}

	if isLarge && rng.NextDouble() <= def.ClusterProbability {
		for _, clusterPos := range oceanRuinClusterPositions(position, rotation, rng) {
			clusterPieces, _, ok := g.buildOceanRuinPieces(def, clusterPos, randomStructureRotation(rng), false, 0.8, rng)
			if !ok {
				continue
			}
			for i := range clusterPieces {
				targetY := g.sampleTemplateFloorY(clusterPieces[i].bounds, minY, maxY)
				translate := cube.Pos{0, targetY - clusterPieces[i].bounds.minY, 0}
				clusterPieces[i].origin = clusterPieces[i].origin.Add(translate)
				clusterPieces[i].bounds = shiftStructureBox(clusterPieces[i].bounds, translate)
				if !clusterPieces[i].bounds.empty() && !clusterPieces[i].bounds.intersects(overall) {
					pieces = append(pieces, clusterPieces[i])
					overall = unionStructureBoxes(overall, clusterPieces[i].bounds)
				}
			}
		}
	}

	rootOrigin, rootSize := pieces[0].bounds.originAndSize()
	return firstTemplate, pieces, overall, rootOrigin, rootSize, true
}

func (g Generator) buildRuinedPortalStructure(def gen.RuinedPortalStructureDef, startX, startZ, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	if len(def.Setups) == 0 {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	setup := chooseRuinedPortalSetup(def.Setups, rng)
	templateName := ruinedPortalTemplates[int(rng.NextInt(uint32(len(ruinedPortalTemplates))))]
	if rng.NextDouble() < 0.05 {
		templateName = ruinedPortalGiantTemplates[int(rng.NextInt(uint32(len(ruinedPortalGiantTemplates))))]
	}
	reference := cube.Pos{startX, 0, startZ}
	rotation := randomStructureRotation(rng)
	mirror := structureMirrorNone
	if rng.NextDouble() < 0.5 {
		mirror = structureMirrorFrontBack
	}

	template, err := g.structureTemplates.Template(templateName)
	if err != nil {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pivot := cube.Pos{template.Size[0] / 2, 0, template.Size[2] / 2}
	bounds := structureTemplateWorldBox(template, reference, rotation, mirror, pivot)
	centerX := (bounds.minX + bounds.maxX) / 2
	centerZ := (bounds.minZ + bounds.maxZ) / 2
	surfaceY := g.preliminarySurfaceLevelAt(centerX, centerZ, minY, maxY) - 1
	ySpan := bounds.maxY - bounds.minY + 1
	reference[1] = ruinedPortalTargetY(setup.Placement, surfaceY, ySpan, minY, maxY, rng)

	piece, ok := g.newTemplateStructurePiece(templateName, reference, rotation, mirror, pivot, true, nil)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	rootOrigin, rootSize := piece.bounds.originAndSize()
	return templateName, []plannedStructurePiece{piece}, piece.bounds, rootOrigin, rootSize, true
}

func (g Generator) buildBuriedTreasureStructure(startX, startZ, minY, maxY int, placement gen.RandomSpreadPlacement) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	chestPos := cube.Pos{startX + 9, g.preliminarySurfaceLevelAt(startX+9, startZ+9, minY, maxY) - 1, startZ + 9}
	if chestPos[1] <= minY {
		chestPos[1] = minY + 1
	}
	sand := gen.BlockState{Name: "sand"}
	stone := gen.BlockState{Name: "stone"}
	chest := gen.BlockState{Name: "chest", Properties: map[string]string{"facing": "north"}}
	blocks := []plannedStructureBlock{
		{worldPos: chestPos, state: chest},
		{worldPos: chestPos.Add(cube.Pos{1, 0, 0}), state: sand},
		{worldPos: chestPos.Add(cube.Pos{-1, 0, 0}), state: sand},
		{worldPos: chestPos.Add(cube.Pos{0, 0, 1}), state: sand},
		{worldPos: chestPos.Add(cube.Pos{0, 0, -1}), state: sand},
		{worldPos: chestPos.Add(cube.Pos{0, -1, 0}), state: stone},
	}
	box := structureBox{
		minX: chestPos[0] - 1, minY: chestPos[1] - 1, minZ: chestPos[2] - 1,
		maxX: chestPos[0] + 1, maxY: chestPos[1], maxZ: chestPos[2] + 1,
	}
	_ = placement
	piece := plannedStructurePiece{origin: chestPos, bounds: box, manualBlocks: blocks, rootPiece: true}
	rootOrigin, rootSize := box.originAndSize()
	return "buried_treasure", []plannedStructurePiece{piece}, box, rootOrigin, rootSize, true
}

func (g Generator) buildSwampHutStructure(startChunk world.ChunkPos, startX, startZ, surfaceY, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	_ = startChunk
	baseY := surfaceY
	if baseY < minY+1 {
		baseY = minY + 1
	}
	rotation := randomStructureRotation(rng)
	origin := cube.Pos{startX, baseY, startZ}
	blocks := buildSwampHutBlocks(origin, rotation, g, minY, maxY)
	if len(blocks) == 0 {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	box := emptyStructureBox()
	for _, block := range blocks {
		box = unionStructureBoxes(box, structureBox{minX: block.worldPos[0], minY: block.worldPos[1], minZ: block.worldPos[2], maxX: block.worldPos[0], maxY: block.worldPos[1], maxZ: block.worldPos[2]})
	}
	piece := plannedStructurePiece{origin: origin, rotation: rotation, bounds: box, manualBlocks: blocks, rootPiece: true}
	rootOrigin, rootSize := box.originAndSize()
	return "swamp_hut", []plannedStructurePiece{piece}, box, rootOrigin, rootSize, true
}

func (g Generator) buildNetherFossilStructure(def gen.NetherFossilStructureDef, startX, startZ, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	blockX := startX + int(rng.NextInt(16))
	blockZ := startZ + int(rng.NextInt(16))
	seaLevel := g.metadata.SeaLevel
	y := clamp(g.sampleStructureHeightProvider(def.Height, minY, maxY, rng), minY+1, maxY)
	for y > seaLevel {
		current := g.sampleStructureSubstanceAt(blockX, y, blockZ)
		below := g.sampleStructureSubstanceAt(blockX, y-1, blockZ)
		if current == structureSubstanceAir && below != structureSubstanceAir {
			break
		}
		y--
	}
	if y <= seaLevel {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}

	templateName := netherFossilTemplates[int(rng.NextInt(uint32(len(netherFossilTemplates))))]
	rotation := randomStructureRotation(rng)
	piece, ok := g.newTemplateStructurePiece(templateName, cube.Pos{blockX, y, blockZ}, rotation, structureMirrorNone, cube.Pos{}, true, nil)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	rootOrigin, rootSize := piece.bounds.originAndSize()
	return templateName, []plannedStructurePiece{piece}, piece.bounds, rootOrigin, rootSize, true
}

type structureSubstance uint8

const (
	structureSubstanceAir structureSubstance = iota
	structureSubstanceSolid
	structureSubstanceFluid
)

func (g Generator) sampleStructureSubstanceAt(blockX, blockY, blockZ int) structureSubstance {
	chunkX := floorDiv(blockX, 16)
	chunkZ := floorDiv(blockZ, 16)
	flat := g.graph.NewFlatCacheGrid(chunkX, chunkZ, g.noises)
	col := g.graph.NewColumnContext(blockX, blockZ, g.noises, flat)
	return g.sampleStructureSubstanceAtWithColumn(blockX, blockY, blockZ, flat, col)
}

func (g Generator) sampleStructureSubstanceAtWithColumn(blockX, blockY, blockZ int, flat *gen.FlatCacheGrid, col *gen.ColumnContext) structureSubstance {
	density := gen.EvalDensityScalar(
		g.graph,
		g.rootIndex("final_density"),
		gen.FunctionContext{BlockX: blockX, BlockY: blockY, BlockZ: blockZ},
		g.noises,
		flat,
		col,
		g.finalDensityScalar,
	)
	if density > 0 {
		return structureSubstanceSolid
	}
	if blockY <= g.metadata.SeaLevel && g.defaultFluidRID != g.airRID {
		return structureSubstanceFluid
	}
	return structureSubstanceAir
}

func (g Generator) newTemplateStructurePiece(templateName string, reference cube.Pos, rotation structureRotation, mirror structureMirror, pivot cube.Pos, ignoreAir bool, processors []structureProcessor) (plannedStructurePiece, bool) {
	template, err := g.structureTemplates.Template(templateName)
	if err != nil {
		return plannedStructurePiece{}, false
	}
	piece := plannedStructurePiece{
		element: resolvedPoolElement{
			placements: []structureTemplatePlacement{{
				templateName: templateName,
				ignoreAir:    ignoreAir,
				processors:   append([]structureProcessor(nil), processors...),
			}},
			size: template.Size,
		},
		origin:               reference,
		rotation:             rotation,
		mirror:               mirror,
		pivot:                pivot,
		useTemplateTransform: true,
	}
	piece.bounds = structureTemplateWorldBox(template, reference, rotation, mirror, pivot)
	return piece, true
}

func shiftStructureBox(box structureBox, offset cube.Pos) structureBox {
	if box.empty() {
		return box
	}
	return structureBox{
		minX: box.minX + offset[0],
		minY: box.minY + offset[1],
		minZ: box.minZ + offset[2],
		maxX: box.maxX + offset[0],
		maxY: box.maxY + offset[1],
		maxZ: box.maxZ + offset[2],
	}
}

func (b structureBox) originAndSizeY() int {
	return b.maxY - b.minY + 1
}

func (g Generator) sampleTemplateMeanY(box structureBox, minY, maxY int) int {
	total := 0
	count := 0
	for x := box.minX; x <= box.maxX; x++ {
		for z := box.minZ; z <= box.maxZ; z++ {
			total += g.preliminarySurfaceLevelAt(x, z, minY, maxY)
			count++
		}
	}
	if count == 0 {
		return minY
	}
	return total / count
}

func (g Generator) sampleTemplateMinY(box structureBox, minY, maxY int) int {
	value := maxY
	for x := box.minX; x <= box.maxX; x++ {
		for z := box.minZ; z <= box.maxZ; z++ {
			y := g.preliminarySurfaceLevelAt(x, z, minY, maxY)
			if y < value {
				value = y
			}
		}
	}
	return value
}

func (g Generator) sampleTemplateFloorY(box structureBox, minY, maxY int) int {
	topY := maxY
	minFloor := maxY
	steep := 0
	for x := box.minX; x <= box.maxX; x++ {
		for z := box.minZ; z <= box.maxZ; z++ {
			y := g.preliminarySurfaceLevelAt(x, z, minY, maxY)
			if y < minFloor {
				minFloor = y
			}
			if y < topY-2 {
				steep++
			}
		}
	}
	width := abs(box.maxX - box.minX)
	if topY-minFloor > 2 && steep > width-2 {
		return minFloor + 1
	}
	if minFloor < maxY {
		return minFloor + 1
	}
	return minY
}

func chooseRuinedPortalSetup(setups []gen.RuinedPortalSetupDef, rng *gen.Xoroshiro128) gen.RuinedPortalSetupDef {
	total := 0.0
	for _, setup := range setups {
		if setup.Weight > 0 {
			total += setup.Weight
		}
	}
	if total <= 0 {
		return setups[0]
	}
	pick := rng.NextDouble() * total
	for _, setup := range setups {
		if setup.Weight <= 0 {
			continue
		}
		if pick < setup.Weight {
			return setup
		}
		pick -= setup.Weight
	}
	return setups[0]
}

func ruinedPortalTargetY(placement string, surfaceY, ySpan, minY, maxY int, rng *gen.Xoroshiro128) int {
	switch placement {
	case "in_mountain":
		return randomBetweenInclusive(rng, 70, max(70, surfaceY-ySpan))
	case "underground":
		return randomBetweenInclusive(rng, minY+15, max(minY+15, surfaceY-ySpan))
	case "partly_buried":
		return surfaceY - ySpan + randomBetweenInclusive(rng, 2, 8)
	default:
		if surfaceY < minY {
			return minY
		}
		if surfaceY > maxY {
			return maxY
		}
		return surfaceY
	}
}

func randomBetweenInclusive(rng *gen.Xoroshiro128, minValue, maxValue int) int {
	if maxValue <= minValue {
		return minValue
	}
	return minValue + int(rng.NextInt(uint32(maxValue-minValue+1)))
}

func (g Generator) buildOceanRuinPieces(def gen.OceanRuinStructureDef, position cube.Pos, rotation structureRotation, isLarge bool, baseIntegrity float64, rng *gen.Xoroshiro128) ([]plannedStructurePiece, string, bool) {
	switch def.BiomeTemp {
	case "warm":
		templates := oceanRuinWarmTemplates
		if isLarge {
			templates = oceanRuinWarmLargeTemplates
		}
		templateName := templates[int(rng.NextInt(uint32(len(templates))))]
		processors := oceanRuinProcessors(baseIntegrity, true)
		piece, ok := g.newTemplateStructurePiece(templateName, position, rotation, structureMirrorNone, cube.Pos{}, true, processors)
		if !ok {
			return nil, "", false
		}
		return []plannedStructurePiece{piece}, templateName, true
	default:
		index := int(rng.NextInt(uint32(len(oceanRuinColdBrickTemplates))))
		templates := []struct {
			name      string
			integrity float64
		}{
			{name: oceanRuinColdBrickTemplates[indexForLarge(isLarge, index, oceanRuinColdBrickTemplates, oceanRuinColdBrickLargeTemplates)], integrity: baseIntegrity},
			{name: oceanRuinColdCrackedTemplates[indexForLarge(isLarge, index, oceanRuinColdCrackedTemplates, oceanRuinColdCrackedLargeTemplates)], integrity: 0.7},
			{name: oceanRuinColdMossyTemplates[indexForLarge(isLarge, index, oceanRuinColdMossyTemplates, oceanRuinColdMossyLargeTemplates)], integrity: 0.5},
		}
		pieces := make([]plannedStructurePiece, 0, len(templates))
		for _, entry := range templates {
			piece, ok := g.newTemplateStructurePiece(entry.name, position, rotation, structureMirrorNone, cube.Pos{}, true, oceanRuinProcessors(entry.integrity, false))
			if !ok {
				continue
			}
			pieces = append(pieces, piece)
		}
		if len(pieces) == 0 {
			return nil, "", false
		}
		return pieces, templates[0].name, true
	}
}

func indexForLarge(isLarge bool, index int, small, large []string) int {
	if isLarge {
		return min(index, len(large)-1)
	}
	return min(index, len(small)-1)
}

func oceanRuinProcessors(integrity float64, warm bool) []structureProcessor {
	processors := []structureProcessor{{
		kind: "block_rot",
		blockRot: &structureBlockRotProcessor{
			integrity: integrity,
		},
	}}
	suspiciousBlock := "gravel"
	loot := "ocean_ruin_cold_archaeology"
	if warm {
		suspiciousBlock = "sand"
		loot = "ocean_ruin_warm_archaeology"
	}
	processors = append(processors, structureProcessor{
		kind: "capped",
		capped: &structureCappedProcessor{
			limit: 5,
			delegate: structureProcessor{
				kind: "rule",
				rule: &structureRuleProcessor{rules: []structureProcessorRule{{
					input:    structureRuleTest{kind: "block_match", block: suspiciousBlock},
					location: structureRuleTest{kind: "always_true"},
					position: structurePosRuleTest{kind: "always_true"},
					output: gen.BlockState{
						Name: map[bool]string{true: "suspicious_sand", false: "suspicious_gravel"}[warm],
					},
					blockEntity: structureBlockEntityModifier{kind: "append_loot", lootTable: loot},
				}}},
			},
		},
	})
	return processors
}

func oceanRuinClusterPositions(origin cube.Pos, rotation structureRotation, rng *gen.Xoroshiro128) []cube.Pos {
	parentPos := cube.Pos{origin[0], 90, origin[2]}
	parentCorner := structureTemplateWorldPos(parentPos, [3]int{15, 0, 15}, rotation, structureMirrorNone, cube.Pos{})
	parentBottomLeft := cube.Pos{min(parentPos[0], parentCorner[0]), parentPos[1], min(parentPos[2], parentCorner[2])}
	positions := []cube.Pos{
		parentBottomLeft.Add(cube.Pos{-16 + randomBetweenInclusive(rng, 1, 8), 0, 16 + randomBetweenInclusive(rng, 1, 7)}),
		parentBottomLeft.Add(cube.Pos{-16 + randomBetweenInclusive(rng, 1, 8), 0, randomBetweenInclusive(rng, 1, 7)}),
		parentBottomLeft.Add(cube.Pos{-16 + randomBetweenInclusive(rng, 1, 8), 0, -16 + randomBetweenInclusive(rng, 4, 8)}),
		parentBottomLeft.Add(cube.Pos{randomBetweenInclusive(rng, 1, 7), 0, 16 + randomBetweenInclusive(rng, 1, 7)}),
		parentBottomLeft.Add(cube.Pos{randomBetweenInclusive(rng, 1, 7), 0, -16 + randomBetweenInclusive(rng, 4, 6)}),
		parentBottomLeft.Add(cube.Pos{16 + randomBetweenInclusive(rng, 1, 7), 0, 16 + randomBetweenInclusive(rng, 3, 8)}),
		parentBottomLeft.Add(cube.Pos{16 + randomBetweenInclusive(rng, 1, 7), 0, randomBetweenInclusive(rng, 1, 7)}),
		parentBottomLeft.Add(cube.Pos{16 + randomBetweenInclusive(rng, 1, 7), 0, -16 + randomBetweenInclusive(rng, 4, 8)}),
	}
	shuffleWithRNG(positions, rng)
	ruins := randomBetweenInclusive(rng, 4, 8)
	if ruins > len(positions) {
		ruins = len(positions)
	}
	return positions[:ruins]
}

func buildSwampHutBlocks(origin cube.Pos, rotation structureRotation, g Generator, minY, maxY int) []plannedStructureBlock {
	blocks := make([]plannedStructureBlock, 0, 256)
	fillBox := func(minPos, maxPos [3]int, state gen.BlockState) {
		for x := minPos[0]; x <= maxPos[0]; x++ {
			for y := minPos[1]; y <= maxPos[1]; y++ {
				for z := minPos[2]; z <= maxPos[2]; z++ {
					blocks = append(blocks, plannedStructureBlock{
						worldPos: structureTemplateWorldPos(origin, [3]int{x, y, z}, rotation, structureMirrorNone, cube.Pos{}),
						state:    applyPlacedStructureStateTransform(state, structureMirrorNone, rotation),
					})
				}
			}
		}
	}
	place := func(pos [3]int, state gen.BlockState) {
		blocks = append(blocks, plannedStructureBlock{
			worldPos: structureTemplateWorldPos(origin, pos, rotation, structureMirrorNone, cube.Pos{}),
			state:    applyPlacedStructureStateTransform(state, structureMirrorNone, rotation),
		})
	}
	sprucePlanks := gen.BlockState{Name: "spruce_planks"}
	oakLog := gen.BlockState{Name: "oak_log", Properties: map[string]string{"pillar_axis": "y"}}
	oakFence := gen.BlockState{Name: "oak_fence"}
	pottedMushroom := gen.BlockState{Name: "potted_red_mushroom"}
	craftingTable := gen.BlockState{Name: "crafting_table"}
	cauldron := gen.BlockState{Name: "cauldron"}
	air := gen.BlockState{Name: "air"}
	northStairs := gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "north"}}
	eastStairs := gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "east"}}
	westStairs := gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "west"}}
	southStairs := gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "south"}}

	fillBox([3]int{1, 1, 1}, [3]int{5, 1, 7}, sprucePlanks)
	fillBox([3]int{1, 4, 2}, [3]int{5, 4, 7}, sprucePlanks)
	fillBox([3]int{2, 1, 0}, [3]int{4, 1, 0}, sprucePlanks)
	fillBox([3]int{2, 2, 2}, [3]int{3, 3, 2}, sprucePlanks)
	fillBox([3]int{1, 2, 3}, [3]int{1, 3, 6}, sprucePlanks)
	fillBox([3]int{5, 2, 3}, [3]int{5, 3, 6}, sprucePlanks)
	fillBox([3]int{2, 2, 7}, [3]int{4, 3, 7}, sprucePlanks)
	fillBox([3]int{1, 0, 2}, [3]int{1, 3, 2}, oakLog)
	fillBox([3]int{5, 0, 2}, [3]int{5, 3, 2}, oakLog)
	fillBox([3]int{1, 0, 7}, [3]int{1, 3, 7}, oakLog)
	fillBox([3]int{5, 0, 7}, [3]int{5, 3, 7}, oakLog)
	place([3]int{2, 3, 2}, oakFence)
	place([3]int{3, 3, 7}, oakFence)
	place([3]int{1, 3, 4}, air)
	place([3]int{5, 3, 4}, air)
	place([3]int{5, 3, 5}, air)
	place([3]int{1, 3, 5}, pottedMushroom)
	place([3]int{3, 2, 6}, craftingTable)
	place([3]int{4, 2, 6}, cauldron)
	place([3]int{1, 2, 1}, oakFence)
	place([3]int{5, 2, 1}, oakFence)
	fillBox([3]int{0, 4, 1}, [3]int{6, 4, 1}, northStairs)
	fillBox([3]int{0, 4, 2}, [3]int{0, 4, 7}, eastStairs)
	fillBox([3]int{6, 4, 2}, [3]int{6, 4, 7}, westStairs)
	fillBox([3]int{0, 4, 8}, [3]int{6, 4, 8}, southStairs)
	for _, entry := range []struct {
		pos   [3]int
		state gen.BlockState
	}{
		{pos: [3]int{0, 4, 1}, state: gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "east"}}},
		{pos: [3]int{6, 4, 1}, state: gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "west"}}},
		{pos: [3]int{0, 4, 8}, state: gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "east"}}},
		{pos: [3]int{6, 4, 8}, state: gen.BlockState{Name: "spruce_stairs", Properties: map[string]string{"facing": "west"}}},
	} {
		place(entry.pos, entry.state)
	}
	for _, z := range []int{2, 7} {
		for _, x := range []int{1, 5} {
			worldPos := structureTemplateWorldPos(origin, [3]int{x, -1, z}, rotation, structureMirrorNone, cube.Pos{})
			groundY := g.preliminarySurfaceLevelAt(worldPos[0], worldPos[2], minY, maxY)
			for y := worldPos[1]; y >= groundY-1; y-- {
				blocks = append(blocks, plannedStructureBlock{
					worldPos: cube.Pos{worldPos[0], y, worldPos[2]},
					state:    applyPlacedStructureStateTransform(oakLog, structureMirrorNone, rotation),
				})
			}
		}
	}
	return blocks
}

func normalizeStructureNameList(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, normalizeStructureName(strings.TrimPrefix(name, "minecraft:")))
	}
	return out
}

var (
	shipwreckBeachedTemplates = normalizeStructureNameList([]string{
		"shipwreck/with_mast",
		"shipwreck/sideways_full",
		"shipwreck/sideways_fronthalf",
		"shipwreck/sideways_backhalf",
		"shipwreck/rightsideup_full",
		"shipwreck/rightsideup_fronthalf",
		"shipwreck/rightsideup_backhalf",
		"shipwreck/with_mast_degraded",
		"shipwreck/rightsideup_full_degraded",
		"shipwreck/rightsideup_fronthalf_degraded",
		"shipwreck/rightsideup_backhalf_degraded",
	})
	shipwreckOceanTemplates = normalizeStructureNameList([]string{
		"shipwreck/with_mast", "shipwreck/upsidedown_full", "shipwreck/upsidedown_fronthalf", "shipwreck/upsidedown_backhalf",
		"shipwreck/sideways_full", "shipwreck/sideways_fronthalf", "shipwreck/sideways_backhalf", "shipwreck/rightsideup_full",
		"shipwreck/rightsideup_fronthalf", "shipwreck/rightsideup_backhalf", "shipwreck/with_mast_degraded",
		"shipwreck/upsidedown_full_degraded", "shipwreck/upsidedown_fronthalf_degraded", "shipwreck/upsidedown_backhalf_degraded",
		"shipwreck/sideways_full_degraded", "shipwreck/sideways_fronthalf_degraded", "shipwreck/sideways_backhalf_degraded",
		"shipwreck/rightsideup_full_degraded", "shipwreck/rightsideup_fronthalf_degraded", "shipwreck/rightsideup_backhalf_degraded",
	})
	oceanRuinWarmTemplates             = normalizeStructureNameList([]string{"underwater_ruin/warm_1", "underwater_ruin/warm_2", "underwater_ruin/warm_3", "underwater_ruin/warm_4", "underwater_ruin/warm_5", "underwater_ruin/warm_6", "underwater_ruin/warm_7", "underwater_ruin/warm_8"})
	oceanRuinWarmLargeTemplates        = normalizeStructureNameList([]string{"underwater_ruin/big_warm_4", "underwater_ruin/big_warm_5", "underwater_ruin/big_warm_6", "underwater_ruin/big_warm_7"})
	oceanRuinColdBrickTemplates        = normalizeStructureNameList([]string{"underwater_ruin/brick_1", "underwater_ruin/brick_2", "underwater_ruin/brick_3", "underwater_ruin/brick_4", "underwater_ruin/brick_5", "underwater_ruin/brick_6", "underwater_ruin/brick_7", "underwater_ruin/brick_8"})
	oceanRuinColdCrackedTemplates      = normalizeStructureNameList([]string{"underwater_ruin/cracked_1", "underwater_ruin/cracked_2", "underwater_ruin/cracked_3", "underwater_ruin/cracked_4", "underwater_ruin/cracked_5", "underwater_ruin/cracked_6", "underwater_ruin/cracked_7", "underwater_ruin/cracked_8"})
	oceanRuinColdMossyTemplates        = normalizeStructureNameList([]string{"underwater_ruin/mossy_1", "underwater_ruin/mossy_2", "underwater_ruin/mossy_3", "underwater_ruin/mossy_4", "underwater_ruin/mossy_5", "underwater_ruin/mossy_6", "underwater_ruin/mossy_7", "underwater_ruin/mossy_8"})
	oceanRuinColdBrickLargeTemplates   = normalizeStructureNameList([]string{"underwater_ruin/big_brick_1", "underwater_ruin/big_brick_2", "underwater_ruin/big_brick_3", "underwater_ruin/big_brick_8"})
	oceanRuinColdCrackedLargeTemplates = normalizeStructureNameList([]string{"underwater_ruin/big_cracked_1", "underwater_ruin/big_cracked_2", "underwater_ruin/big_cracked_3", "underwater_ruin/big_cracked_8"})
	oceanRuinColdMossyLargeTemplates   = normalizeStructureNameList([]string{"underwater_ruin/big_mossy_1", "underwater_ruin/big_mossy_2", "underwater_ruin/big_mossy_3", "underwater_ruin/big_mossy_8"})
	ruinedPortalTemplates              = normalizeStructureNameList([]string{
		"ruined_portal/portal_1", "ruined_portal/portal_2", "ruined_portal/portal_3", "ruined_portal/portal_4", "ruined_portal/portal_5",
		"ruined_portal/portal_6", "ruined_portal/portal_7", "ruined_portal/portal_8", "ruined_portal/portal_9", "ruined_portal/portal_10",
	})
	ruinedPortalGiantTemplates = normalizeStructureNameList([]string{"ruined_portal/giant_portal_1", "ruined_portal/giant_portal_2", "ruined_portal/giant_portal_3"})
	netherFossilTemplates      = normalizeStructureNameList([]string{
		"nether_fossils/fossil_1",
		"nether_fossils/fossil_2",
		"nether_fossils/fossil_3",
		"nether_fossils/fossil_4",
		"nether_fossils/fossil_5",
		"nether_fossils/fossil_6",
		"nether_fossils/fossil_7",
		"nether_fossils/fossil_8",
		"nether_fossils/fossil_9",
		"nether_fossils/fossil_10",
		"nether_fossils/fossil_11",
		"nether_fossils/fossil_12",
		"nether_fossils/fossil_13",
		"nether_fossils/fossil_14",
	})
)
