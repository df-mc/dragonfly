package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// growTree selects the Java-inspired local tree generator that matches the sapling wood type.
func (s Sapling) growTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	switch s.Wood {
	case OakWood():
		if r.Float64() < 0.1 {
			return s.growFancyOak(pos, tx, r)
		}
		return s.growStraightBlob(pos, tx, r, 4, 2, 0, 2)
	case SpruceWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			if r.Float64() < 0.5 {
				return s.growMegaPine(origin, tx, r)
			}
			return s.growMegaSpruce(origin, tx, r)
		}
		return s.growSpruce(pos, tx, r)
	case BirchWood():
		return s.growStraightBlob(pos, tx, r, 5, 2, 0, 2)
	case JungleWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			return s.growMegaJungle(origin, tx, r)
		}
		return s.growStraightBlob(pos, tx, r, 4, 8, 0, 2)
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
func (s Sapling) growStraightBlob(pos cube.Pos, tx *world.Tx, r *rand.Rand, baseHeight, heightRandA, heightRandB, leafRadius int) bool {
	height := treeHeight(r, baseHeight, heightRandA, heightRandB)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	layout.blobCanopy(pos.Add(cube.Pos{0, height, 0}), leafRadius, s.Wood)
	return layout.apply()
}

// growFancyOak places a taller oak with multiple foliage attachments to better match Java's fancy oak grower.
func (s Sapling) growFancyOak(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 3, 11, 0)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)

	trunkTop := pos.Add(cube.Pos{0, height, 0})
	layout.blobCanopy(trunkTop, 2, s.Wood)

	branchCount := 1 + r.IntN(3)
	for i := 0; i < branchCount; i++ {
		dir := cube.Directions()[r.IntN(len(cube.Directions()))]
		startY := max(2, height-4-r.IntN(3))
		branchStart := pos.Add(cube.Pos{0, startY, 0})
		branchLength := 1 + r.IntN(3)
		branchRise := 1 + r.IntN(2)
		branchTop := layout.branch(branchStart, dir, branchLength, branchRise, s.Wood)
		layout.blobCanopy(branchTop, 2, s.Wood)
	}
	return layout.apply()
}

// growSpruce places the narrow conical spruce grown from a single spruce sapling.
func (s Sapling) growSpruce(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 5, 2, 1)
	crownBase := max(1, height-4-r.IntN(2))
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	for y := crownBase; y <= height; y++ {
		radius := min(2, 1+(height-y+1)/2)
		if y == height {
			radius = 0
		}
		layout.leafLayer(pos.Add(cube.Pos{0, y, 0}), radius, s.Wood, true)
	}
	layout.leafLayer(pos.Add(cube.Pos{0, height + 1, 0}), 0, s.Wood, false)
	return layout.apply()
}

// growAcacia places a leaning acacia trunk with an optional second fork based on Java's forking trunk placer.
func (s Sapling) growAcacia(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 5, 2, 2)
	leanDirection := cube.Directions()[r.IntN(len(cube.Directions()))]
	leanHeight := height - r.IntN(4) - 1
	leanSteps := 3 - r.IntN(3)
	layout := newSaplingTreeLayout(tx)

	current := pos
	for y := 0; y < height; y++ {
		if y >= leanHeight && leanSteps > 0 {
			current = current.Add(offset(leanDirection, 1))
			leanSteps--
		}
		layout.set(cube.Pos{current[0], pos[1] + y, current[2]}, Log{Wood: s.Wood, Axis: cube.Y})
	}

	mainTop := cube.Pos{current[0], pos[1] + height, current[2]}
	layout.flatCanopy(mainTop, 2, s.Wood)

	branchDirection := cube.Directions()[r.IntN(len(cube.Directions()))]
	if branchDirection != leanDirection {
		branchY := leanHeight - r.IntN(2) - 1
		branchSteps := 1 + r.IntN(3)
		branchPos := pos
		var branchTop cube.Pos
		for y := branchY; y < height && branchSteps > 0; y++ {
			if y < 1 {
				continue
			}
			branchPos = branchPos.Add(offset(branchDirection, 1))
			branchTop = cube.Pos{branchPos[0], pos[1] + y, branchPos[2]}
			layout.set(branchTop, Log{Wood: s.Wood, Axis: cube.Y})
			branchSteps--
		}
		if branchTop != (cube.Pos{}) {
			layout.flatCanopy(branchTop.Add(cube.Pos{0, 1, 0}), 1, s.Wood)
		}
	}
	return layout.apply()
}

// growDarkOak places the leaning 2x2 trunk and side branch stubs used by dark oak and pale oak saplings.
func (s Sapling) growDarkOak(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 6, 2, 1)
	leanDirection := cube.Directions()[r.IntN(len(cube.Directions()))]
	leanHeight := height - r.IntN(4)
	leanSteps := 2 - r.IntN(3)
	layout := newSaplingTreeLayout(tx)

	current := origin
	for y := 0; y < height; y++ {
		if y >= leanHeight && leanSteps > 0 {
			current = current.Add(offset(leanDirection, 1))
			leanSteps--
		}
		layout.twoByTwoLayer(cube.Pos{current[0], origin[1] + y, current[2]}, s.Wood)
	}

	top := cube.Pos{current[0], origin[1] + height - 1, current[2]}
	layout.darkOakCanopy(top, s.Wood, r)

	for x := -1; x <= 2; x++ {
		for z := -1; z <= 2; z++ {
			if x >= 0 && x <= 1 && z >= 0 && z <= 1 {
				continue
			}
			if r.IntN(3) != 0 {
				continue
			}
			length := 2 + r.IntN(3)
			for branchY := 0; branchY < length; branchY++ {
				layout.set(origin.Add(cube.Pos{x, height - branchY - 2, z}), Log{Wood: s.Wood, Axis: cube.Y})
			}
			layout.leafLayer(origin.Add(cube.Pos{x, height, z}), 1, s.Wood, true)
		}
	}
	return layout.apply()
}

// growCherry places a cherry tree with two raised canopies to better match the multi-branch Java cherry feature.
func (s Sapling) growCherry(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 7, 1, 0)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)

	left := cube.Directions()[r.IntN(len(cube.Directions()))]
	right := left.RotateRight()
	if r.IntN(2) == 0 {
		right = left.RotateLeft()
	}

	leftTop := layout.branch(pos.Add(cube.Pos{0, height - 2, 0}), left, 2+r.IntN(2), 2, s.Wood)
	rightTop := layout.branch(pos.Add(cube.Pos{0, height - 3, 0}), right, 2+r.IntN(2), 2, s.Wood)
	layout.cherryCanopy(pos.Add(cube.Pos{0, height, 0}), s.Wood)
	layout.cherryCanopy(leftTop, s.Wood)
	layout.cherryCanopy(rightTop, s.Wood)
	return layout.apply()
}

// growMegaSpruce places the broader-crowned mega spruce selected by half of 2x2 spruce growth attempts.
func (s Sapling) growMegaSpruce(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMegaConifer(origin, tx, r, 13, 2, 14, 13, 17, 3)
}

// growMegaPine places the narrower mega pine selected by the other half of 2x2 spruce growth attempts.
func (s Sapling) growMegaPine(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	return s.growMegaConifer(origin, tx, r, 13, 2, 14, 3, 7, 2)
}

// growMegaConifer places a 2x2 giant spruce-family tree and converts nearby ground to podzol.
func (s Sapling) growMegaConifer(origin cube.Pos, tx *world.Tx, r *rand.Rand, baseHeight, heightRandA, heightRandB, crownMin, crownMax, maxRadius int) bool {
	height := treeHeight(r, baseHeight, heightRandA, heightRandB)
	crownHeight := crownMin + r.IntN(crownMax-crownMin+1)
	layout := newSaplingTreeLayout(tx)
	layout.twoByTwoTrunk(origin, height, s.Wood)

	topY := origin[1] + height
	bottomY := max(origin[1]+height-crownHeight, origin[1]+1)
	for y := topY; y >= bottomY; y-- {
		progress := topY - y
		radius := min(maxRadius, progress/2)
		if y >= topY-1 {
			radius = 0
		} else if radius == 0 {
			radius = 1
		}
		layout.leafSquare(cube.Pos{origin[0], y, origin[2]}, radius, s.Wood, true)
	}

	for x := -2; x <= 3; x++ {
		for z := -2; z <= 3; z++ {
			if abs(x) == 3 && abs(z) == 3 {
				continue
			}
			ground := origin.Add(cube.Pos{x, -1, z})
			switch tx.Block(ground).(type) {
			case Dirt, Grass:
				layout.set(ground, Podzol{})
			}
		}
	}
	return layout.apply()
}

// growMegaJungle places a giant jungle tree with side canopies and added trunk vines.
func (s Sapling) growMegaJungle(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := treeHeight(r, 10, 2, 19)
	layout := newSaplingTreeLayout(tx)
	layout.twoByTwoTrunk(origin, height, s.Wood)

	for i := 0; i < 5; i++ {
		radius := max(1, 3-i/2)
		layout.leafSquare(origin.Add(cube.Pos{0, height - i, 0}), radius, s.Wood, true)
	}

	for _, dir := range cube.Directions() {
		branchY := height - 3 - r.IntN(4)
		branchLength := 2 + r.IntN(3)
		branchTop := layout.branch(origin.Add(cube.Pos{0, branchY, 0}), dir, branchLength, 1, s.Wood)
		layout.blobCanopy(branchTop, 2, s.Wood)
	}

	if !layout.apply() {
		return false
	}
	for y := 1; y < height-1; y++ {
		s.placeTrunkVines(origin.Add(cube.Pos{0, y, 0}), tx, r)
		s.placeTrunkVines(origin.Add(cube.Pos{1, y, 0}), tx, r)
		s.placeTrunkVines(origin.Add(cube.Pos{0, y, 1}), tx, r)
		s.placeTrunkVines(origin.Add(cube.Pos{1, y, 1}), tx, r)
	}
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
			if s.isTwoByTwo(pos.Add(cube.Pos{dx, 0, dz}), tx) {
				return pos.Add(cube.Pos{dx, 0, dz}), true
			}
		}
	}
	return cube.Pos{}, false
}

// isTwoByTwo checks if a 2x2 square from the origin contains matching saplings.
func (s Sapling) isTwoByTwo(origin cube.Pos, tx *world.Tx) bool {
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

// placeTrunkVines attempts to attach jungle vines to the air blocks around a trunk position.
func (s Sapling) placeTrunkVines(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	for _, dir := range cube.Directions() {
		if r.Float64() >= 0.25 {
			continue
		}
		vinePos := pos.Add(offset(dir, 1))
		if _, ok := tx.Block(vinePos).(Air); !ok {
			continue
		}
		tx.SetBlock(vinePos, (Vines{}).WithAttachment(dir.Opposite(), true), nil)
	}
}

// placeHangingPropagules attaches hanging propagules below generated mangrove leaves when space is available.
func (s Sapling) placeHangingPropagules(tx *world.Tx, blocks map[cube.Pos]world.Block, r *rand.Rand, chance float64) {
	for pos, block := range blocks {
		leaves, ok := block.(Leaves)
		if !ok || leaves.Wood != MangroveWood() {
			continue
		}
		under := pos.Side(cube.FaceDown)
		if _, ok := tx.Block(under).(Air); !ok || r.Float64() >= chance {
			continue
		}
		tx.SetBlock(under, Sapling{Wood: MangroveWood(), Hanging: true, Age: r.IntN(5)}, nil)
	}
}

// placeMangroveMuddyRoots pre-places muddy mangrove roots in nearby mud blocks to match the Java feature decor.
func (s Sapling) placeMangroveMuddyRoots(pos cube.Pos, tx *world.Tx, layout *saplingTreeLayout) {
	for _, dir := range cube.Directions() {
		ground := pos.Add(offset(dir, 1)).Side(cube.FaceDown)
		if _, ok := tx.Block(ground).(Mud); !ok {
			continue
		}
		layout.set(ground, MuddyMangroveRoots{Axis: axisForDirection(dir)})
	}
}

// treeHeight mirrors Java's trunk placer height roll of base + rand(heightRandA+1) + rand(heightRandB+1).
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

// verticalTrunk adds a one-block-wide vertical trunk to the layout.
func (p *saplingTreeLayout) verticalTrunk(pos cube.Pos, height int, wood WoodType) {
	for y := 0; y < height; y++ {
		p.set(pos.Add(cube.Pos{0, y, 0}), Log{Wood: wood, Axis: cube.Y})
	}
}

// twoByTwoTrunk adds a 2x2 vertical trunk to the layout.
func (p *saplingTreeLayout) twoByTwoTrunk(origin cube.Pos, height int, wood WoodType) {
	for y := 0; y < height; y++ {
		p.twoByTwoLayer(origin.Add(cube.Pos{0, y, 0}), wood)
	}
}

// twoByTwoLayer adds a single 2x2 trunk layer at the y-level of the origin passed.
func (p *saplingTreeLayout) twoByTwoLayer(origin cube.Pos, wood WoodType) {
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			p.set(origin.Add(cube.Pos{dx, 0, dz}), Log{Wood: wood, Axis: cube.Y})
		}
	}
}

// branch adds a simple branch that steps sideways and rises near the end, returning the topmost branch position.
func (p *saplingTreeLayout) branch(start cube.Pos, dir cube.Direction, length, rise int, wood WoodType) cube.Pos {
	tip := start
	for i := 1; i <= length; i++ {
		tip = start.Add(offset(dir, i))
		p.set(tip, Log{Wood: wood, Axis: cube.Y})
	}
	for y := 1; y <= rise; y++ {
		tip = tip.Add(cube.Pos{0, 1, 0})
		p.set(tip, Log{Wood: wood, Axis: cube.Y})
	}
	return tip
}

// leafLayer adds a leaf disk around the provided center position.
func (p *saplingTreeLayout) leafLayer(center cube.Pos, radius int, wood WoodType, trimCorners bool) {
	leaf := Leaves{Wood: wood, ShouldUpdate: true}
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if trimCorners && radius > 0 && abs(x) == radius && abs(z) == radius {
				continue
			}
			p.set(center.Add(cube.Pos{x, 0, z}), leaf)
		}
	}
}

// leafSquare adds overlapping leaf layers around a 2x2 trunk top.
func (p *saplingTreeLayout) leafSquare(origin cube.Pos, radius int, wood WoodType, trimCorners bool) {
	p.leafLayer(origin, radius, wood, trimCorners)
	p.leafLayer(origin.Add(cube.Pos{1, 0, 0}), radius, wood, trimCorners)
	p.leafLayer(origin.Add(cube.Pos{0, 0, 1}), radius, wood, trimCorners)
	p.leafLayer(origin.Add(cube.Pos{1, 0, 1}), radius, wood, trimCorners)
}

// flatCanopy adds a flat-topped canopy around a center point.
func (p *saplingTreeLayout) flatCanopy(center cube.Pos, radius int, wood WoodType) {
	p.leafLayer(center, radius, wood, true)
	if radius > 0 {
		p.leafLayer(center.Add(cube.Pos{0, 1, 0}), radius-1, wood, true)
	}
}

// blobCanopy adds the four foliage rows used by Java's blob foliage placer.
func (p *saplingTreeLayout) blobCanopy(top cube.Pos, radius int, wood WoodType) {
	for i := 0; i < 4; i++ {
		layerRadius := max(radius-1, 0)
		if i >= 2 {
			layerRadius = radius
		}
		p.leafLayer(top.Add(cube.Pos{0, -i, 0}), layerRadius, wood, true)
	}
}

// darkOakCanopy adds the layered 2x2 canopy used by Java dark oak-style trees.
func (p *saplingTreeLayout) darkOakCanopy(top cube.Pos, wood WoodType, r *rand.Rand) {
	p.leafSquare(top.Add(cube.Pos{0, -1, 0}), 2, wood, true)
	p.leafSquare(top, 3, wood, true)
	p.leafSquare(top.Add(cube.Pos{0, 1, 0}), 2, wood, true)
	if r.IntN(2) == 0 {
		p.leafSquare(top.Add(cube.Pos{0, 2, 0}), 1, wood, false)
	}
}

// cherryCanopy adds the broad rounded canopy used by the local cherry approximation.
func (p *saplingTreeLayout) cherryCanopy(center cube.Pos, wood WoodType) {
	p.leafLayer(center, 3, wood, true)
	p.leafLayer(center.Add(cube.Pos{0, 1, 0}), 2, wood, true)
	p.leafLayer(center.Add(cube.Pos{0, 2, 0}), 1, wood, false)
	p.leafLayer(center.Add(cube.Pos{0, -1, 0}), 2, wood, true)
}

// apply validates and commits all arranged block placements.
func (p *saplingTreeLayout) apply() bool {
	for pos, block := range p.blocks {
		if pos.OutOfBounds(p.tx.Range()) || !canReplaceTreeBlock(p.tx, pos, block) {
			return false
		}
	}
	for pos, block := range p.blocks {
		p.tx.SetBlock(pos, block, nil)
	}
	return true
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
