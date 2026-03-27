package vanilla

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

type endCityBridgeAttachment struct {
	rotation structureRotation
	offset   cube.Pos
}

type endCityState struct {
	shipCreated bool
}

type endCitySectionGenerator func(g Generator, genDepth int, parent plannedStructurePiece, offset cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool

var (
	endCityTowerBridges = []endCityBridgeAttachment{
		{rotation: structureRotationNone, offset: cube.Pos{1, -1, 0}},
		{rotation: structureRotationClockwise90, offset: cube.Pos{6, -1, 1}},
		{rotation: structureRotationCounterclockwise90, offset: cube.Pos{0, -1, 5}},
		{rotation: structureRotationClockwise180, offset: cube.Pos{5, -1, 6}},
	}
	endCityFatTowerBridges = []endCityBridgeAttachment{
		{rotation: structureRotationNone, offset: cube.Pos{4, -1, 0}},
		{rotation: structureRotationClockwise90, offset: cube.Pos{12, -1, 4}},
		{rotation: structureRotationCounterclockwise90, offset: cube.Pos{0, -1, 8}},
		{rotation: structureRotationClockwise180, offset: cube.Pos{8, -1, 12}},
	}
)

func (g Generator) buildEndCityStructure(startChunk world.ChunkPos, _, _ int, minY, maxY int, rng *gen.Xoroshiro128) (string, []plannedStructurePiece, structureBox, cube.Pos, [3]int, bool) {
	rotation := randomStructureRotation(rng)
	startPos, ok := g.endCityStartPos(startChunk, rotation, minY, maxY)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}

	pieces := make([]plannedStructurePiece, 0, 10)
	root, ok := g.newEndCityRootPiece("base_floor", startPos, rotation, true)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	root.rootPiece = true
	pieces = append(pieces, root)

	lastPiece, ok := g.addEndCityPiece(root, cube.Pos{-1, 0, -1}, "second_floor_1", rotation, false)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "third_floor_1", rotation, false)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 8, -1}, "third_roof", rotation, true)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)

	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{3 + int(rng.NextInt(2)), -3, 3 + int(rng.NextInt(2))}, "tower_base", rotation, true)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 7, 0}, "tower_piece", rotation, true)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)
	for i := 0; i < 1+int(rng.NextInt(2)); i++ {
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 4, 0}, "tower_piece", rotation, true)
		if !ok {
			return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
		}
		pieces = append(pieces, lastPiece)
	}
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "tower_top", rotation, true)
	if !ok {
		return "", nil, emptyStructureBox(), cube.Pos{}, [3]int{}, false
	}
	pieces = append(pieces, lastPiece)

	overall := emptyStructureBox()
	for _, piece := range pieces {
		overall = unionStructureBoxes(overall, piece.bounds)
	}
	rootOrigin, rootSize := root.bounds.originAndSize()
	return "end_city/base_floor", pieces, overall, rootOrigin, rootSize, true
}

func (g Generator) endCityStartPos(startChunk world.ChunkPos, rotation structureRotation, minY, maxY int) (cube.Pos, bool) {
	blockX := int(startChunk[0])*16 + 7
	blockZ := int(startChunk[1])*16 + 7
	offsetX, offsetZ := 5, 5
	switch rotation {
	case structureRotationClockwise90:
		offsetX = -5
	case structureRotationClockwise180:
		offsetX, offsetZ = -5, -5
	case structureRotationCounterclockwise90:
		offsetZ = -5
	}

	y := min(
		min(g.highestStructureSolidYAt(blockX, blockZ, minY, maxY), g.highestStructureSolidYAt(blockX+offsetX, blockZ, minY, maxY)),
		min(g.highestStructureSolidYAt(blockX, blockZ+offsetZ, minY, maxY), g.highestStructureSolidYAt(blockX+offsetX, blockZ+offsetZ, minY, maxY)),
	)
	if y <= minY {
		y = clamp(g.preliminarySurfaceLevelAt(blockX, blockZ, minY, maxY), minY+1, maxY)
	}
	if y <= minY {
		y = min(maxY, max(64, minY+1))
	}
	return cube.Pos{blockX, y, blockZ}, true
}

func (g Generator) newEndCityRootPiece(templateName string, origin cube.Pos, rotation structureRotation, overwrite bool) (plannedStructurePiece, bool) {
	piece, ok := g.newTemplateStructurePiece("end_city/"+templateName, origin, rotation, structureMirrorNone, cube.Pos{}, !overwrite, nil)
	if !ok {
		return plannedStructurePiece{}, false
	}
	return piece, true
}

func (g Generator) addEndCityPiece(parent plannedStructurePiece, offset cube.Pos, templateName string, rotation structureRotation, overwrite bool) (plannedStructurePiece, bool) {
	origin := structureTemplateWorldPos(parent.origin, [3]int{offset[0], offset[1], offset[2]}, parent.rotation, structureMirrorNone, cube.Pos{})
	return g.newEndCityRootPiece(templateName, origin, rotation, overwrite)
}

func (g Generator) endCityRecursiveChildren(generator endCitySectionGenerator, genDepth int, parent plannedStructurePiece, offset cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool {
	if genDepth > 8 {
		return false
	}

	childPieces := make([]plannedStructurePiece, 0, 16)
	if !generator(g, genDepth, parent, offset, &childPieces, rng) {
		return false
	}

	childTag := int(rng.NextLong())
	for i := range childPieces {
		childPieces[i].genTag = childTag
	}
	for _, child := range childPieces {
		if collision, ok := endCityFindCollision(*pieces, child.bounds); ok && collision.genTag != parent.genTag {
			return false
		}
	}
	*pieces = append(*pieces, childPieces...)
	return true
}

func endCityFindCollision(pieces []plannedStructurePiece, bounds structureBox) (plannedStructurePiece, bool) {
	for _, piece := range pieces {
		if piece.bounds.intersects(bounds) {
			return piece, true
		}
	}
	return plannedStructurePiece{}, false
}

func rotationAdd(a, b structureRotation) structureRotation {
	return structureRotation((int(a) + int(b)) & 3)
}

func (g Generator) highestStructureSolidYAt(blockX, blockZ, minY, maxY int) int {
	chunkX := floorDiv(blockX, 16)
	chunkZ := floorDiv(blockZ, 16)
	flat := g.graph.NewFlatCacheGrid(chunkX, chunkZ, g.noises)
	col := g.graph.NewColumnContext(blockX, blockZ, g.noises, flat)
	guess := clamp(g.preliminarySurfaceLevelAt(blockX, blockZ, minY, maxY), minY, maxY)
	low := max(minY, guess-96)
	high := min(maxY, guess+96)
	for y := high; y >= low; y-- {
		if g.sampleStructureSubstanceAtWithColumn(blockX, y, blockZ, flat, col) == structureSubstanceSolid {
			return y
		}
	}
	return minY
}

func (s *endCityState) generateHouseTower(g Generator, genDepth int, parent plannedStructurePiece, offset cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool {
	rotation := parent.rotation
	lastPiece, ok := g.addEndCityPiece(parent, offset, "base_floor", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)

	switch floors := int(rng.NextInt(3)); floors {
	case 0:
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "base_roof", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		return true
	case 1:
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 0, -1}, "second_floor_2", rotation, false)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 8, -1}, "second_roof", rotation, false)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		return g.endCityRecursiveChildren(s.generateTower, genDepth+1, lastPiece, cube.Pos{}, pieces, rng)
	default:
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 0, -1}, "second_floor_2", rotation, false)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "third_floor_2", rotation, false)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 8, -1}, "third_roof", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		return g.endCityRecursiveChildren(s.generateTower, genDepth+1, lastPiece, cube.Pos{}, pieces, rng)
	}
}

func (s *endCityState) generateTower(g Generator, genDepth int, parent plannedStructurePiece, _ cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool {
	rotation := parent.rotation
	lastPiece, ok := g.addEndCityPiece(parent, cube.Pos{3 + int(rng.NextInt(2)), -3, 3 + int(rng.NextInt(2))}, "tower_base", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 7, 0}, "tower_piece", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)

	var bridgePiece *plannedStructurePiece
	if rng.NextInt(3) == 0 {
		candidate := lastPiece
		bridgePiece = &candidate
	}

	towerHeight := 1 + int(rng.NextInt(3))
	for i := 0; i < towerHeight; i++ {
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 4, 0}, "tower_piece", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		if i < towerHeight-1 && rng.NextInt(2) == 0 {
			candidate := lastPiece
			bridgePiece = &candidate
		}
	}

	if bridgePiece != nil {
		for _, bridge := range endCityTowerBridges {
			if rng.NextInt(2) != 0 {
				continue
			}
			bridgeStart, ok := g.addEndCityPiece(*bridgePiece, bridge.offset, "bridge_end", rotationAdd(rotation, bridge.rotation), true)
			if !ok {
				continue
			}
			if !g.endCityRecursiveChildren(s.generateTowerBridge, genDepth+1, bridgeStart, cube.Pos{}, pieces, rng) {
				return false
			}
		}

		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "tower_top", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
		return true
	}

	if genDepth != 7 {
		return g.endCityRecursiveChildren(s.generateFatTower, genDepth+1, lastPiece, cube.Pos{}, pieces, rng)
	}

	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-1, 4, -1}, "tower_top", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)
	return true
}

func (s *endCityState) generateTowerBridge(g Generator, genDepth int, parent plannedStructurePiece, _ cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool {
	rotation := parent.rotation
	bridgeLength := int(rng.NextInt(4)) + 1
	lastPiece, ok := g.addEndCityPiece(parent, cube.Pos{0, 0, -4}, "bridge_piece", rotation, true)
	if !ok {
		return false
	}
	lastPiece.genTag = -1
	*pieces = append(*pieces, lastPiece)
	nextY := 0

	for i := 0; i < bridgeLength; i++ {
		if rng.NextInt(2) == 0 {
			lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, nextY, -4}, "bridge_piece", rotation, true)
			nextY = 0
		} else {
			if rng.NextInt(2) == 0 {
				lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, nextY, -4}, "bridge_steep_stairs", rotation, true)
			} else {
				lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, nextY, -8}, "bridge_gentle_stairs", rotation, true)
			}
			nextY = 4
		}
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)
	}

	if !s.shipCreated && rng.NextInt(uint32(max(1, 10-genDepth))) == 0 {
		shipPiece, ok := g.addEndCityPiece(lastPiece, cube.Pos{-8 + int(rng.NextInt(8)), nextY, -70 + int(rng.NextInt(10))}, "ship", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, shipPiece)
		s.shipCreated = true
	} else if !g.endCityRecursiveChildren(s.generateHouseTower, genDepth+1, lastPiece, cube.Pos{-3, nextY + 1, -11}, pieces, rng) {
		return false
	}

	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{4, nextY, 0}, "bridge_end", rotationAdd(rotation, structureRotationClockwise180), true)
	if !ok {
		return false
	}
	lastPiece.genTag = -1
	*pieces = append(*pieces, lastPiece)
	return true
}

func (s *endCityState) generateFatTower(g Generator, genDepth int, parent plannedStructurePiece, _ cube.Pos, pieces *[]plannedStructurePiece, rng *gen.Xoroshiro128) bool {
	rotation := parent.rotation
	lastPiece, ok := g.addEndCityPiece(parent, cube.Pos{-3, 4, -3}, "fat_tower_base", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)
	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 4, 0}, "fat_tower_middle", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)

	for i := 0; i < 2 && rng.NextInt(3) != 0; i++ {
		lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{0, 8, 0}, "fat_tower_middle", rotation, true)
		if !ok {
			return false
		}
		*pieces = append(*pieces, lastPiece)

		for _, bridge := range endCityFatTowerBridges {
			if rng.NextInt(2) != 0 {
				continue
			}
			bridgeStart, ok := g.addEndCityPiece(lastPiece, bridge.offset, "bridge_end", rotationAdd(rotation, bridge.rotation), true)
			if !ok {
				continue
			}
			if !g.endCityRecursiveChildren(s.generateTowerBridge, genDepth+1, bridgeStart, cube.Pos{}, pieces, rng) {
				return false
			}
		}
	}

	lastPiece, ok = g.addEndCityPiece(lastPiece, cube.Pos{-2, 8, -2}, "fat_tower_top", rotation, true)
	if !ok {
		return false
	}
	*pieces = append(*pieces, lastPiece)
	return true
}
