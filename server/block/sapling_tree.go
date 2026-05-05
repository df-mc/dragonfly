package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type saplingTreePlan struct {
	blocks map[cube.Pos]world.Block
	dirt   []cube.Pos
}

func newSaplingTreePlan() *saplingTreePlan {
	return &saplingTreePlan{blocks: map[cube.Pos]world.Block{}}
}

func (p *saplingTreePlan) set(pos cube.Pos, b world.Block) { p.blocks[pos] = b }

func (p *saplingTreePlan) setLog(pos cube.Pos, wood WoodType, axis cube.Axis) {
	p.set(pos, Log{Wood: wood, Axis: axis})
}

func (p *saplingTreePlan) setLeaves(pos cube.Pos, leaves LeavesType) {
	if _, ok := p.blocks[pos].(Log); ok {
		return
	}
	p.set(pos, Leaves{Type: leaves})
}

func (p *saplingTreePlan) markDirt(pos cube.Pos) { p.dirt = append(p.dirt, pos) }

func (p *saplingTreePlan) fits(tx *world.Tx) bool {
	for pos, b := range p.blocks {
		if !saplingGrowthReplaceable(tx, pos, b) {
			return false
		}
	}
	return true
}

func (p *saplingTreePlan) apply(tx *world.Tx) {
	for _, pos := range p.dirt {
		if _, ok := tx.Block(pos).(Grass); ok {
			tx.SetBlock(pos, Dirt{}, nil)
		}
	}
	for pos, b := range p.blocks {
		tx.SetBlock(pos, b, nil)
	}
}

func (p *saplingTreePlan) commit(tx *world.Tx) bool {
	if !p.fits(tx) {
		return false
	}
	p.apply(tx)
	return true
}

func (p *saplingTreePlan) trunkColumn(pos cube.Pos, height int, wood WoodType) {
	p.markDirt(pos.Side(cube.FaceDown))
	for y := 0; y < height; y++ {
		p.setLog(pos.Add(cube.Pos{0, y, 0}), wood, cube.Y)
	}
}

func (p *saplingTreePlan) smallTreeLeaves(x, y, z, treeH int, leaves LeavesType, r *rand.Rand) {
	for yy := y - 3 + treeH; yy <= y+treeH; yy++ {
		yOff := yy - (y + treeH)
		mid := 1 - yOff/2
		for xx := x - mid; xx <= x+mid; xx++ {
			for zz := z - mid; zz <= z+mid; zz++ {
				if abs(xx-x) == mid && abs(zz-z) == mid && (yOff == 0 || r.IntN(2) == 0) {
					continue
				}
				p.setLeaves(cube.Pos{xx, yy, zz}, leaves)
			}
		}
	}
}

func (p *saplingTreePlan) spruceLeaves(foot cube.Pos, treeH int, leaves LeavesType, r *rand.Rand) {
	x, y, z := foot[0], foot[1], foot[2]
	topN := treeH - (1 + r.IntN(2))
	lR := 2 + r.IntN(2)
	rad, maxR, minR := r.IntN(2), 1, 0
	for i := 0; i <= topN; i++ {
		ly := y + treeH - i
		for xx := x - rad; xx <= x+rad; xx++ {
			for zz := z - rad; zz <= z+rad; zz++ {
				if abs(xx-x) == rad && abs(zz-z) == rad && rad > 0 {
					continue
				}
				p.setLeaves(cube.Pos{xx, ly, zz}, leaves)
			}
		}
		if rad >= maxR {
			rad, minR = minR, 1
			maxR++
			if maxR > lR {
				maxR = lR
			}
		} else {
			rad++
		}
	}
}

func (p *saplingTreePlan) acaciaBranch(start cube.Pos, dir cube.Direction, maxDiag, n int, wood WoodType) cube.Pos {
	next := start
	d := 0
	for i := 0; i < n; i++ {
		next = next.Side(cube.FaceUp)
		if d < maxDiag {
			next = next.Side(dir.Face())
			d++
		}
		p.setLog(next, wood, cube.Y)
	}
	return next
}

func (p *saplingTreePlan) acaciaLeavesPlate(center cube.Pos, rad, maxTaxicab int, leaves LeavesType) {
	cx, cy, cz := center[0], center[1], center[2]
	for x := cx - rad; x <= cx+rad; x++ {
		for z := cz - rad; z <= cz+rad; z++ {
			if abs(x-cx)+abs(z-cz) <= maxTaxicab {
				p.setLeaves(cube.Pos{x, cy, z}, leaves)
			}
		}
	}
}

func (p *saplingTreePlan) strip2x2ExceptCorner(base cube.Pos) {
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			if dx == 0 && dz == 0 {
				continue
			}
			p.set(base.Add(cube.Pos{dx, 0, dz}), Air{})
		}
	}
}

func saplingGrowthReplaceable(tx *world.Tx, pos cube.Pos, with world.Block) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	if _, ok := tx.Liquid(pos); ok {
		return false
	}
	if saplingTreeReplaceable(tx.Block(pos)) {
		return true
	}
	return replaceableWith(tx, pos, with)
}

func saplingTreeReplaceable(b world.Block) bool {
	switch b.(type) {
	case Air, Sapling, Leaves, ShortGrass, Fern, DoubleTallGrass, Flower, DoubleFlower, PinkPetals, Vines, MossCarpet:
		return true
	}
	return false
}

func saplingTreeGrowthOverwrite(b world.Block) bool {
	if saplingTreeReplaceable(b) {
		return true
	}
	rp, ok := b.(Replaceable)
	return ok && rp.ReplaceableBy(Air{})
}

func saplingTreeCellClear(tx *world.Tx, p cube.Pos) bool {
	if p.OutOfBounds(tx.Range()) {
		return false
	}
	if _, ok := tx.Liquid(p); ok {
		return false
	}
	return saplingTreeGrowthOverwrite(tx.Block(p))
}

func saplingGrowthAllowed(pos cube.Pos, tx *world.Tx) bool {
	return tx.Light(pos) >= 8 || tx.Light(pos.Side(cube.FaceUp)) >= 9
}

func saplingShouldUproot(pos cube.Pos, tx *world.Tx) bool {
	return tx.Light(pos) <= 7 && tx.SkyLight(pos) < 15
}

func saplingSquare(pos cube.Pos, tx *world.Tx, typ SaplingType) (cube.Pos, bool) {
	for _, base := range []cube.Pos{
		pos, pos.Add(cube.Pos{0, 0, -1}), pos.Add(cube.Pos{-1, 0, 0}), pos.Add(cube.Pos{-1, 0, -1}),
	} {
		ok := true
		for dx := 0; dx < 2 && ok; dx++ {
			for dz := 0; dz < 2 && ok; dz++ {
				s, o := tx.Block(base.Add(cube.Pos{dx, 0, dz})).(Sapling)
				ok = o && s.Type == typ
			}
		}
		if ok {
			return base, true
		}
	}
	return cube.Pos{}, false
}

func saplingGrowthBaseValid(pos cube.Pos, tx *world.Tx) bool {
	return supportsVegetation(Sapling{}, tx.Block(pos.Side(cube.FaceDown)))
}

func saplingGrowthSquareValid(base cube.Pos, tx *world.Tx) bool {
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			if !saplingGrowthBaseValid(base.Add(cube.Pos{dx, 0, dz}), tx) {
				return false
			}
		}
	}
	return true
}

func saplingTreeFootprintClear(tx *world.Tx, pos cube.Pos, treeH int) bool {
	x, y, z := pos[0], pos[1], pos[2]
	radius := 0
	for yy := 0; yy < treeH+3; yy++ {
		if yy == 1 || yy == treeH {
			radius++
		}
		for xx := -radius; xx <= radius; xx++ {
			for zz := -radius; zz <= radius; zz++ {
				if !saplingTreeCellClear(tx, cube.Pos{x + xx, y + yy, z + zz}) {
					return false
				}
			}
		}
	}
	return true
}

func saplingTreeRegionClear(tx *world.Tx, min, max cube.Pos) bool {
	for x := min[0]; x <= max[0]; x++ {
		for y := min[1]; y <= max[1]; y++ {
			for z := min[2]; z <= max[2]; z++ {
				if !saplingTreeCellClear(tx, cube.Pos{x, y, z}) {
					return false
				}
			}
		}
	}
	return true
}

func sapling2x2Foot(pos cube.Pos, tx *world.Tx, typ SaplingType) (foot cube.Pos, twoByTwo bool, ok bool) {
	if b, sq := saplingSquare(pos, tx, typ); sq {
		if !saplingGrowthSquareValid(b, tx) {
			return cube.Pos{}, false, false
		}
		return b, true, true
	}
	return pos, false, true
}

func saplingRandBetween(r *rand.Rand, min, maxInclusive int) int {
	return min + r.IntN(maxInclusive-min+1)
}

func growSaplingTree(pos cube.Pos, tx *world.Tx, typ SaplingType, r *rand.Rand) bool {
	switch typ {
	case OakSapling():
		return growColumnTree(pos, tx, OakWood(), OakLeaves(), 4+r.IntN(3), r)
	case SpruceSapling():
		return growSpruceTree(pos, tx, r)
	case BirchSapling():
		h := 5 + r.IntN(3)
		if r.IntN(39) == 0 {
			h += 5
		}
		return growColumnTree(pos, tx, BirchWood(), BirchLeaves(), h, r)
	case JungleSapling():
		return growJungleTree(pos, tx, r)
	case AcaciaSapling():
		return growAcaciaTree(pos, tx, r)
	case DarkOakSapling():
		return growQuadTrunkTree(pos, tx, r, DarkOakSapling(), DarkOakWood(), DarkOakLeaves())
	case CherrySapling():
		return growColumnTree(pos, tx, CherryWood(), CherryLeaves(), 5+r.IntN(4), r)
	case PaleOakSapling():
		return growQuadTrunkTree(pos, tx, r, PaleOakSapling(), PaleOakWood(), PaleOakLeaves())
	default:
		return false
	}
}

func growColumnTree(pos cube.Pos, tx *world.Tx, wood WoodType, leaves LeavesType, treeH int, r *rand.Rand) bool {
	if !saplingTreeFootprintClear(tx, pos, treeH) {
		return false
	}
	p := newSaplingTreePlan()
	p.trunkColumn(pos, treeH-1, wood)
	p.smallTreeLeaves(pos[0], pos[1], pos[2], treeH, leaves, r)
	return p.commit(tx)
}

func growSpruceTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	foot, twoByTwo, ok := sapling2x2Foot(pos, tx, SpruceSapling())
	if !ok {
		return false
	}
	treeH := 6 + r.IntN(4)
	if !saplingTreeFootprintClear(tx, foot, treeH) {
		return false
	}
	p := newSaplingTreePlan()
	if twoByTwo {
		p.strip2x2ExceptCorner(foot)
	}
	p.trunkColumn(foot, treeH-r.IntN(3), SpruceWood())
	p.spruceLeaves(foot, treeH, SpruceLeaves(), r)
	return p.commit(tx)
}

func growJungleTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	const treeH = 8
	foot, twoByTwo, ok := sapling2x2Foot(pos, tx, JungleSapling())
	if !ok {
		return false
	}
	if !saplingTreeFootprintClear(tx, foot, treeH) {
		return false
	}
	p := newSaplingTreePlan()
	if twoByTwo {
		p.strip2x2ExceptCorner(foot)
	}
	p.trunkColumn(foot, treeH-1, JungleWood())
	p.smallTreeLeaves(foot[0], foot[1], foot[2], treeH, JungleLeaves(), r)
	return p.commit(tx)
}

func growAcaciaTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingTreeRegionClear(tx, pos.Add(cube.Pos{-6, 0, -6}), pos.Add(cube.Pos{6, 18, 6})) {
		return false
	}
	rb := saplingRandBetween
	th := 5 + rb(r, 0, 2) + rb(r, 0, 2)
	fbh := th - 1 - rb(r, 0, 3)

	p := newSaplingTreePlan()
	p.markDirt(pos.Side(cube.FaceDown))
	for y := 0; y <= fbh; y++ {
		p.setLog(pos.Add(cube.Pos{0, y, 0}), AcaciaWood(), cube.Y)
	}
	dirs := cube.Directions()
	mf := dirs[r.IntN(len(dirs))]
	main := p.acaciaBranch(pos.Add(cube.Pos{0, fbh, 0}), mf, rb(r, 1, 3), th-fbh, AcaciaWood())

	sf := dirs[r.IntN(len(dirs))]
	var sec cube.Pos
	secOK := sf != mf
	if secOK {
		sec = p.acaciaBranch(pos.Add(cube.Pos{0, fbh - rb(r, 0, 1), 0}), sf, rb(r, 1, 3), rb(r, 1, 3), AcaciaWood())
	}

	p.acaciaLeavesPlate(main, 3, 5, AcaciaLeaves())
	p.acaciaLeavesPlate(main.Side(cube.FaceUp), 2, 2, AcaciaLeaves())
	if secOK {
		p.acaciaLeavesPlate(sec, 2, 3, AcaciaLeaves())
		p.acaciaLeavesPlate(sec.Side(cube.FaceUp), 1, 2, AcaciaLeaves())
	}
	return p.commit(tx)
}

func growQuadTrunkTree(pos cube.Pos, tx *world.Tx, r *rand.Rand, typ SaplingType, wood WoodType, leaves LeavesType) bool {
	base, ok := saplingSquare(pos, tx, typ)
	if !ok || !saplingGrowthSquareValid(base, tx) {
		return false
	}
	treeH := 4 + r.IntN(3)
	if !saplingTreeRegionClear(tx, base.Add(cube.Pos{-5, 0, -5}), base.Add(cube.Pos{6, treeH + 8, 6})) {
		return false
	}
	cx, cz := base[0]+1, base[2]+1
	p := newSaplingTreePlan()
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			p.trunkColumn(base.Add(cube.Pos{dx, 0, dz}), treeH-1, wood)
		}
	}
	p.smallTreeLeaves(cx, base[1], cz, treeH, leaves, r)
	return p.commit(tx)
}
