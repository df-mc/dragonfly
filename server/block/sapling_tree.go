package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// growTree selects the tree generator that matches the sapling wood type.
func (s Sapling) growTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	switch s.Wood {
	case OakWood():
		if r.Float64() < 0.1 {
			return s.growFancyOak(pos, tx, r)
		}
		return s.growStraightBlob(pos, tx, 4+r.IntN(2), 2)
	case SpruceWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			return s.growMegaSpruce(origin, tx, r)
		}
		return s.growSpruce(pos, tx, r)
	case BirchWood():
		return s.growStraightBlob(pos, tx, 5+r.IntN(2), 2)
	case JungleWood():
		if origin, ok := s.twoByTwoOrigin(pos, tx); ok {
			return s.growMegaJungle(origin, tx, r)
		}
		return s.growStraightBlob(pos, tx, 6+r.IntN(6), 2)
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
		return s.growMangrove(pos, tx, r)
	default:
		return false
	}
}

// growStraightBlob places a simple straight-trunk tree with a blob canopy.
func (s Sapling) growStraightBlob(pos cube.Pos, tx *world.Tx, height, radius int) bool {
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	for y := height - 2; y <= height; y++ {
		layerRadius := radius
		if y == height {
			layerRadius--
		}
		layout.leafLayer(pos.Add(cube.Pos{0, y, 0}), layerRadius, s.Wood, true)
	}
	layout.leafLayer(pos.Add(cube.Pos{0, height + 1, 0}), 0, s.Wood, false)
	return layout.apply()
}

// growFancyOak places a taller oak variant with a side branch and layered canopy.
func (s Sapling) growFancyOak(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 6 + r.IntN(5)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	dir := cube.Directions()[r.IntN(len(cube.Directions()))]
	branchStart := pos.Add(cube.Pos{0, height - 3, 0})
	layout.branch(branchStart, dir, 2, 2, s.Wood)
	layout.leafLayer(pos.Add(cube.Pos{0, height - 1, 0}), 2, s.Wood, true)
	layout.leafLayer(pos.Add(cube.Pos{0, height, 0}), 3, s.Wood, true)
	layout.leafLayer(pos.Add(cube.Pos{0, height + 1, 0}), 2, s.Wood, true)
	layout.leafLayer(pos.Add(cube.Pos{0, height + 2, 0}), 1, s.Wood, false)
	return layout.apply()
}

// growSpruce places a small spruce tree with a narrow layered canopy.
func (s Sapling) growSpruce(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 5 + r.IntN(3)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	for i := 0; i < 5; i++ {
		radius := min(2, 1+(4-i)/2)
		if i == 0 {
			radius = 0
		}
		layout.leafLayer(pos.Add(cube.Pos{0, height - i, 0}), radius, s.Wood, true)
	}
	layout.leafLayer(pos.Add(cube.Pos{0, height + 1, 0}), 0, s.Wood, false)
	return layout.apply()
}

// growAcacia places an acacia tree with a bent trunk and flat canopies.
func (s Sapling) growAcacia(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 5 + r.IntN(4)
	dir := cube.Directions()[r.IntN(len(cube.Directions()))]
	layout := newSaplingTreeLayout(tx)
	for y := 0; y < height-2; y++ {
		layout.set(pos.Add(cube.Pos{0, y, 0}), Log{Wood: s.Wood, Axis: cube.Y})
	}
	top := pos.Add(cube.Pos{0, height - 2, 0})
	branchLength := 2 + r.IntN(2)
	layout.branch(top, dir, branchLength, 2, s.Wood)
	head := top.Add(offset(dir, branchLength)).Add(cube.Pos{0, 2, 0})
	layout.flatCanopy(head, 2, s.Wood)
	side := dir.RotateLeft()
	branchBase := pos.Add(cube.Pos{0, height - 3, 0})
	layout.branch(branchBase, side, 1, 1, s.Wood)
	layout.flatCanopy(branchBase.Add(offset(side, 1)).Add(cube.Pos{0, 1, 0}), 1, s.Wood)
	return layout.apply()
}

// growDarkOak places a 2x2 dark oak or pale oak style tree.
func (s Sapling) growDarkOak(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 6 + r.IntN(3)
	layout := newSaplingTreeLayout(tx)
	layout.twoByTwoTrunk(origin, height, s.Wood)
	for y := height - 2; y <= height; y++ {
		radius := 3
		if y == height {
			radius = 2
		}
		layout.leafSquare(origin.Add(cube.Pos{0, y, 0}), radius, s.Wood, true)
	}
	layout.leafSquare(origin.Add(cube.Pos{0, height + 1, 0}), 1, s.Wood, false)
	return layout.apply()
}

// growCherry places a cherry tree with two raised side branches.
func (s Sapling) growCherry(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 7 + r.IntN(2)
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)

	left, right := cube.Directions()[r.IntN(len(cube.Directions()))], cube.Directions()[r.IntN(len(cube.Directions()))]
	if right == left || right == left.Opposite() {
		right = left.RotateRight()
	}
	leftTop := pos.Add(cube.Pos{0, height - 2, 0})
	rightTop := pos.Add(cube.Pos{0, height - 3, 0})
	layout.branch(leftTop, left, 2, 2, s.Wood)
	layout.branch(rightTop, right, 2, 2, s.Wood)

	layout.leafLayer(pos.Add(cube.Pos{0, height - 1, 0}), 2, s.Wood, true)
	layout.leafLayer(pos.Add(cube.Pos{0, height, 0}), 3, s.Wood, true)
	layout.leafLayer(pos.Add(cube.Pos{0, height + 1, 0}), 2, s.Wood, true)
	layout.leafLayer(leftTop.Add(offset(left, 2)).Add(cube.Pos{0, 1, 0}), 2, s.Wood, true)
	layout.leafLayer(rightTop.Add(offset(right, 2)).Add(cube.Pos{0, 1, 0}), 2, s.Wood, true)
	return layout.apply()
}

// growMegaSpruce places a 2x2 giant spruce and converts nearby dirt to podzol.
func (s Sapling) growMegaSpruce(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 13 + r.IntN(4)
	layout := newSaplingTreeLayout(tx)
	layout.twoByTwoTrunk(origin, height, s.Wood)
	for i := 0; i < 7; i++ {
		radius := min(3, 1+i/2)
		layout.leafSquare(origin.Add(cube.Pos{0, height - i, 0}), radius, s.Wood, true)
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

// growMegaJungle places a 2x2 jungle tree with a broad canopy and side branches.
func (s Sapling) growMegaJungle(origin cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 12 + r.IntN(6)
	layout := newSaplingTreeLayout(tx)
	layout.twoByTwoTrunk(origin, height, s.Wood)
	for i := 0; i < 4; i++ {
		layout.leafSquare(origin.Add(cube.Pos{0, height - i, 0}), 3-i/2, s.Wood, true)
	}
	for _, dir := range cube.Directions() {
		branchY := height - 4 - r.IntN(3)
		layout.branch(origin.Add(cube.Pos{0, branchY, 0}), dir, 2, 1, s.Wood)
		layout.leafLayer(origin.Add(cube.Pos{0, branchY + 1, 0}).Add(offset(dir, 2)), 2, s.Wood, true)
	}
	return layout.apply()
}

// growMangrove places a mangrove-style tree and may attach hanging propagules below leaves.
func (s Sapling) growMangrove(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	height := 6 + r.IntN(3)
	if r.Float64() < 0.85 {
		height += 2 + r.IntN(3)
	}
	layout := newSaplingTreeLayout(tx)
	layout.verticalTrunk(pos, height, s.Wood)
	top := pos.Add(cube.Pos{0, height, 0})
	for i := 0; i < 4; i++ {
		layout.leafLayer(top.Add(cube.Pos{0, -i, 0}), 3-i/2, s.Wood, true)
	}
	for _, dir := range cube.Directions() {
		branchStart := pos.Add(cube.Pos{0, height - 3 - r.IntN(2), 0})
		layout.branch(branchStart, dir, 2, 1, s.Wood)
		layout.leafLayer(branchStart.Add(offset(dir, 2)).Add(cube.Pos{0, 1, 0}), 2, s.Wood, true)
	}
	if !layout.apply() {
		return false
	}
	for _, dir := range cube.Directions() {
		under := top.Add(offset(dir, 2)).Side(cube.FaceDown)
		if _, ok := tx.Block(under).(Air); ok && r.Float64() < 0.35 {
			tx.SetBlock(under, Sapling{Wood: MangroveWood(), Hanging: true, Age: r.IntN(5)}, nil)
		}
	}
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
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				p.set(origin.Add(cube.Pos{dx, y, dz}), Log{Wood: wood, Axis: cube.Y})
			}
		}
	}
}

// branch adds a simple horizontal branch with an optional rise at the end.
func (p *saplingTreeLayout) branch(start cube.Pos, dir cube.Direction, length, rise int, wood WoodType) {
	for i := 1; i <= length; i++ {
		pos := start.Add(offset(dir, i))
		p.set(pos, Log{Wood: wood, Axis: dirAxis(dir)})
		if rise > 0 && i == length {
			for y := 1; y <= rise; y++ {
				p.set(pos.Add(cube.Pos{0, y, 0}), Log{Wood: wood, Axis: cube.Y})
			}
		}
	}
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
	center := origin.Add(cube.Pos{0, 0, 0})
	p.leafLayer(center.Add(cube.Pos{0, 0, 0}), radius, wood, trimCorners)
	p.leafLayer(center.Add(cube.Pos{1, 0, 0}), radius, wood, trimCorners)
	p.leafLayer(center.Add(cube.Pos{0, 0, 1}), radius, wood, trimCorners)
	p.leafLayer(center.Add(cube.Pos{1, 0, 1}), radius, wood, trimCorners)
}

// flatCanopy adds a flat-topped canopy around a center point.
func (p *saplingTreeLayout) flatCanopy(center cube.Pos, radius int, wood WoodType) {
	p.leafLayer(center, radius, wood, true)
	p.leafLayer(center.Add(cube.Pos{0, 1, 0}), radius-1, wood, true)
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

// dirAxis returns the log axis for a horizontal branch direction.
func dirAxis(dir cube.Direction) cube.Axis {
	if dir == cube.North || dir == cube.South {
		return cube.Z
	}
	return cube.X
}
