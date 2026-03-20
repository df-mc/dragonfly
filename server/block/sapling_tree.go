package block

import (
	"math"
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// growTree selects the local tree generator that matches the sapling wood type.
func (s Sapling) growTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	switch s.Wood {
	case OakWood():
		if r.Float64() < 0.1 {
			return s.growFancyOak(pos, tx, r)
		}
		return s.growStraightBlob(pos, tx, r, 4, 2, 0)
	case SpruceWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			if r.Float64() < 0.5 {
				return s.growMegaPine(origin, tx, r)
			}
			return s.growMegaSpruce(origin, tx, r)
		}
		return s.growSpruce(pos, tx, r)
	case BirchWood():
		return s.growStraightBlob(pos, tx, r, 5, 2, 0)
	case JungleWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			return s.growMegaJungle(origin, tx, r)
		}
		return s.growStraightBlob(pos, tx, r, 4, 8, 0)
	case AcaciaWood():
		return s.growAcacia(pos, tx, r)
	case DarkOakWood(), PaleOakWood():
		origin, ok := s.twoByTwoOrigin(pos, tx)
		if !ok {
			return false
		}
		return s.growDarkOak(origin, tx, r)
	case CherryWood():
		return s.growCherry(pos, tx, r)
	case MangroveWood():
		if r.Float64() < 0.85 {
			return s.growTallMangrove(pos, tx, r)
		}
		return s.growMangrove(pos, tx, r)
	default:
		return false
	}
}

// growStraightBlob places the straight-trunk blob-canopy trees used by oak, birch, and small jungle saplings.
func (s Sapling) growStraightBlob(pos cube.Pos, tx *world.Tx, r *rand.Rand, baseHeight, heightRandA, heightRandB int) bool {
	height := treeHeight(r, baseHeight, heightRandA, heightRandB)
	if !s.canGrowOnTreeBase(pos, tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, pos, func(_, currentHeight int) int {
		if currentHeight < 1 {
			return 0
		}
		return 1
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, maxFreeHeight, s.Wood)
	layout.blobFoliage(foliageAttachment{pos: pos.Add(cube.Pos{0, maxFreeHeight, 0})}, 3, 2, 0, s.Wood, r)
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(pos.Side(cube.FaceDown), tx)
	return true
}

// growFancyOak places the branching fancy oak tree shape.
func (s Sapling) growFancyOak(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 3, 11, 0)
	if !s.canGrowOnTreeBase(pos, tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, pos, func(_, _ int) int {
		return 0
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height && maxFreeHeight < 4 {
		return false
	}

	trunkAndFoliageHeight := maxFreeHeight + 2
	trunkTopOffset := int(math.Floor(float64(trunkAndFoliageHeight) * 0.618))
	branchBaseLimit := pos[1] + trunkTopOffset
	foliageStart := trunkAndFoliageHeight - 5
	layout := newSaplingTreeLayout(tx)
	foliageCoords := []fancyFoliageCoord{{attachment: pos.Add(cube.Pos{0, foliageStart, 0}), branchBase: branchBaseLimit}}

	for current := foliageStart; current >= 0; current-- {
		treeShape := fancyTreeShape(trunkAndFoliageHeight, current)
		if treeShape < 0 {
			continue
		}

		branchLength := treeShape * (r.Float64() + 0.328)
		angle := r.Float64() * math.Pi * 2
		branchX := pos[0] + int(math.Floor(branchLength*math.Sin(angle)+0.5))
		branchZ := pos[2] + int(math.Floor(branchLength*math.Cos(angle)+0.5))
		foliagePos := cube.Pos{branchX, pos[1] + current - 1, branchZ}
		foliageCheckTop := cube.Pos{branchX, pos[1] + current + 4, branchZ}
		if !layout.limb(foliagePos, foliageCheckTop, s.Wood, false) {
			continue
		}

		xDiff, zDiff := pos[0]-branchX, pos[2]-branchZ
		branchBaseY := float64(foliagePos[1]) - math.Sqrt(float64(xDiff*xDiff+zDiff*zDiff))*0.381
		attachmentBaseY := int(branchBaseY)
		if branchBaseY > float64(branchBaseLimit) {
			attachmentBaseY = branchBaseLimit
		}
		branchBase := cube.Pos{pos[0], attachmentBaseY, pos[2]}
		if layout.limb(branchBase, foliagePos, s.Wood, false) {
			foliageCoords = append(foliageCoords, fancyFoliageCoord{attachment: foliagePos, branchBase: attachmentBaseY})
		}
	}

	layout.limb(pos, pos.Add(cube.Pos{0, trunkTopOffset, 0}), s.Wood, true)
	for _, foliageCoord := range foliageCoords {
		if !trimFancyBranch(trunkAndFoliageHeight, foliageCoord.branchBase-pos[1]) {
			continue
		}
		branchStart := cube.Pos{pos[0], foliageCoord.branchBase, pos[2]}
		if branchStart != foliageCoord.attachment {
			layout.limb(branchStart, foliageCoord.attachment, s.Wood, true)
		}
	}
	for _, foliageCoord := range foliageCoords {
		if !trimFancyBranch(trunkAndFoliageHeight, foliageCoord.branchBase-pos[1]) {
			continue
		}
		layout.fancyFoliage(foliageAttachment{pos: foliageCoord.attachment}, 4, 2, 4, s.Wood)
	}

	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(pos.Side(cube.FaceDown), tx)
	return true
}

// growSpruce places a single spruce using the straight-trunk and spruce-foliage rules.
func (s Sapling) growSpruce(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 5, 2, 1)
	if !s.canGrowOnTreeBase(pos, tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, pos, func(_, currentHeight int) int {
		if currentHeight < 2 {
			return 0
		}
		return 2
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	foliageHeight := max(4, maxFreeHeight-(1+r.IntN(2)))
	foliageRadius := 2 + r.IntN(2)
	offset := r.IntN(3)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, maxFreeHeight, s.Wood)
	layout.spruceFoliage(foliageAttachment{pos: pos.Add(cube.Pos{0, maxFreeHeight, 0})}, foliageHeight, foliageRadius, offset, s.Wood, r)
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(pos.Side(cube.FaceDown), tx)
	return true
}

// growAcacia places an acacia using the forking trunk and acacia foliage rules.
func (s Sapling) growAcacia(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 5, 2, 2)
	if !s.canGrowOnTreeBase(pos, tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, pos, func(_, currentHeight int) int {
		if currentHeight < 1 {
			return 0
		}
		return 2
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	layout := newSaplingTreeLayout(tx)
	attachments := make([]foliageAttachment, 0, 2)
	horizontal := cube.Directions()
	primaryDirection := horizontal[r.IntN(len(horizontal))]
	bendStart := maxFreeHeight - r.IntN(4) - 1
	bendLength := 3 - r.IntN(3)

	currentX, currentZ := pos[0], pos[2]
	lastPlacedPos := pos
	for dy := 0; dy < maxFreeHeight; dy++ {
		logY := pos[1] + dy
		if dy >= bendStart && bendLength > 0 {
			step := offset(primaryDirection, 1)
			currentX += step[0]
			currentZ += step[2]
			bendLength--
		}
		if layout.setIfValid(cube.Pos{currentX, logY, currentZ}, Log{Wood: s.Wood, Axis: cube.Y}) {
			lastPlacedPos = cube.Pos{currentX, logY, currentZ}
		}
	}
	attachments = append(attachments, foliageAttachment{pos: lastPlacedPos.Add(cube.Pos{0, 1, 0}), radiusOffset: 1})

	secondaryDirection := horizontal[r.IntN(len(horizontal))]
	if secondaryDirection != primaryDirection {
		secondaryStart := bendStart - r.IntN(2) - 1
		secondaryLength := 1 + r.IntN(3)
		currentX, currentZ = pos[0], pos[2]
		var secondaryTopPos cube.Pos
		secondaryPlaced := false
		for dy := secondaryStart; dy < maxFreeHeight && secondaryLength > 0; secondaryLength, dy = secondaryLength-1, dy+1 {
			if dy >= 1 {
				logY := pos[1] + dy
				step := offset(secondaryDirection, 1)
				currentX += step[0]
				currentZ += step[2]
				if layout.setIfValid(cube.Pos{currentX, logY, currentZ}, Log{Wood: s.Wood, Axis: cube.Y}) {
					secondaryTopPos = cube.Pos{currentX, logY, currentZ}
					secondaryPlaced = true
				}
			}
		}
		if secondaryPlaced {
			attachments = append(attachments, foliageAttachment{pos: secondaryTopPos.Add(cube.Pos{0, 1, 0})})
		}
	}
	for _, attachment := range attachments {
		layout.acaciaFoliage(attachment, 2, 0, s.Wood)
	}
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(pos.Side(cube.FaceDown), tx)
	return true
}

// growDarkOak places the leaning 2x2 trunk and side branch stubs used by dark oak and pale oak saplings.
func (s Sapling) growDarkOak(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 6, 2, 1)
	if !s.canGrowOnTreeBase(origin, tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 0}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{0, 0, 1}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 1}), tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, origin, func(treeHeight, currentHeight int) int {
		if currentHeight < 1 {
			return 0
		}
		if currentHeight >= treeHeight-1 {
			return 2
		}
		return 1
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	layout := newSaplingTreeLayout(tx)
	horizontal := cube.Directions()
	bendDirection := horizontal[r.IntN(len(horizontal))]
	bendStart := maxFreeHeight - r.IntN(4)
	bendLength := 2 - r.IntN(3)
	currentX, currentZ := origin[0], origin[2]
	topY := origin[1] + maxFreeHeight - 1

	for dy := 0; dy < maxFreeHeight; dy++ {
		if dy >= bendStart && bendLength > 0 {
			step := offset(bendDirection, 1)
			currentX += step[0]
			currentZ += step[2]
			bendLength--
		}
		logY := origin[1] + dy
		if isAirOrLeaves(tx, cube.Pos{currentX, logY, currentZ}) {
			layout.setIfValid(cube.Pos{currentX, logY, currentZ}, Log{Wood: s.Wood, Axis: cube.Y})
			layout.setIfValid(cube.Pos{currentX + 1, logY, currentZ}, Log{Wood: s.Wood, Axis: cube.Y})
			layout.setIfValid(cube.Pos{currentX, logY, currentZ + 1}, Log{Wood: s.Wood, Axis: cube.Y})
			layout.setIfValid(cube.Pos{currentX + 1, logY, currentZ + 1}, Log{Wood: s.Wood, Axis: cube.Y})
		}
	}

	attachments := []foliageAttachment{{pos: cube.Pos{currentX, topY, currentZ}, doubleTrunk: true}}

	for dx := -1; dx <= 2; dx++ {
		for dz := -1; dz <= 2; dz++ {
			if dx >= 0 && dx <= 1 && dz >= 0 && dz <= 1 {
				continue
			}
			if r.IntN(3) > 0 {
				continue
			}
			length := 2 + r.IntN(3)
			for i := 0; i < length; i++ {
				layout.setIfValid(cube.Pos{origin[0] + dx, topY - i - 1, origin[2] + dz}, Log{Wood: s.Wood, Axis: cube.Y})
			}
			attachments = append(attachments, foliageAttachment{pos: cube.Pos{origin[0] + dx, topY, origin[2] + dz}})
		}
	}
	for _, attachment := range attachments {
		layout.darkOakFoliage(attachment, 0, 0, s.Wood, r)
	}
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(origin.Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 0}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{0, 0, 1}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 1}).Side(cube.FaceDown), tx)
	return true
}

// growCherry places the branching cherry tree shape.
func (s Sapling) growCherry(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 7, 1, 0)
	if !s.canGrowOnTreeBase(pos, tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, pos, func(_, currentHeight int) int {
		if currentHeight < 1 {
			return 0
		}
		return 2
	}, true, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	firstBranchStart := max(0, maxFreeHeight-1+sampleUniform(r, -4, -3))
	secondBranchStart := max(0, maxFreeHeight-5)
	if secondBranchStart >= firstBranchStart {
		secondBranchStart++
	}

	branchCount := r.IntN(3) + 1
	placeTopAttachment := branchCount == 3
	placeSecondBranch := branchCount >= 2
	trunkHeight := firstBranchStart + 1
	if placeTopAttachment {
		trunkHeight = maxFreeHeight
	} else if placeSecondBranch {
		trunkHeight = max(firstBranchStart, secondBranchStart) + 1
	}

	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, trunkHeight, s.Wood)
	attachments := make([]foliageAttachment, 0, 3)
	if placeTopAttachment {
		attachments = append(attachments, foliageAttachment{pos: pos.Add(cube.Pos{0, trunkHeight, 0})})
	}

	horizontal := cube.Directions()
	branchDirection := horizontal[r.IntN(len(horizontal))]
	attachments = append(attachments, layout.cherryBranch(pos, maxFreeHeight, branchDirection, firstBranchStart, firstBranchStart < trunkHeight-1, s.Wood, r))
	if placeSecondBranch {
		attachments = append(attachments, layout.cherryBranch(pos, maxFreeHeight, branchDirection.Opposite(), secondBranchStart, secondBranchStart < trunkHeight-1, s.Wood, r))
	}

	for _, attachment := range attachments {
		layout.cherryFoliage(attachment, 5, 4, 0, 0.25, 0.5, 1.0/6.0, 1.0/3.0, s.Wood, r)
	}
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(pos.Side(cube.FaceDown), tx)
	return true
}

// growMegaSpruce places the mega spruce shape selected by half of 2x2 spruce growth attempts.
func (s Sapling) growMegaSpruce(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMegaConifer(origin, tx, r, 13, 2, 14, 13, 17)
}

// growMegaPine places the mega pine shape selected by the other half of 2x2 spruce growth attempts.
func (s Sapling) growMegaPine(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMegaConifer(origin, tx, r, 13, 2, 14, 3, 7)
}

// growMegaConifer places a giant spruce-family tree using the giant-trunk, mega-pine foliage, and podzol decorator rules.
func (s Sapling) growMegaConifer(origin cube.Pos, tx *world.Tx, r *rand.Rand, baseHeight, heightRandA, heightRandB, crownMin, crownMax int) bool {
	height := treeHeight(r, baseHeight, heightRandA, heightRandB)
	if !s.canGrowOnTreeBase(origin, tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 0}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{0, 0, 1}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 1}), tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, origin, func(_, currentHeight int) int {
		if currentHeight < 1 {
			return 1
		}
		return 2
	}, false, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	crownHeight := crownMin + r.IntN(crownMax-crownMin+1)
	layout := newSaplingTreeLayout(tx)
	layout.giantTrunk(origin, maxFreeHeight, s.Wood)
	layout.megaPineFoliage(foliageAttachment{pos: origin.Add(cube.Pos{0, maxFreeHeight, 0}), doubleTrunk: true}, crownHeight, 0, 0, s.Wood)
	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(origin.Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 0}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{0, 0, 1}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 1}).Side(cube.FaceDown), tx)
	s.placeMegaConiferPodzol(origin, tx, r)
	return true
}

// growMegaJungle places a giant jungle tree with side canopies and added trunk vines.
func (s Sapling) growMegaJungle(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 10, 2, 19)
	if !s.canGrowOnTreeBase(origin, tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 0}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{0, 0, 1}), tx) || !s.canGrowOnTreeBase(origin.Add(cube.Pos{1, 0, 1}), tx) {
		return false
	}
	maxFreeHeight := maxFreeTreeHeight(tx, height, origin, func(_, currentHeight int) int {
		if currentHeight < 1 {
			return 1
		}
		return 2
	}, false, isFreeTreeBlock)
	if maxFreeHeight < height {
		return false
	}

	layout := newSaplingTreeLayout(tx)
	layout.giantTrunk(origin, maxFreeHeight, s.Wood)
	attachments := []foliageAttachment{{pos: origin.Add(cube.Pos{0, maxFreeHeight, 0}), doubleTrunk: true}}

	for branchStart := maxFreeHeight - 2 - r.IntN(4); branchStart > maxFreeHeight/2; branchStart -= 2 + r.IntN(4) {
		angle := r.Float64() * math.Pi * 2
		branchX, branchZ := 0, 0
		for i := 0; i < 5; i++ {
			branchX = int(1.5 + math.Cos(angle)*float64(i))
			branchZ = int(1.5 + math.Sin(angle)*float64(i))
			layout.setIfValid(cube.Pos{origin[0] + branchX, origin[1] + branchStart - 3 + i/2, origin[2] + branchZ}, Log{Wood: s.Wood, Axis: cube.Y})
		}
		attachments = append(attachments, foliageAttachment{pos: cube.Pos{origin[0] + branchX, origin[1] + branchStart, origin[2] + branchZ}, radiusOffset: -2})
	}
	for _, attachment := range attachments {
		layout.megaJungleFoliage(attachment, 2, 2, 0, s.Wood, r)
	}

	if !layout.apply() {
		return false
	}
	s.placeBelowOverworldTrunk(origin.Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 0}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{0, 0, 1}).Side(cube.FaceDown), tx)
	s.placeBelowOverworldTrunk(origin.Add(cube.Pos{1, 0, 1}).Side(cube.FaceDown), tx)
	s.placeTrunkVines(tx, layout.blocks, r)
	s.placeLeafVines(tx, layout.blocks, r, 0.25)
	return true
}

// growMangrove places the shorter mangrove feature used by the primary mangrove tree selection.
func (s Sapling) growMangrove(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMangroveVariant(pos, tx, r, 2, 1, 4, 1, 4)
}

// growTallMangrove places the taller mangrove selected by most mangrove growth attempts.
func (s Sapling) growTallMangrove(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMangroveVariant(pos, tx, r, 4, 1, 9, 1, 6)
}

// growMangroveVariant places a branching mangrove trunk and decorates it with propagules and muddy roots.
func (s Sapling) growMangroveVariant(pos cube.Pos, tx *world.Tx, r *rand.Rand, baseHeight, heightRandA, heightRandB, branchMin, branchMax int) bool {
	height := treeHeight(r, baseHeight, heightRandA, heightRandB)
	layout := newSaplingTreeLayout(tx)

	layout.verticalTrunk(pos, height, s.Wood)
	top := pos.Add(cube.Pos{0, height, 0})
	layout.blobCanopy(top, 3, s.Wood)
	layout.leafLayer(top.Add(cube.Pos{0, 1, 0}), 2, s.Wood, true)

	branchCount := 2 + r.IntN(2)
	for i := 0; i < branchCount; i++ {
		dir := cube.Directions()[r.IntN(len(cube.Directions()))]
		startY := max(2, height-3-r.IntN(3))
		length := branchMin + r.IntN(branchMax-branchMin+1)
		branchTop := layout.branch(pos.Add(cube.Pos{0, startY, 0}), dir, length, 1+r.IntN(2), s.Wood)
		layout.blobCanopy(branchTop, 2, s.Wood)
	}

	s.placeMangroveMuddyRoots(pos, tx, layout)
	if !layout.apply() {
		return false
	}
	s.placeHangingPropagules(tx, layout.blocks, r, 0.14)
	return true
}

// twoByTwoOrigin finds the north-west origin of a matching 2x2 sapling arrangement.
func (s Sapling) twoByTwoOrigin(pos cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	for dx := 0; dx >= -1; dx-- {
		for dz := 0; dz >= -1; dz-- {
			if s.twoByTwo(pos.Add(cube.Pos{dx, 0, dz}), tx) {
				return pos.Add(cube.Pos{dx, 0, dz}), true
			}
		}
	}
	return cube.Pos{}, false
}

// twoByTwo checks if a 2x2 square from the origin contains matching saplings.
func (s Sapling) twoByTwo(origin cube.Pos, tx *world.Tx) bool {
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			other, ok := tx.Block(origin.Add(cube.Pos{dx, 0, dz})).(Sapling)
			if !ok || other.Wood != s.Wood || other.Hanging {
				return false
			}
		}
	}
	return true
}

// placeTrunkVines attempts to attach jungle vines to placed trunk logs.
func (s Sapling) placeTrunkVines(tx *world.Tx, blocks map[cube.Pos]world.Block, r *rand.Rand) {
	for pos, block := range blocks {
		if _, ok := block.(Log); !ok {
			continue
		}
		for _, dir := range cube.Directions() {
			if r.IntN(3) == 0 {
				continue
			}
			vinePos := pos.Add(offset(dir, 1))
			if !isAir(tx, vinePos) {
				continue
			}
			tx.SetBlock(vinePos, (Vines{}).WithAttachment(dir.Opposite(), true), nil)
		}
	}
}

// placeLeafVines attempts to grow hanging vines from placed leaves.
func (s Sapling) placeLeafVines(tx *world.Tx, blocks map[cube.Pos]world.Block, r *rand.Rand, probability float64) {
	for pos, block := range blocks {
		if _, ok := block.(Leaves); !ok {
			continue
		}
		for _, dir := range cube.Directions() {
			if r.Float64() >= probability {
				continue
			}
			vinePos := pos.Add(offset(dir, 1))
			if !isAir(tx, vinePos) {
				continue
			}
			s.addHangingVine(tx, vinePos, dir.Opposite())
		}
	}
}

// addHangingVine places a vine and extends it downward while air remains below it.
func (s Sapling) addHangingVine(tx *world.Tx, pos cube.Pos, attached cube.Direction) {
	tx.SetBlock(pos, (Vines{}).WithAttachment(attached, true), nil)
	for i := 1; i <= 4; i++ {
		under := pos.Add(cube.Pos{0, -i, 0})
		if !isAir(tx, under) {
			return
		}
		tx.SetBlock(under, (Vines{}).WithAttachment(attached, true), nil)
	}
}

// placeHangingPropagules attaches hanging propagules below generated mangrove leaves when space is available.
func (s Sapling) placeHangingPropagules(tx *world.Tx, blocks map[cube.Pos]world.Block, r *rand.Rand, chance float64) {
	exclusion := make([]cube.Pos, 0)
	for pos, block := range blocks {
		leaves, ok := block.(Leaves)
		if !ok || leaves.Wood != MangroveWood() {
			continue
		}
		under := pos.Side(cube.FaceDown)
		twoDown := under.Side(cube.FaceDown)
		if r.Float64() >= chance || !isAir(tx, under) || !isAir(tx, twoDown) || isNearExcluded(under, exclusion, 1, 0) {
			continue
		}
		exclusion = append(exclusion, under)
		tx.SetBlock(under, Sapling{Wood: MangroveWood(), Hanging: true, Age: r.IntN(5)}, nil)
	}
}

// placeMangroveMuddyRoots pre-places muddy mangrove roots in nearby mud blocks.
func (s Sapling) placeMangroveMuddyRoots(pos cube.Pos, tx *world.Tx, layout *saplingTreeLayout) {
	for _, dir := range cube.Directions() {
		ground := pos.Add(offset(dir, 1)).Side(cube.FaceDown)
		if _, ok := tx.Block(ground).(Mud); !ok {
			continue
		}
		layout.set(ground, MuddyMangroveRoots{Axis: axisForDirection(dir)})
	}
}

// treeHeight rolls tree height as base + rand(heightRandA+1) + rand(heightRandB+1).
func treeHeight(r *rand.Rand, baseHeight, heightRandA, heightRandB int) int {
	return baseHeight + r.IntN(heightRandA+1) + r.IntN(heightRandB+1)
}

// axisForDirection returns the horizontal axis matching the direction passed.
func axisForDirection(dir cube.Direction) cube.Axis {
	if dir == cube.North || dir == cube.South {
		return cube.Z
	}
	return cube.X
}

// sampleUniform samples an integer uniformly from the inclusive range passed.
func sampleUniform(r *rand.Rand, minValue, maxValue int) int {
	return minValue + r.IntN(maxValue-minValue+1)
}

// maxFreeTreeHeight returns the tallest height the tree can grow to under the supplied radius rules.
func maxFreeTreeHeight(tx *world.Tx, treeHeight int, pos cube.Pos, sizeAtHeight func(treeHeight, currentHeight int) int, ignoreVines bool, freeBlockChecker func(*world.Tx, cube.Pos) bool) int {
	for dy := 0; dy <= treeHeight+1; dy++ {
		size := sizeAtHeight(treeHeight, dy)
		for dx := -size; dx <= size; dx++ {
			for dz := -size; dz <= size; dz++ {
				check := pos.Add(cube.Pos{dx, dy, dz})
				if !freeBlockChecker(tx, check) || (!ignoreVines && isVine(tx.Block(check))) {
					return dy - 2
				}
			}
		}
	}
	return treeHeight
}

// canGrowOnTreeBase reports whether the block below a trunk position can support tree growth.
func (s Sapling) canGrowOnTreeBase(pos cube.Pos, tx *world.Tx) bool {
	below := tx.Block(pos.Side(cube.FaceDown))
	if s.Wood == MangroveWood() {
		if _, ok := below.(Clay); ok {
			return true
		}
	}
	return supportsVegetation(s, below)
}

// isFreeTreeBlock reports whether a tree may consider a position free during clearance checks.
func isFreeTreeBlock(tx *world.Tx, pos cube.Pos) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	if canReplaceTreeBlock(tx, pos, Air{}) {
		return true
	}
	_, ok := tx.Block(pos).(Log)
	return ok
}

// isAir reports whether the block at the position is air.
func isAir(tx *world.Tx, pos cube.Pos) bool {
	_, ok := tx.Block(pos).(Air)
	return ok
}

// isAirOrLeaves reports whether the block at the position is air or leaves.
func isAirOrLeaves(tx *world.Tx, pos cube.Pos) bool {
	switch tx.Block(pos).(type) {
	case Air, Leaves:
		return true
	default:
		return false
	}
}

// isVine reports whether the block passed is a vine block.
func isVine(b world.Block) bool {
	_, ok := b.(Vines)
	return ok
}

// isNearExcluded reports whether a position lies inside the exclusion radius of an existing propagule.
func isNearExcluded(pos cube.Pos, centers []cube.Pos, radiusXZ, radiusY int) bool {
	for _, center := range centers {
		if abs(center[0]-pos[0]) <= radiusXZ && abs(center[1]-pos[1]) <= radiusY && abs(center[2]-pos[2]) <= radiusXZ {
			return true
		}
	}
	return false
}

// isPodzolReplaceable reports whether a block may be converted to podzol by the mega spruce decorator.
func isPodzolReplaceable(b world.Block) bool {
	switch b.(type) {
	case Dirt, Grass:
		return true
	default:
		return false
	}
}

// placeBelowOverworldTrunk applies the default below-trunk provider for overworld trees.
func (s Sapling) placeBelowOverworldTrunk(pos cube.Pos, tx *world.Tx) {
	switch tx.Block(pos).(type) {
	case Grass, Farmland:
		tx.SetBlock(pos, Dirt{}, nil)
	}
}

// foliageAttachment stores a foliage anchor together with its radius and trunk metadata.
type foliageAttachment struct {
	// pos is the anchor position of the foliage attachment.
	pos cube.Pos
	// radiusOffset shifts the foliage radius for the attached canopy.
	radiusOffset int
	// doubleTrunk marks foliage attached to a 2x2 trunk.
	doubleTrunk bool
}

// fancyFoliageCoord stores the attachment position and branch base of a fancy oak canopy.
type fancyFoliageCoord struct {
	// attachment is the canopy anchor position.
	attachment cube.Pos
	// branchBase is the Y position where the branch leaves the trunk.
	branchBase int
}

// rowCoord stores the signed and absolute coordinates of a leaf row cell.
type rowCoord struct {
	// signedX is the signed X offset from the row center.
	signedX int
	// signedZ is the signed Z offset from the row center.
	signedZ int
	// localX is the absolute X distance used by skip rules.
	localX int
	// localZ is the absolute Z distance used by skip rules.
	localZ int
}

// saplingTreeLayout stores a staged set of block placements for tree generation.
type saplingTreeLayout struct {
	// tx is the world transaction the tree is generated in.
	tx *world.Tx
	// blocks holds the arranged block placements indexed by position.
	blocks map[cube.Pos]world.Block
}

// newSaplingTreeLayout creates a new tree placement layout for the current transaction.
func newSaplingTreeLayout(tx *world.Tx) *saplingTreeLayout {
	return &saplingTreeLayout{tx: tx, blocks: map[cube.Pos]world.Block{}}
}

// set records a block placement, keeping trunk blocks when leaves overlap them.
func (p *saplingTreeLayout) set(pos cube.Pos, b world.Block) {
	if current, ok := p.blocks[pos]; ok {
		if _, isLog := current.(Log); isLog {
			if _, isLeaf := b.(Leaves); isLeaf {
				return
			}
		}
	}
	p.blocks[pos] = b
}

// setIfValid records a block placement only if the world currently allows it to be replaced.
func (p *saplingTreeLayout) setIfValid(pos cube.Pos, b world.Block) bool {
	if pos.OutOfBounds(p.tx.Range()) {
		return false
	}
	if _, ok := p.blocks[pos]; ok || canReplaceTreeBlock(p.tx, pos, b) {
		p.set(pos, b)
		return true
	}
	return false
}

// verticalTrunk adds a one-block-wide vertical trunk to the layout.
func (p *saplingTreeLayout) verticalTrunk(pos cube.Pos, height int, wood WoodType) {
	for y := 0; y < height; y++ {
		p.set(pos.Add(cube.Pos{0, y, 0}), Log{Wood: wood, Axis: cube.Y})
	}
}

// giantTrunk adds the giant trunk shape, where the topmost log layer only keeps the north-west column.
func (p *saplingTreeLayout) giantTrunk(origin cube.Pos, height int, wood WoodType) {
	for y := 0; y < height; y++ {
		p.set(origin.Add(cube.Pos{0, y, 0}), Log{Wood: wood, Axis: cube.Y})
		if y < height-1 {
			p.set(origin.Add(cube.Pos{1, y, 0}), Log{Wood: wood, Axis: cube.Y})
			p.set(origin.Add(cube.Pos{1, y, 1}), Log{Wood: wood, Axis: cube.Y})
			p.set(origin.Add(cube.Pos{0, y, 1}), Log{Wood: wood, Axis: cube.Y})
		}
	}
}

// branch adds a simple branch that steps sideways and rises near the end, returning the topmost branch position.
func (p *saplingTreeLayout) branch(start cube.Pos, dir cube.Direction, length, rise int, wood WoodType) cube.Pos {
	tip := start
	for i := 1; i <= length; i++ {
		tip = start.Add(offset(dir, i))
		p.setIfValid(tip, Log{Wood: wood, Axis: cube.Y})
	}
	for y := 1; y <= rise; y++ {
		tip = tip.Add(cube.Pos{0, 1, 0})
		p.setIfValid(tip, Log{Wood: wood, Axis: cube.Y})
	}
	return tip
}

// leafLayer adds a leaf disk around the provided center position.
func (p *saplingTreeLayout) leafLayer(center cube.Pos, radius int, wood WoodType, trimCorners bool) {
	leaf := Leaves{Wood: wood}
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if trimCorners && radius > 0 && abs(x) == radius && abs(z) == radius {
				continue
			}
			p.setIfValid(center.Add(cube.Pos{x, 0, z}), leaf)
		}
	}
}

// limb either checks or places the branch path between two positions.
func (p *saplingTreeLayout) limb(from, to cube.Pos, wood WoodType, place bool) bool {
	if !place && from == to {
		return true
	}

	deltaX, deltaY, deltaZ := to[0]-from[0], to[1]-from[1], to[2]-from[2]
	steps := max(abs(deltaX), max(abs(deltaY), abs(deltaZ)))
	stepX := float64(deltaX) / float64(steps)
	stepY := float64(deltaY) / float64(steps)
	stepZ := float64(deltaZ) / float64(steps)

	for i := 0; i <= steps; i++ {
		current := cube.Pos{
			from[0] + int(math.Floor(0.5+float64(i)*stepX)),
			from[1] + int(math.Floor(0.5+float64(i)*stepY)),
			from[2] + int(math.Floor(0.5+float64(i)*stepZ)),
		}
		if place {
			p.setIfValid(current, Log{Wood: wood, Axis: logAxisForLimb(from, current)})
			continue
		}
		if !isFreeTreeBlock(p.tx, current) {
			return false
		}
	}
	return true
}

// placeLeavesRow places one foliage row using the provided skip rules.
func (p *saplingTreeLayout) placeLeavesRow(center cube.Pos, radius, localY int, large bool, wood WoodType, skip func(rowCoord, int, int, bool, *rand.Rand) bool, r *rand.Rand) {
	extra := 0
	if large {
		extra = 1
	}
	leaf := Leaves{Wood: wood}
	for signedX := -radius; signedX <= radius+extra; signedX++ {
		for signedZ := -radius; signedZ <= radius+extra; signedZ++ {
			localX, localZ := abs(signedX), abs(signedZ)
			if large {
				localX = min(abs(signedX), abs(signedX-1))
				localZ = min(abs(signedZ), abs(signedZ-1))
			}
			coord := rowCoord{signedX: signedX, signedZ: signedZ, localX: localX, localZ: localZ}
			if skip(coord, localY, radius, large, r) {
				continue
			}
			p.setIfValid(center.Add(cube.Pos{signedX, localY, signedZ}), leaf)
		}
	}
}

// blobFoliage places the rounded canopy used by straight blob trees.
func (p *saplingTreeLayout) blobFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wood WoodType, r *rand.Rand) {
	for localY := offset; localY >= offset-foliageHeight; localY-- {
		rangeValue := max(foliageRadius+attachment.radiusOffset-1-localY/2, 0)
		p.placeLeavesRow(attachment.pos, rangeValue, localY, attachment.doubleTrunk, wood, func(coord rowCoord, rowY, rowRange int, _ bool, randSource *rand.Rand) bool {
			return coord.localX == rowRange && coord.localZ == rowRange && (randSource.IntN(2) == 0 || rowY == 0)
		}, r)
	}
}

// blobCanopy keeps the older four-row blob canopy helper used by the mangrove fallback path.
func (p *saplingTreeLayout) blobCanopy(top cube.Pos, radius int, wood WoodType) {
	for i := 0; i < 4; i++ {
		layerRadius := max(radius-1, 0)
		if i >= 2 {
			layerRadius = radius
		}
		p.leafLayer(top.Add(cube.Pos{0, -i, 0}), layerRadius, wood, true)
	}
}

// fancyFoliage places the rounded fancy oak canopy layers.
func (p *saplingTreeLayout) fancyFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wood WoodType) {
	for localY := offset; localY >= offset-foliageHeight; localY-- {
		rangeValue := foliageRadius
		if localY != offset && localY != offset-foliageHeight {
			rangeValue++
		}
		p.placeLeavesRow(attachment.pos, rangeValue, localY, attachment.doubleTrunk, wood, func(coord rowCoord, _, rowRange int, _ bool, _ *rand.Rand) bool {
			dx := float64(coord.localX) + 0.5
			dz := float64(coord.localZ) + 0.5
			return dx*dx+dz*dz > float64(rowRange*rowRange)
		}, nil)
	}
}

// acaciaFoliage places foliage using the acacia row layout.
func (p *saplingTreeLayout) acaciaFoliage(attachment foliageAttachment, foliageRadius, offset int, wood WoodType) {
	base := attachment.pos.Add(cube.Pos{0, offset, 0})
	p.placeLeavesRow(base, foliageRadius+attachment.radiusOffset, -1, attachment.doubleTrunk, wood, func(coord rowCoord, _, rowRange int, _ bool, _ *rand.Rand) bool {
		return coord.localX == rowRange && coord.localZ == rowRange && rowRange > 0
	}, nil)
	p.placeLeavesRow(base, foliageRadius-1, 0, attachment.doubleTrunk, wood, func(coord rowCoord, _, _ int, _ bool, _ *rand.Rand) bool {
		return (coord.localX > 1 || coord.localZ > 1) && coord.localX != 0 && coord.localZ != 0
	}, nil)
	p.placeLeavesRow(base, foliageRadius+attachment.radiusOffset-1, 0, attachment.doubleTrunk, wood, func(coord rowCoord, _, _ int, _ bool, _ *rand.Rand) bool {
		return (coord.localX > 1 || coord.localZ > 1) && coord.localX != 0 && coord.localZ != 0
	}, nil)
}

// spruceFoliage places foliage using the spruce foliage state machine.
func (p *saplingTreeLayout) spruceFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wood WoodType, r *rand.Rand) {
	currentRadius := r.IntN(2)
	maxRadius := 1
	minRadius := 0
	for localY := offset; localY >= -foliageHeight; localY-- {
		p.placeLeavesRow(attachment.pos, currentRadius, localY, attachment.doubleTrunk, wood, func(coord rowCoord, _, rowRange int, _ bool, _ *rand.Rand) bool {
			return coord.localX == rowRange && coord.localZ == rowRange && rowRange > 0
		}, nil)
		if currentRadius >= maxRadius {
			currentRadius = minRadius
			minRadius = 1
			maxRadius = min(maxRadius+1, foliageRadius+attachment.radiusOffset)
		} else {
			currentRadius++
		}
	}
}

// megaPineFoliage places foliage for both mega spruce variants.
func (p *saplingTreeLayout) megaPineFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wood WoodType) {
	prevRadius := 0
	for y := attachment.pos[1] - foliageHeight + offset; y <= attachment.pos[1]+offset; y++ {
		heightFromTop := attachment.pos[1] - y
		smoothRadius := foliageRadius + attachment.radiusOffset + int(math.Floor(float64(heightFromTop)/float64(foliageHeight)*3.5))
		jaggedRadius := smoothRadius
		if heightFromTop > 0 && smoothRadius == prevRadius && (y&1) == 0 {
			jaggedRadius = smoothRadius + 1
		}
		p.placeLeavesRow(cube.Pos{attachment.pos[0], y, attachment.pos[2]}, jaggedRadius, 0, attachment.doubleTrunk, wood, func(coord rowCoord, _, rowRange int, _ bool, _ *rand.Rand) bool {
			return coord.localX+coord.localZ >= 7 || coord.localX*coord.localX+coord.localZ*coord.localZ > rowRange*rowRange
		}, nil)
		prevRadius = smoothRadius
	}
}

// darkOakFoliage places the layered canopy used by dark oak-style trees.
func (p *saplingTreeLayout) darkOakFoliage(attachment foliageAttachment, foliageRadius, offset int, wood WoodType, r *rand.Rand) {
	base := attachment.pos.Add(cube.Pos{0, offset, 0})
	if attachment.doubleTrunk {
		p.placeDarkOakRow(base, foliageRadius+2, -1, true, wood, r)
		p.placeDarkOakRow(base, foliageRadius+3, 0, true, wood, r)
		p.placeDarkOakRow(base, foliageRadius+2, 1, true, wood, r)
		if r.IntN(2) == 0 {
			p.placeDarkOakRow(base, foliageRadius, 2, true, wood, r)
		}
		return
	}
	p.placeDarkOakRow(base, foliageRadius+2, -1, false, wood, r)
	p.placeDarkOakRow(base, foliageRadius+1, 0, false, wood, r)
}

// placeDarkOakRow places one row of dark oak foliage.
func (p *saplingTreeLayout) placeDarkOakRow(center cube.Pos, radius, localY int, large bool, wood WoodType, r *rand.Rand) {
	p.placeLeavesRow(center, radius, localY, large, wood, func(coord rowCoord, rowY, rowRange int, isLarge bool, _ *rand.Rand) bool {
		if rowY == 0 && isLarge && (coord.signedX == -rowRange || coord.signedX >= rowRange) && (coord.signedZ == -rowRange || coord.signedZ >= rowRange) {
			return true
		}
		if rowY == -1 && !isLarge {
			return coord.localX == rowRange && coord.localZ == rowRange
		}
		return rowY == 1 && coord.localX+coord.localZ > rowRange*2-2
	}, r)
}

// megaJungleFoliage places the canopy rows used by mega jungle branches and trunk tops.
func (p *saplingTreeLayout) megaJungleFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wood WoodType, r *rand.Rand) {
	foliageLayers := 1 + r.IntN(2)
	if attachment.doubleTrunk {
		foliageLayers = foliageHeight
	}
	for localY := offset; localY >= offset-foliageLayers; localY-- {
		rangeValue := foliageRadius + attachment.radiusOffset + 1 - localY
		p.placeLeavesRow(attachment.pos, rangeValue, localY, attachment.doubleTrunk, wood, func(coord rowCoord, _, rowRange int, _ bool, _ *rand.Rand) bool {
			return coord.localX+coord.localZ >= 7 || coord.localX*coord.localX+coord.localZ*coord.localZ > rowRange*rowRange
		}, nil)
	}
}

// cherryBranch places one cherry side branch and returns its foliage attachment.
func (p *saplingTreeLayout) cherryBranch(origin cube.Pos, treeHeight int, direction cube.Direction, branchStart int, doubleBranch bool, wood WoodType, r *rand.Rand) foliageAttachment {
	currentX, currentY, currentZ := origin[0], origin[1]+branchStart, origin[2]
	branchEndY := treeHeight - 1 + sampleUniform(r, -1, 0)
	extended := doubleBranch || branchEndY < branchStart
	horizontalLength := sampleUniform(r, 2, 4)
	if extended {
		horizontalLength++
	}
	targetX := origin[0] + offset(direction, horizontalLength)[0]
	targetY := origin[1] + branchEndY
	targetZ := origin[2] + offset(direction, horizontalLength)[2]
	firstSteps := 1
	if extended {
		firstSteps = 2
	}
	horizontalAxis := axisForDirection(direction)

	for i := 0; i < firstSteps; i++ {
		step := offset(direction, 1)
		currentX += step[0]
		currentZ += step[2]
		p.setIfValid(cube.Pos{currentX, currentY, currentZ}, Log{Wood: wood, Axis: horizontalAxis})
	}

	verticalStep := -1
	if targetY > currentY {
		verticalStep = 1
	}
	for {
		distance := abs(targetX-currentX) + abs(targetY-currentY) + abs(targetZ-currentZ)
		if distance == 0 {
			return foliageAttachment{pos: cube.Pos{targetX, targetY + 1, targetZ}}
		}
		moveVertical := r.Float64() < float64(abs(targetY-currentY))/float64(distance)
		if moveVertical {
			currentY += verticalStep
			p.setIfValid(cube.Pos{currentX, currentY, currentZ}, Log{Wood: wood, Axis: cube.Y})
			continue
		}
		step := offset(direction, 1)
		currentX += step[0]
		currentZ += step[2]
		p.setIfValid(cube.Pos{currentX, currentY, currentZ}, Log{Wood: wood, Axis: horizontalAxis})
	}
}

// cherryFoliage places the layered cherry canopy and its hanging leaf extensions.
func (p *saplingTreeLayout) cherryFoliage(attachment foliageAttachment, foliageHeight, foliageRadius, offset int, wideBottomLayerHoleChance, cornerHoleChance, hangingLeavesChance, hangingLeavesExtensionChance float64, wood WoodType, r *rand.Rand) {
	base := attachment.pos.Add(cube.Pos{0, offset, 0})
	rangeValue := foliageRadius + attachment.radiusOffset - 1
	skipper := func(coord rowCoord, localY, rowRange int, _ bool, randSource *rand.Rand) bool {
		if localY == -1 && (coord.localX == rowRange || coord.localZ == rowRange) && randSource.Float64() < wideBottomLayerHoleChance {
			return true
		}
		isCorner := coord.localX == rowRange && coord.localZ == rowRange
		if rowRange > 2 {
			return isCorner || (coord.localX+coord.localZ > rowRange*2-2 && randSource.Float64() < cornerHoleChance)
		}
		return isCorner && randSource.Float64() < cornerHoleChance
	}

	p.placeLeavesRow(base, rangeValue-2, foliageHeight-3, attachment.doubleTrunk, wood, skipper, r)
	p.placeLeavesRow(base, rangeValue-1, foliageHeight-4, attachment.doubleTrunk, wood, skipper, r)
	for localY := foliageHeight - 5; localY >= 0; localY-- {
		p.placeLeavesRow(base, rangeValue, localY, attachment.doubleTrunk, wood, skipper, r)
	}
	p.placeLeavesRowWithExtensions(base, rangeValue, -1, attachment.doubleTrunk, wood, skipper, hangingLeavesChance, hangingLeavesExtensionChance, r)
	p.placeLeavesRowWithExtensions(base, rangeValue-1, -2, attachment.doubleTrunk, wood, skipper, hangingLeavesChance, hangingLeavesExtensionChance, r)
}

// placeLeavesRowWithExtensions places a cherry leaf row and tries to extend leaves below its edges.
func (p *saplingTreeLayout) placeLeavesRowWithExtensions(center cube.Pos, radius, localY int, large bool, wood WoodType, skip func(rowCoord, int, int, bool, *rand.Rand) bool, hangingLeavesChance, hangingLeavesExtensionChance float64, r *rand.Rand) {
	p.placeLeavesRow(center, radius, localY, large, wood, skip, r)

	extra := 0
	if large {
		extra = 1
	}
	leafY := center[1] + localY
	originBelowY := center[1] - 1
	for _, direction := range cube.Directions() {
		cw := direction.RotateRight()
		edge := radius
		if cw == cube.South || cw == cube.East {
			edge = radius + extra
		}
		cursorX := center[0] + offset(cw, edge)[0] + offset(direction, -radius)[0]
		cursorZ := center[2] + offset(cw, edge)[2] + offset(direction, -radius)[2]
		for i := -radius; i < radius+extra; i++ {
			leafPos := cube.Pos{cursorX, leafY, cursorZ}
			if placed, ok := p.blocks[leafPos].(Leaves); ok && placed.Wood == wood {
				if p.tryLeafExtension(cube.Pos{cursorX, leafY - 1, cursorZ}, cube.Pos{center[0], originBelowY, center[2]}, hangingLeavesChance, wood, r) {
					p.tryLeafExtension(cube.Pos{cursorX, leafY - 2, cursorZ}, cube.Pos{center[0], originBelowY, center[2]}, hangingLeavesExtensionChance, wood, r)
				}
			}
			cursorX += offset(direction, 1)[0]
			cursorZ += offset(direction, 1)[2]
		}
	}
}

// tryLeafExtension tries to place one hanging cherry leaf extension.
func (p *saplingTreeLayout) tryLeafExtension(pos, origin cube.Pos, chance float64, wood WoodType, r *rand.Rand) bool {
	if abs(pos[0]-origin[0])+abs(pos[1]-origin[1])+abs(pos[2]-origin[2]) >= 7 || r.Float64() > chance {
		return false
	}
	return p.setIfValid(pos, Leaves{Wood: wood})
}

// logAxisForLimb returns the dominant axis of a limb segment.
func logAxisForLimb(from, to cube.Pos) cube.Axis {
	xDiff, zDiff := abs(to[0]-from[0]), abs(to[2]-from[2])
	maxDiff := max(xDiff, zDiff)
	if maxDiff == 0 {
		return cube.Y
	}
	if xDiff == maxDiff {
		return cube.X
	}
	return cube.Z
}

// trimFancyBranch reports whether a fancy oak branch should be kept at the current height.
func trimFancyBranch(maxHeight, currentHeight int) bool {
	return float64(currentHeight) >= float64(maxHeight)*0.2
}

// fancyTreeShape returns the branch radius factor at a given height of a fancy oak.
func fancyTreeShape(height, currentY int) float64 {
	if float64(currentY) < float64(height)*0.3 {
		return -1
	}
	midpoint := float64(height) / 2
	heightFromMid := midpoint - float64(currentY)
	if math.Abs(heightFromMid) >= midpoint {
		if heightFromMid == 0 {
			return midpoint * 0.5
		}
		return 0
	}
	radius := math.Sqrt(midpoint*midpoint - heightFromMid*heightFromMid)
	if heightFromMid == 0 {
		radius = midpoint
	}
	return radius * 0.5
}

// apply validates and commits all arranged block placements.
func (p *saplingTreeLayout) apply() bool {
	for pos, block := range p.blocks {
		if pos.OutOfBounds(p.tx.Range()) || !canReplaceTreeBlock(p.tx, pos, block) {
			return false
		}
	}
	for pos, block := range p.blocks {
		if _, ok := block.(Leaves); ok {
			continue
		}
		p.tx.SetBlock(pos, block, nil)
	}
	for pos, block := range p.blocks {
		if _, ok := block.(Leaves); !ok {
			continue
		}
		p.tx.SetBlock(pos, block, nil)
	}
	return true
}

// placeMegaConiferPodzol runs the ground decorator around the base of a mega spruce-family tree.
func (s Sapling) placeMegaConiferPodzol(origin cube.Pos, tx *world.Tx, r *rand.Rand) {
	bases := []cube.Pos{
		origin,
		origin.Add(cube.Pos{1, 0, 0}),
		origin.Add(cube.Pos{0, 0, 1}),
		origin.Add(cube.Pos{1, 0, 1}),
	}
	for _, base := range bases {
		s.placePodzolCircle(base.Add(cube.Pos{-1, 0, -1}), tx)
		s.placePodzolCircle(base.Add(cube.Pos{2, 0, -1}), tx)
		s.placePodzolCircle(base.Add(cube.Pos{-1, 0, 2}), tx)
		s.placePodzolCircle(base.Add(cube.Pos{2, 0, 2}), tx)
		for i := 0; i < 5; i++ {
			placement := r.IntN(64)
			x := placement % 8
			z := placement / 8
			if x == 0 || x == 7 || z == 0 || z == 7 {
				s.placePodzolCircle(base.Add(cube.Pos{-3 + x, 0, -3 + z}), tx)
			}
		}
	}
}

// placePodzolCircle places the 5x5 rounded podzol circle used by the ground decorator.
func (s Sapling) placePodzolCircle(pos cube.Pos, tx *world.Tx) {
	for x := -2; x <= 2; x++ {
		for z := -2; z <= 2; z++ {
			if abs(x) == 2 && abs(z) == 2 {
				continue
			}
			s.placePodzolAt(pos.Add(cube.Pos{x, 0, z}), tx)
		}
	}
}

// placePodzolAt searches downward using the decorator scan rules and converts the first replaceable ground to podzol.
func (s Sapling) placePodzolAt(pos cube.Pos, tx *world.Tx) {
	for dy := 2; dy >= -3; dy-- {
		cursor := pos.Add(cube.Pos{0, dy, 0})
		if isPodzolReplaceable(tx.Block(cursor)) {
			tx.SetBlock(cursor, Podzol{}, nil)
			return
		}
		if _, ok := tx.Block(cursor).(Air); !ok && dy < 0 {
			return
		}
	}
}

// canReplaceTreeBlock reports if an existing block may be replaced during tree growth.
func canReplaceTreeBlock(tx *world.Tx, pos cube.Pos, placing world.Block) bool {
	existing := tx.Block(pos)
	switch existing.(type) {
	case Air, Sapling, Leaves, Flower, DoubleFlower, ShortGrass, Fern, DoubleTallGrass, PinkPetals, DeadBush, NetherSprouts:
		return true
	case Crop:
		return true
	}
	if _, ok := existing.(Replaceable); ok {
		return true
	}
	if _, isLeaves := placing.(Leaves); isLeaves {
		if _, ok := tx.Liquid(pos); ok {
			return true
		}
	}
	return false
}

// offset converts a horizontal direction and length to a block offset.
func offset(dir cube.Direction, amount int) cube.Pos {
	switch dir {
	case cube.North:
		return cube.Pos{0, 0, -amount}
	case cube.South:
		return cube.Pos{0, 0, amount}
	case cube.West:
		return cube.Pos{-amount, 0, 0}
	default:
		return cube.Pos{amount, 0, 0}
	}
}
