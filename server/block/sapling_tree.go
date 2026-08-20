package block

import (
	"math"
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// intRange is the range_min and range_max pair of a tree feature. The maximum is exclusive.
type intRange struct {
	min, max int
}

// at returns a value from the range.
func (i intRange) at() int {
	if i.min >= i.max-1 {
		return i.min
	}
	return i.min + rand.IntN(i.max-i.min)
}

// atInclusive returns a value from the range with the maximum included. Only a cherry draws this way.
func (i intRange) atInclusive() int {
	if i.max <= i.min {
		return i.min
	}
	return i.min + rand.IntN(i.max-i.min+1)
}

// heightRange is the trunk_height of a tree feature that states a base and the intervals added to it.
type heightRange struct {
	base      int
	intervals []int
	// inclusive makes each interval contribute a value up to and including itself. Only a cherry is built this way.
	inclusive bool
}

// at returns a height from the range.
func (h heightRange) at() int {
	height := h.base
	for _, interval := range h.intervals {
		if h.inclusive {
			height += rand.IntN(interval + 1)
			continue
		}
		if interval > 0 {
			height += rand.IntN(interval)
		}
	}
	return height
}

// treeChance is a chance field of a tree feature, held as the chance in a hundred that it fires.
type treeChance float64

// roll reports whether the chance fired. One that cannot fail or cannot pass is not rolled for.
func (c treeChance) roll() bool {
	if c >= 100 {
		return true
	}
	if c <= 0 {
		return false
	}
	return float64(c) > rand.Float64()*100
}

// treeDirections holds the four directions a trunk leans and a branch grows in, in the order the game stores them.
var treeDirections = [4]cube.Pos{{0, 0, 1}, {-1, 0, 0}, {0, 0, -1}, {1, 0, 0}}

// treeSoil returns whether a tree may grow on the block passed, holding the may_grow_on of a tree feature.
func treeSoil(b world.Block) bool {
	switch b.(type) {
	case Grass, Dirt, Podzol, Farmland, Mud, MuddyMangroveRoots:
		return true
	}
	return false
}

// treeGrowsThrough is the may_grow_through of a tree feature, tested while a tree is checked for room.
func treeGrowsThrough(tx *world.Tx, pos cube.Pos) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	switch b := tx.Block(pos).(type) {
	case Air, Leaves, Vines, Sapling, ShortGrass, Fern, DoubleTallGrass:
		return true
	default:
		return treeSoil(b)
	}
}

// treeReplaceable returns whether a tree may put the block passed at the position passed. Most of a tree is
// placed through this.
func treeReplaceable(tx *world.Tx, pos cube.Pos, b world.Block) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	switch tx.Block(pos).(type) {
	case Air, Leaves, Vines, Sapling:
		return true
	}
	return replaceableWith(tx, pos, b)
}

// treeLeafReplaceable returns whether a canopy may take the position passed. Canopies grow into air and into
// leaves and stop at anything else.
func treeLeafReplaceable(tx *world.Tx, pos cube.Pos) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	switch tx.Block(pos).(type) {
	case Air, Leaves:
		return true
	}
	return false
}

// treeAir returns whether the position passed holds air. A dark oak canopy grows into nothing else.
func treeAir(tx *world.Tx, pos cube.Pos) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	_, air := tx.Block(pos).(Air)
	return air
}

// treeLog places a log, reporting whether it did.
func treeLog(tx *world.Tx, pos cube.Pos, wood WoodType, axis cube.Axis) bool {
	l := Log{Wood: wood, Axis: axis}
	if !treeReplaceable(tx, pos, l) {
		return false
	}
	tx.SetBlock(pos, l, nil)
	return true
}

// treeLeaf places a leaf through the test passed, reporting whether it did.
func treeLeaf(tx *world.Tx, pos cube.Pos, leaves LeavesType, replaceable func(*world.Tx, cube.Pos) bool) bool {
	if !replaceable(tx, pos) {
		return false
	}
	tx.SetBlock(pos, Leaves{Type: leaves}, nil)
	return true
}

// treeRoom returns whether a tree of the height passed has room above the position passed, being the only check
// made before it starts placing blocks.
func treeRoom(tx *world.Tx, pos cube.Pos, height int) bool {
	if !treeSoil(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	for y := pos[1]; y <= pos[1]+height+1; y++ {
		r := 1
		if y == pos[1] {
			r = 0
		}
		if y >= pos[1]+height-1 {
			r = 2
		}
		for x := pos[0] - r; x <= pos[0]+r; x++ {
			for z := pos[2] - r; z <= pos[2]+r; z++ {
				if !treeGrowsThrough(tx, cube.Pos{x, y, z}) {
					return false
				}
			}
		}
	}
	return true
}

// vineDecoration is the trunk_decoration and canopy_decoration of a tree feature.
type vineDecoration struct {
	// chance is the decoration_chance, the chance that a vine hangs off a side that is exposed.
	chance treeChance
	// steps is the num_steps, how far down the vine reaches. It stops at the first block in the way.
	steps int
}

// treeVines hangs vines off the sides the mask marks as exposed, in the order west, east, north, south.
func treeVines(tx *world.Tx, pos cube.Pos, d vineDecoration, sides [4]bool) {
	if d.chance <= 0 || d.steps <= 0 {
		return
	}
	for i, v := range [4]struct {
		side cube.Pos
		vine Vines
	}{
		{cube.Pos{-1}, Vines{EastDirection: true}},
		{cube.Pos{1}, Vines{WestDirection: true}},
		{cube.Pos{0, 0, -1}, Vines{SouthDirection: true}},
		{cube.Pos{0, 0, 1}, Vines{NorthDirection: true}},
	} {
		if !sides[i] || !d.chance.roll() {
			continue
		}
		at := pos.Add(v.side)
		if !treeAir(tx, at) {
			continue
		}
		// The vine reaches down as far as it is told to, stopping at the first block in its way.
		for range d.steps {
			tx.SetBlock(at, v.vine, nil)
			if at = at.Side(cube.FaceDown); !treeAir(tx, at) {
				break
			}
		}
	}
}

// treeCanopy is the part of a tree its trunk carries and places itself.
type treeCanopy interface {
	// place places the canopy at the position passed, told how wide the trunk carrying it is.
	place(tx *world.Tx, origin cube.Pos, width int, leaves LeavesType)
}

// treeTrunk is the whole of a tree: it places its logs and then hands off to the canopy it carries.
type treeTrunk interface {
	// height returns the number of logs the trunk is made of. It is drawn before anything is placed.
	height() int
	// place places the tree, reporting whether it was placed at all.
	place(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) bool
}

// simpleCanopy is the canopy of an oak, a birch and a jungle tree: a stack of squares with corners left out.
type simpleCanopy struct {
	// offset is the canopy_offset, holding the rows the canopy covers relative to its own position.
	offset intRange
	// minWidth is the min_width, the radius every row has before the slope is taken off it.
	minWidth int
	// slopeRun and slopeRise are the canopy_slope. The rise is held as one over itself, which is how it is read.
	slopeRun  int
	slopeRise float64
	// variation holds, for every row from the lowest up, the chance that a corner of it is left out.
	variation []treeChance
	// vines is the canopy_decoration of a jungle tree.
	vines vineDecoration
}

// radius returns the radius of the row at the offset passed.
func (c simpleCanopy) radius(offset int) int {
	slope := func(o int) int {
		return int(float64(o*c.slopeRun) * c.slopeRise)
	}
	return c.minWidth + slope(c.offset.max) - slope(offset)
}

// place ...
func (c simpleCanopy) place(tx *world.Tx, origin cube.Pos, _ int, leaves LeavesType) {
	for offset := c.offset.min; offset <= c.offset.max; offset++ {
		r, variation := c.radius(offset), c.variation[min(offset-c.offset.min, len(c.variation)-1)]
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				// Only the four corners of a row are ever left out, and each of them is rolled for on its own.
				if abs(dx) == r && abs(dz) == r && variation.roll() {
					continue
				}
				treeLeaf(tx, origin.Add(cube.Pos{dx, offset, dz}), leaves, treeLeafReplaceable)
			}
		}
	}
	if c.vines.chance <= 0 {
		return
	}
	// The vines are hung in a second pass, off anything holding the leaves of the tree by then.
	for offset := c.offset.min; offset <= c.offset.max; offset++ {
		r := c.radius(offset)
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				at := origin.Add(cube.Pos{dx, offset, dz})
				if l, ok := tx.Block(at).(Leaves); ok && l.Type == leaves {
					treeVines(tx, at, c.vines, [4]bool{true, true, true, true})
				}
			}
		}
	}
}

// spruceCanopy is the spruce_canopy of a tree feature, widening and narrowing again below a single leaf.
type spruceCanopy struct {
	// lowerOffset and upperOffset are how far below the ground it reaches and how far above its position it starts.
	lowerOffset, upperOffset intRange
	// maxRadius is the radius the widest row of the canopy may reach.
	maxRadius intRange
}

// place ...
func (c spruceCanopy) place(tx *world.Tx, origin cube.Pos, _ int, leaves LeavesType) {
	// The canopy reaches down to the ground the tree grew from, so the ground is looked for first.
	ground := origin[1]
	for {
		if ground <= tx.Range().Min() {
			return
		}
		if treeSoil(tx.Block(cube.Pos{origin[0], ground - 1, origin[2]})) {
			break
		}
		ground--
	}

	lower, upper, widest := c.lowerOffset.at(), c.upperOffset.at(), c.maxRadius.at()
	radius := rand.IntN(max(1, c.maxRadius.max-c.maxRadius.min))
	reach, narrowest := c.maxRadius.max-c.maxRadius.min-1, 0

	top := origin[1] + upper
	for row := 0; row < top-(ground+lower)+1; row++ {
		y := top - row
		for dx := -radius; dx <= radius; dx++ {
			for dz := -radius; dz <= radius; dz++ {
				if radius != 0 && abs(dx) == radius && abs(dz) == radius {
					continue
				}
				treeLeaf(tx, cube.Pos{origin[0] + dx, y, origin[2] + dz}, leaves, treeLeafReplaceable)
			}
		}
		next := min(reach+1, widest)
		if reach <= radius {
			radius, narrowest, reach = narrowest, 1, next
			continue
		}
		radius++
	}
}

// roofedCanopy is the roofed_canopy of a dark oak and a pale oak: a slab of leaves drooping at its corners.
type roofedCanopy struct {
	// height is the canopy_height. Three is taken off it when read, leaving a dark oak a single middle layer.
	height int
	// coreWidth is the core_width, the width of the trunk the canopy sits on.
	coreWidth int
	// outerRadius is the radius of the layers below and above the canopy, innerRadius that of those between.
	outerRadius, innerRadius int
}

// layers returns the number of layers the canopy carries between its two rims.
func (c roofedCanopy) layers() int {
	return max(1, c.height-3)
}

// place ...
func (c roofedCanopy) place(tx *world.Tx, origin cube.Pos, width int, leaves LeavesType) {
	size := width
	if size <= 0 {
		size = c.coreWidth
	}
	r := c.outerRadius
	for dx := -r; dx <= size-1+r; dx++ {
		for dz := -r; dz <= size-1+r; dz++ {
			treeLeaf(tx, origin.Add(cube.Pos{dx, -1, dz}), leaves, treeAir)
			// The layer above the canopy has three cells cut from each corner, measured outward from the trunk.
			if outX, outZ := max(-dx, dx-size+1, 0), max(-dz, dz-size+1, 0); outX+outZ >= r*2-1 {
				continue
			}
			treeLeaf(tx, origin.Add(cube.Pos{dx, c.layers(), dz}), leaves, treeAir)
		}
	}

	r = c.innerRadius
	for y := 0; y < c.layers(); y++ {
		for dx := -r; dx < size+r; dx++ {
			for dz := -r; dz < size+r; dz++ {
				if abs(dx) >= r && abs(dz) >= r {
					continue
				}
				treeLeaf(tx, origin.Add(cube.Pos{dx, y, dz}), leaves, treeAir)
			}
		}
	}

	if rand.IntN(2) != 0 {
		return
	}
	for dx := range size {
		for dz := range size {
			treeLeaf(tx, origin.Add(cube.Pos{dx, c.layers() + 1, dz}), leaves, treeAir)
		}
	}
}

// acaciaCanopy is the canopy an acacia carries at its top and every branch carries at its end.
type acaciaCanopy struct {
	// size is the canopy_size, the radius of the widest row of the canopy.
	size int
	// simplified makes the row above the widest one a solid square. The widest row is laid either way.
	simplified bool
}

// place ...
func (c acaciaCanopy) place(tx *world.Tx, origin cube.Pos, _ int, leaves LeavesType) {
	m := max(1, c.size)
	if c.simplified {
		for dx := 1 - m; dx < m; dx++ {
			for dz := 1 - m; dz < m; dz++ {
				treeLeaf(tx, origin.Add(cube.Pos{dx, 1, dz}), leaves, treeLeafReplaceable)
			}
		}
	} else {
		k := max(2, c.size)
		for dx := 2 - k; dx < k-1; dx++ {
			for dz := 2 - k; dz < k-1; dz++ {
				treeLeaf(tx, origin.Add(cube.Pos{dx, 1, dz}), leaves, treeLeafReplaceable)
			}
		}
		for _, arm := range [4]cube.Pos{{m - 1, 1}, {1 - m, 1}, {0, 1, m - 1}, {0, 1, 1 - m}} {
			treeLeaf(tx, origin.Add(arm), leaves, treeLeafReplaceable)
		}
	}
	// The widest row is laid whether the canopy was simplified or not.
	for dx := -c.size; dx <= c.size; dx++ {
		for dz := -c.size; dz <= c.size; dz++ {
			if abs(dx) == c.size && abs(dz) == c.size {
				continue
			}
			treeLeaf(tx, origin.Add(cube.Pos{dx, 0, dz}), leaves, treeLeafReplaceable)
		}
	}
}

// treeRow places a row of leaves around a trunk of the size passed, keeping a cell outside the radius if the one
// towards the trunk is inside it.
func treeRow(tx *world.Tx, pos cube.Pos, r, sizeX, sizeZ int, leaves LeavesType) {
	full := func(tx *world.Tx, p cube.Pos) bool {
		return treeReplaceable(tx, p, Leaves{Type: leaves})
	}
	for dx := -r; dx < sizeX+r; dx++ {
		for dz := -r; dz < sizeZ+r; dz++ {
			if dx*dx+dz*dz > r*r {
				if (dx-1)*(dx-1)+(dz-1)*(dz-1) > r*r && dx*dx+(dz-1)*(dz-1) > r*r && (dx-1)*(dx-1)+dz*dz > r*r {
					continue
				}
			}
			treeLeaf(tx, pos.Add(cube.Pos{dx, 0, dz}), leaves, full)
		}
	}
}

// megaCanopy is the mega_canopy of a tree feature, carried by a mega jungle tree and widening on the way down.
type megaCanopy struct {
	// height is the canopy_height, the number of rows the canopy covers.
	height intRange
	// baseRadius is the base_radius, the radius of the row at the top of the canopy.
	baseRadius int
	// coreWidth is the core_width, the width of the trunk the canopy sits on.
	coreWidth int
	// simplified replaces the trunk it was told about with its own core_width.
	simplified bool
}

// place ...
func (c megaCanopy) place(tx *world.Tx, origin cube.Pos, width int, leaves LeavesType) {
	size := width
	if c.simplified || size <= 0 {
		size = c.coreWidth
	}
	h := c.height.at()
	for row := h - 1; row >= 0; row-- {
		treeRow(tx, cube.Pos{origin[0], origin[1] - row, origin[2]}, c.coreWidth/2+c.baseRadius+row, size, size, leaves)
	}
}

// megaPineCanopy is the mega_pine_canopy of a mega spruce: a cone frilled with rows repeating the width below.
type megaPineCanopy struct {
	// height is the canopy_height, the number of rows the canopy covers.
	height intRange
	// baseRadius is the base_radius, added to the radius of every row.
	baseRadius int
	// radiusStep is the radius_step_modifier, deciding how quickly the cone widens on the way down.
	radiusStep float64
	// coreWidth is the core_width, the width of the trunk the canopy sits on.
	coreWidth int
}

// place ...
func (c megaPineCanopy) place(tx *world.Tx, origin cube.Pos, width int, leaves LeavesType) {
	size := width
	if size <= 0 {
		size = c.coreWidth
	}
	h, last := c.height.at(), 0
	for down := h; down >= 0; down-- {
		y := origin[1] - down
		r := c.baseRadius + int(math.Floor(float64(down)/float64(h)*c.radiusStep))
		frill := r
		if down >= 1 && r == last && y%2 == 0 {
			frill++
		}
		treeRow(tx, cube.Pos{origin[0], y, origin[2]}, frill, size, size, leaves)
		last = r
	}
}

// cherryCanopy is the cherry_canopy of a tree feature. Every layer of it is a square with its corners rolled for.
type cherryCanopy struct {
	// height holds the rows the crown covers and radius the value every row of it takes its own radius from.
	height, radius intRange
	// coreWidth is the width of the trunk the crown sits on.
	coreWidth int
	// wideBottomHole is the wide_bottom_layer_hole_chance, rolled for the outer ring of the layer below the middle.
	wideBottomHole treeChance
	// cornerHole is the corner_hole_chance, rolled for the corners of a layer and the cells beyond its diagonal.
	cornerHole treeChance
	// hangingLeaves and hangingExtension are the hanging_leaves_chance and hanging_leaves_extension_chance.
	hangingLeaves, hangingExtension treeChance
}

// place ...
func (c cherryCanopy) place(tx *world.Tx, origin cube.Pos, _ int, leaves LeavesType) {
	r, h := c.radius.atInclusive(), c.height.atInclusive()
	// The crown is built from its top down, the two layers below its middle hanging leaves off their edges.
	widest := c.coreWidth + r - 2
	c.layer(tx, origin, h-3, widest-2, leaves)
	c.layer(tx, origin, h-4, widest-1, leaves)
	for dy := h - 5; dy >= 0; dy-- {
		c.layer(tx, origin, dy, widest, leaves)
	}
	for _, l := range [2]struct{ dy, radius int }{{-1, widest}, {-2, widest - 1}} {
		c.layer(tx, origin, l.dy, l.radius, leaves)
		c.hang(tx, origin, l.dy, l.radius, leaves)
	}
}

// layer places one layer of the crown.
func (c cherryCanopy) layer(tx *world.Tx, origin cube.Pos, dy, r int, leaves LeavesType) {
	w := c.coreWidth
	for dx := -r; dx < w+r; dx++ {
		for dz := -r; dz < w+r; dz++ {
			adx, adz := abs(dx), abs(dz)
			if w >= 2 {
				adx, adz = min(adx, abs(dx-(w-1))), min(adz, abs(dz-(w-1)))
			}
			// The whole outer ring of the layer below the middle is rolled for before anything else.
			if dy == -1 && (adx == r || adz == r) && c.wideBottomHole.roll() {
				continue
			}
			corner := adx == r && adz == r
			if (corner || (r >= 3 && adx+adz > 2*r-2)) && c.cornerHole.roll() {
				continue
			}
			treeLeaf(tx, origin.Add(cube.Pos{dx, dy, dz}), leaves, treeLeafReplaceable)
		}
	}
}

// hang hangs leaves below the edges of the layer passed. The edges are walked one after the other, so a corner
// is only ever rolled for once.
func (c cherryCanopy) hang(tx *world.Tx, origin cube.Pos, dy, r int, leaves LeavesType) {
	e := c.coreWidth + r - 1
	for _, edge := range [4]func(t int) cube.Pos{
		func(t int) cube.Pos { return cube.Pos{e, dy, t} },
		func(t int) cube.Pos { return cube.Pos{-t, dy, e} },
		func(t int) cube.Pos { return cube.Pos{-r, dy, -t} },
		func(t int) cube.Pos { return cube.Pos{t, dy, -r} },
	} {
		for t := -r; t < e; t++ {
			above := origin.Add(edge(t))
			if l, ok := tx.Block(above).(Leaves); !ok || l.Type != leaves {
				continue
			}
			if !c.hangingLeaves.roll() {
				continue
			}
			// The leaf below the edge has to go down before the one below that is rolled for at all.
			below := above.Side(cube.FaceDown)
			if !treeLeaf(tx, below, leaves, treeLeafReplaceable) {
				continue
			}
			if c.hangingExtension.roll() {
				treeLeaf(tx, below.Side(cube.FaceDown), leaves, treeLeafReplaceable)
			}
		}
	}
}

// simpleTrunk is a straight trunk one block wide, grown by an oak, a birch, a jungle tree and a spruce.
type simpleTrunk struct {
	trunkHeight    intRange
	heightModifier intRange
	canopy         treeCanopy
	// vines is the trunk_decoration of a jungle tree.
	vines vineDecoration
}

// height ...
func (t simpleTrunk) height() int {
	return t.trunkHeight.at()
}

// place ...
func (t simpleTrunk) place(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) bool {
	if !treeRoom(tx, pos, height) {
		return false
	}
	logs, top, placed := height+t.heightModifier.at(), pos, false
	for i := range logs {
		at := pos.Add(cube.Pos{0, i})
		if !treeLog(tx, at, wood, cube.Y) {
			// A level the trunk cannot reach is passed over, leaving a gap behind rather than ending the tree.
			continue
		}
		treeVines(tx, at, t.vines, [4]bool{true, true, true, true})
		top, placed = at, true
	}
	if !placed {
		// A trunk that got nowhere carries nothing and turns nothing over.
		return false
	}
	t.canopy.place(tx, top.Side(cube.FaceUp), 1, leaves)

	if below := pos.Side(cube.FaceDown); !treeSoil(tx.Block(below)) {
		return true
	} else if _, dirt := tx.Block(below).(Dirt); !dirt {
		tx.SetBlock(below, Dirt{}, nil)
	}
	return true
}

// acaciaTrunk is the trunk of an acacia, a dark oak and a pale oak. An acacia leans near its top and carries one
// branch; a dark oak stays upright and carries a ring of them.
type acaciaTrunk struct {
	trunkHeight heightRange
	trunkWidth  int
	// minHeightForCanopy is how far up the trunk a log has to be for the canopy to be carried by it.
	minHeightForCanopy int
	// leanHeight and leanSteps are the trunk_lean, where the lean starts and how often it steps aside.
	leanHeight, leanSteps intRange
	// lean makes the trunk lean at all. A dark oak and a pale oak have it off and only their canopy moves.
	lean bool
	// branchLength, branchPosition and branchChance are the branches of the feature.
	branchLength, branchPosition intRange
	branchChance                 treeChance
	branchCanopy                 treeCanopy
	canopy                       treeCanopy
}

// height ...
func (t acaciaTrunk) height() int {
	return t.trunkHeight.at()
}

// place ...
func (t acaciaTrunk) place(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) bool {
	if !treeRoom(tx, pos, height) {
		return false
	}
	// The block below the trunk is turned over before the trunk is placed rather than after it.
	for dx := range t.trunkWidth {
		for dz := range t.trunkWidth {
			below := pos.Add(cube.Pos{dx, -1, dz})
			if _, dirt := tx.Block(below).(Dirt); !dirt {
				tx.SetBlock(below, Dirt{}, nil)
			}
		}
	}

	dir := rand.IntN(4)
	leanHeight, leanSteps := t.leanHeight.at(), t.leanSteps.at()
	leanStart, cur := height-leanHeight, pos
	var anchor cube.Pos
	var anchored bool

	for i := range height {
		if i >= leanStart && leanSteps > 0 {
			cur, leanSteps = cur.Add(treeDirections[dir]), leanSteps-1
		}
		cur[1] = pos[1] + i
		base, y := cur, cur[1]
		if !t.lean {
			base, y = pos, pos[1]+i
		}
		for dx := range t.trunkWidth {
			for dz := range t.trunkWidth {
				at := cube.Pos{base[0] + dx, y, base[2] + dz}
				if !treeLog(tx, at, wood, cube.Y) {
					continue
				}
				if i >= t.minHeightForCanopy {
					anchor, anchored = at, true
				}
			}
		}
	}
	if !anchored {
		return true
	}
	t.canopy.place(tx, anchor, t.trunkWidth, leaves)
	t.branches(tx, pos, anchor, dir, height, leanStart, wood, leaves)
	return true
}

// branches grows the branches leaving the trunk, each of which carries a canopy of its own.
func (t acaciaTrunk) branches(tx *world.Tx, pos, anchor cube.Pos, dir, height, leanStart int, wood WoodType, leaves LeavesType) {
	if t.lean {
		// A leaning trunk grows a single branch, and only if it takes a different direction than the lean did.
		to := rand.IntN(4)
		if to == dir || !t.branchChance.roll() {
			return
		}
		// The branch leaves the column the tree grew from rather than the leaning top the canopy sits on.
		position, length := t.branchPosition.at(), t.branchLength.at()
		cur := cube.Pos{pos[0], anchor[1], pos[2]}
		for y := leanStart - position; y < height && length > 0; y, length = y+1, length-1 {
			if y < 1 {
				// A level at or below the ground is stepped over without the branch reaching out any further.
				continue
			}
			cur = cur.Add(treeDirections[to])
			cur[1] = pos[1] + y
			treeLog(tx, cur, wood, cube.Y)
		}
		if t.branchCanopy != nil {
			t.branchCanopy.place(tx, cur, 1, leaves)
		}
		return
	}
	for dx := -1; dx <= t.trunkWidth; dx++ {
		for dz := -1; dz <= t.trunkWidth; dz++ {
			if dx >= 0 && dx < t.trunkWidth && dz >= 0 && dz < t.trunkWidth {
				continue
			}
			if !t.branchChance.roll() {
				continue
			}
			length, position := t.branchLength.at(), t.branchPosition.at()
			for k := range length {
				treeLog(tx, cube.Pos{pos[0] + dx, anchor[1] - k - position, pos[2] + dz}, wood, cube.Y)
			}
			if t.branchCanopy != nil {
				// The canopy hangs off the anchor, a trunk corner away from the logs of the branch.
				t.branchCanopy.place(tx, cube.Pos{anchor[0] + dx, anchor[1] - position, anchor[2] + dz}, 1, leaves)
			}
		}
	}
}

// megaTrunk is the trunk of a mega spruce and a mega jungle tree, grown from a two by two square of saplings.
type megaTrunk struct {
	trunkHeight heightRange
	trunkWidth  int
	canopy      treeCanopy
	// branchLength, branchInterval and branchSlope are the branches a mega jungle tree carries.
	branchLength   int
	branchInterval intRange
	branchSlope    float64
	// altitudeFactor is the branch_altitude_factor, the part of the trunk branches leave it over.
	altitudeFactor [2]float64
	branchCanopy   treeCanopy
	// vines is the trunk_decoration of a mega jungle tree.
	vines vineDecoration
	// base is the base_block, which every column the trunk stands on is turned over to.
	base world.Block
	// cluster is the base_cluster a mega spruce spreads around itself.
	cluster baseCluster
}

// height ...
func (t megaTrunk) height() int {
	return t.trunkHeight.at()
}

// place ...
func (t megaTrunk) place(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) bool {
	if !treeSoil(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	// A wide trunk needs its own footprint at the bottom and one block around it the rest of the way up.
	for i := 0; i <= height+1; i++ {
		r := t.trunkWidth
		if i == 0 {
			r = max(0, t.trunkWidth-1)
		}
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				if !treeGrowsThrough(tx, pos.Add(cube.Pos{dx, i, dz})) {
					return false
				}
			}
		}
	}

	// The canopy and the branches go down first, so that the trunk is laid over the leaves it runs into.
	t.canopy.place(tx, pos.Add(cube.Pos{0, height}), t.trunkWidth, leaves)
	t.branches(tx, pos, height, wood, leaves)

	for i := range height {
		for dx := range t.trunkWidth {
			for dz := range t.trunkWidth {
				at := pos.Add(cube.Pos{dx, i, dz})
				if !treeLog(tx, at, wood, cube.Y) {
					continue
				}
				treeVines(tx, at, t.vines, [4]bool{
					dx == 0, dx == t.trunkWidth-1, dz == 0, dz == t.trunkWidth-1,
				})
			}
		}
	}

	// Every column the trunk stands on is turned over, whatever it held.
	for dx := range t.trunkWidth {
		for dz := range t.trunkWidth {
			if below := pos.Add(cube.Pos{dx, -1, dz}); tx.Block(below) != t.base {
				tx.SetBlock(below, t.base, nil)
			}
		}
	}
	t.cluster.place(tx, pos, t.trunkWidth)
	return true
}

// branches grows the branches a mega jungle tree carries, each ending in a canopy of its own.
func (t megaTrunk) branches(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) {
	if t.branchLength <= 0 {
		return
	}
	y := pos[1] - t.branchInterval.at() + int(t.altitudeFactor[1]*float64(height))
	for y > pos[1]+int(t.altitudeFactor[0]*float64(height)) {
		angle := rand.Float64() * 2 * math.Pi
		bx, bz := pos[0]+1, pos[2]+1
		base := float64(y-1) - float64(t.branchLength-1)*t.branchSlope
		for b := range t.branchLength {
			bx = pos[0] + int(1.5+math.Cos(angle)*float64(b))
			bz = pos[2] + int(1.5+math.Sin(angle)*float64(b))
			treeLog(tx, cube.Pos{bx, int(base + float64(b)*t.branchSlope), bz}, wood, cube.Y)
		}
		if t.branchCanopy != nil {
			t.branchCanopy.place(tx, cube.Pos{bx, y, bz}, 1, leaves)
		}
		y -= t.branchInterval.at()
	}
}

// baseCluster is the base_cluster of a tree feature, spread around the base of a mega spruce.
type baseCluster struct {
	// clusters is the num_clusters, spread besides the four at the corners, and radius the cluster_radius.
	clusters, radius int
}

// place spreads the patches around the trunk passed.
func (c baseCluster) place(tx *world.Tx, pos cube.Pos, width int) {
	if c.clusters <= 0 {
		return
	}
	for _, corner := range [4]cube.Pos{
		{pos[0] - 1, pos[1], pos[2] - 1}, {pos[0] + width, pos[1], pos[2] - 1},
		{pos[0] - 1, pos[1], pos[2] + width}, {pos[0] + width, pos[1], pos[2] + width},
	} {
		c.patch(tx, corner, pos[1])
	}
	for range c.clusters {
		// Only the edge of the eight by eight grid is used: a draw landing inside spreads nothing at all.
		n := rand.IntN(64)
		if q := n % 8; q == 0 || q == 7 || n < 8 || n >= 56 {
			c.patch(tx, cube.Pos{pos[0] - 3 + q, pos[1], pos[2] - 3 + n/8}, pos[1])
		}
	}
}

// patch spreads one patch of podzol around the position passed.
func (c baseCluster) patch(tx *world.Tx, at cube.Pos, y int) {
	for dx := -c.radius; dx <= c.radius; dx++ {
		for dz := -c.radius; dz <= c.radius; dz++ {
			if abs(dx) == c.radius && abs(dz) == c.radius {
				continue
			}
			// Every column is searched downwards for the first block the patch may take over.
			for dy := 2; dy >= -3; dy-- {
				pos := cube.Pos{at[0] + dx, y - 1 + dy, at[2] + dz}
				switch tx.Block(pos).(type) {
				case Podzol:
					dy = -4
				case Grass, Dirt:
					tx.SetBlock(pos, Podzol{}, nil)
					dy = -4
				}
			}
		}
	}
}

// cherryTrunk is the trunk of a cherry tree, growing one branch, two, or two and a crown, each as often.
type cherryTrunk struct {
	trunkHeight heightRange
	// branchLength is the branch_horizontal_length, how far a branch reaches out before it rises.
	branchLength intRange
	// branchOffset is how far below the top of the trunk a branch leaves it, read once for each of the two.
	branchOffset intRange
	// branchRise is how far a branch rises over its length.
	branchRise intRange
	canopy     treeCanopy
}

// height ...
func (t cherryTrunk) height() int {
	return t.trunkHeight.at()
}

// place ...
func (t cherryTrunk) place(tx *world.Tx, pos cube.Pos, height int, wood WoodType, leaves LeavesType) bool {
	if !treeRoom(tx, pos, height) {
		return false
	}
	shape := rand.IntN(3)
	start := max(0, t.branchOffset.atInclusive()+height-1)
	end := max(0, t.branchOffset.at()+height-1)
	if start <= end {
		end++
	}
	dir := rand.IntN(4)

	logs := height
	switch shape {
	case 0:
		logs = start + 1
	case 1:
		logs = max(start, end) + 1
	}
	below := pos.Side(cube.FaceDown)
	if _, dirt := tx.Block(below).(Dirt); !dirt && treeSoil(tx.Block(below)) {
		tx.SetBlock(below, Dirt{}, nil)
	}
	for i := range logs {
		// A level the trunk cannot reach is passed over, leaving a gap behind rather than ending the trunk.
		treeLog(tx, pos.Add(cube.Pos{0, i}), wood, cube.Y)
	}

	// Both branches grow before any crown does, so that a crown gives way to the logs of the branch beside it.
	crowns := make([]cube.Pos, 0, 3)
	if shape == 2 {
		crowns = append(crowns, pos.Add(cube.Pos{0, height}))
	}
	crowns = append(crowns, t.branch(tx, pos, treeDirections[dir], height, start, logs, wood))
	if shape >= 1 {
		to := cube.Pos{-treeDirections[dir][0], 0, -treeDirections[dir][2]}
		crowns = append(crowns, t.branch(tx, pos, to, height, end, logs, wood))
	}
	for _, crown := range crowns {
		t.canopy.place(tx, crown, 1, leaves)
	}
	return true
}

// branch grows one branch out of the trunk and returns the position its crown is carried at.
func (t cherryTrunk) branch(tx *world.Tx, pos, dir cube.Pos, height, at, logs int, wood WoodType) cube.Pos {
	rise := height - 1 + t.branchRise.atInclusive()
	// A branch is made longer when it ends below where it left the trunk, or leaves it below its topmost log.
	long := at < logs-1 || rise < at
	length := t.branchLength.atInclusive()
	if long {
		length++
	}

	axis := cube.X
	if dir[2] != 0 {
		axis = cube.Z
	}
	cur, steps := cube.Pos{pos[0], pos[1] + at, pos[2]}, 1
	if long {
		steps = 2
	}
	for range steps {
		cur = cur.Add(dir)
		treeLog(tx, cur, wood, axis)
	}
	// The branch walks to where it ends, each step up or sideways as often as the height left is of the distance.
	target := cube.Pos{pos[0] + dir[0]*length, pos[1] + rise, pos[2] + dir[2]*length}
	for cur != target {
		dx, dy, dz := target[0]-cur[0], target[1]-cur[1], target[2]-cur[2]
		if rand.Float64() < float64(abs(dy))/float64(abs(dx)+abs(dy)+abs(dz)) {
			cur[1] += sign(dy)
			treeLog(tx, cur, wood, cube.Y)
			continue
		}
		if abs(dx) >= abs(dz) {
			cur[0] += sign(dx)
		} else {
			cur[2] += sign(dz)
		}
		treeLog(tx, cur, wood, axis)
	}
	return target.Side(cube.FaceUp)
}

// The trees below hold the values of the tree feature of the same name, as Bedrock ships them under
// behavior_packs/vanilla_*/features. Only the trees a sapling grows into are listed.
var (
	// blobVariation is the variation_chance of an oak, a birch and a jungle tree, from the lowest row up.
	blobVariation = []treeChance{50, 50, 50, 100}

	// oakTree holds the values of oak_tree_feature.json.
	oakTree = simpleTrunk{
		trunkHeight: intRange{4, 7},
		canopy: simpleCanopy{
			offset: intRange{-3, 0}, minWidth: 1, slopeRun: 1, slopeRise: 0.5, variation: blobVariation,
		},
	}
	// birchTree holds the values of birch_tree_feature.json.
	birchTree = simpleTrunk{
		trunkHeight: intRange{5, 8},
		canopy: simpleCanopy{
			offset: intRange{-3, 0}, minWidth: 1, slopeRun: 1, slopeRise: 0.5, variation: blobVariation,
		},
	}
	// jungleTree holds the values of jungle_tree_feature.json.
	jungleTree = simpleTrunk{
		trunkHeight: intRange{4, 11},
		vines:       vineDecoration{chance: 33.33, steps: 1},
		canopy: simpleCanopy{
			offset: intRange{-3, 0}, minWidth: 1, slopeRun: 1, slopeRise: 0.5, variation: blobVariation,
			vines: vineDecoration{chance: 25, steps: 4},
		},
	}
	// spruceTree holds the values of spruce_tree_feature.json.
	spruceTree = simpleTrunk{
		trunkHeight:    intRange{6, 10},
		heightModifier: intRange{-2, 1},
		canopy: spruceCanopy{
			lowerOffset: intRange{1, 3}, upperOffset: intRange{0, 3}, maxRadius: intRange{2, 4},
		},
	}
	// savannaTree holds the values of savanna_tree_feature.json, which an acacia sapling grows into.
	savannaTree = acaciaTrunk{
		trunkHeight:        heightRange{base: 5, intervals: []int{3, 3}},
		trunkWidth:         1,
		minHeightForCanopy: 3,
		leanHeight:         intRange{1, 5},
		leanSteps:          intRange{1, 4},
		lean:               true,
		branchLength:       intRange{1, 4},
		branchPosition:     intRange{1, 3},
		branchChance:       100,
		branchCanopy:       acaciaCanopy{size: 2, simplified: true},
		canopy:             acaciaCanopy{size: 3},
	}
	// roofedTree holds the values of roofed_tree_feature.json, which a dark oak sapling grows into.
	// pale_oak_tree_feature.json holds the same values.
	roofedTree = acaciaTrunk{
		trunkHeight:        heightRange{base: 6, intervals: []int{3, 2}},
		trunkWidth:         2,
		minHeightForCanopy: 3,
		leanHeight:         intRange{0, 4},
		leanSteps:          intRange{0, 3},
		branchLength:       intRange{2, 5},
		branchPosition:     intRange{1, 1},
		branchChance:       100.0 / 3.0,
		branchCanopy:       acaciaCanopy{size: 2, simplified: true},
		canopy:             roofedCanopy{height: 4, coreWidth: 2, outerRadius: 2, innerRadius: 3},
	}
	// megaSpruceTree holds the values of mega_spruce_tree_feature.json.
	megaSpruceTree = megaTrunk{
		trunkHeight: heightRange{base: 13, intervals: []int{3, 15}},
		trunkWidth:  2,
		canopy: megaPineCanopy{
			height: intRange{13, 18}, baseRadius: 0, radiusStep: 3.5, coreWidth: 2,
		},
		base:    Podzol{},
		cluster: baseCluster{clusters: 5, radius: 2},
	}
	// megaJungleTree holds the values of mega_jungle_tree_feature.json.
	megaJungleTree = megaTrunk{
		trunkHeight:    heightRange{base: 10, intervals: []int{3, 20}},
		trunkWidth:     2,
		branchLength:   5,
		branchInterval: intRange{2, 6},
		branchSlope:    0.5,
		altitudeFactor: [2]float64{0.5, 1},
		base:           Dirt{},
		vines:          vineDecoration{chance: 100.0 / 3.0, steps: 1},
		canopy:         megaCanopy{height: intRange{3, 3}, baseRadius: 2, coreWidth: 2},
		branchCanopy:   megaCanopy{height: intRange{2, 4}, baseRadius: 1, coreWidth: 1, simplified: true},
	}
	// cherryTree holds the values of cherry_tree_feature.json.
	cherryTree = cherryTrunk{
		trunkHeight:  heightRange{base: 7, intervals: []int{1}, inclusive: true},
		branchLength: intRange{2, 4},
		branchOffset: intRange{-4, -3},
		branchRise:   intRange{-1, 0},
		canopy: cherryCanopy{
			height: intRange{5, 5}, radius: intRange{4, 4}, coreWidth: 1,
			wideBottomHole: 25, cornerHole: 50, hangingLeaves: 16.6666, hangingExtension: 33.3333,
		},
	}
)

// growTree grows the tree of the sapling type passed at the position passed, reporting whether it grew. Trees grown
// from a two by two square of saplings are grown from the north west corner of that square.
func growTree(t SaplingType, pos cube.Pos, tx *world.Tx) bool {
	wood, leaves := t.Wood(), t.Leaves()

	var trunk treeTrunk
	switch t {
	case DarkOakSapling(), PaleOakSapling():
		corner, ok := saplingSquare(pos, tx, t)
		if !ok {
			return false
		}
		pos, trunk = corner, roofedTree
	case SpruceSapling():
		if corner, ok := saplingSquare(pos, tx, t); ok {
			pos, trunk = corner, megaSpruceTree
			break
		}
		trunk = spruceTree
	case JungleSapling():
		if corner, ok := saplingSquare(pos, tx, t); ok {
			pos, trunk = corner, megaJungleTree
			break
		}
		trunk = jungleTree
	case OakSapling():
		trunk = oakTree
	case BirchSapling():
		trunk = birchTree
	case AcaciaSapling():
		trunk = savannaTree
	case CherrySapling():
		trunk = cherryTree
	default:
		return false
	}
	return trunk.place(tx, pos, trunk.height(), wood, leaves)
}

// saplingSquare returns the north west corner of a two by two square of saplings of the type passed that the position
// passed is part of. The second return value is false if the sapling is not part of such a square.
func saplingSquare(pos cube.Pos, tx *world.Tx, t SaplingType) (cube.Pos, bool) {
	for _, corner := range []cube.Pos{{}, {-1}, {0, 0, -1}, {-1, 0, -1}} {
		corner = pos.Add(corner)
		if saplingAt(corner, tx, t) && saplingAt(corner.Add(cube.Pos{1}), tx, t) &&
			saplingAt(corner.Add(cube.Pos{0, 0, 1}), tx, t) && saplingAt(corner.Add(cube.Pos{1, 0, 1}), tx, t) {
			return corner, true
		}
	}
	return cube.Pos{}, false
}

// saplingAt returns whether a sapling of the type passed is at the position passed.
func saplingAt(pos cube.Pos, tx *world.Tx, t SaplingType) bool {
	s, ok := tx.Block(pos).(Sapling)
	return ok && s.Type == t
}
