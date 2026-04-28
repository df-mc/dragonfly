package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type saplingTreePlan struct {
	blocks map[cube.Pos]world.Block
	dirt   []cube.Pos
	podzol []cube.Pos
}

func newSaplingTreePlan() *saplingTreePlan {
	return &saplingTreePlan{blocks: map[cube.Pos]world.Block{}}
}

func (p *saplingTreePlan) set(pos cube.Pos, b world.Block) {
	p.blocks[pos] = b
}

func (p *saplingTreePlan) setLog(pos cube.Pos, wood WoodType, axis cube.Axis) {
	p.set(pos, Log{Wood: wood, Axis: axis})
}

func (p *saplingTreePlan) setLeaves(pos cube.Pos, leaves LeavesType) {
	if _, exists := p.blocks[pos]; exists {
		if _, ok := p.blocks[pos].(Log); ok {
			return
		}
	}
	p.set(pos, Leaves{Type: leaves})
}

func (p *saplingTreePlan) setVines(pos cube.Pos, direction cube.Direction) {
	if existing, ok := p.blocks[pos].(Vines); ok {
		p.blocks[pos] = existing.WithAttachment(direction, true)
		return
	}
	p.blocks[pos] = (Vines{}).WithAttachment(direction, true)
}

func (p *saplingTreePlan) markDirt(pos cube.Pos) {
	p.dirt = append(p.dirt, pos)
}

func (p *saplingTreePlan) markPodzol(pos cube.Pos) {
	p.podzol = append(p.podzol, pos)
}

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
	for _, pos := range p.podzol {
		if _, ok := tx.Block(pos).(Grass); ok {
			tx.SetBlock(pos, Podzol{}, nil)
		}
	}
	for pos, b := range p.blocks {
		tx.SetBlock(pos, b, nil)
	}
}

func saplingGrowthReplaceable(tx *world.Tx, pos cube.Pos, with world.Block) bool {
	if pos.OutOfBounds(tx.Range()) {
		return false
	}
	if _, ok := tx.Liquid(pos); ok {
		return false
	}
	return replaceableWith(tx, pos, with)
}

func saplingGrowthAllowed(pos cube.Pos, tx *world.Tx) bool {
	return tx.Light(pos) >= 8 || tx.Light(pos.Side(cube.FaceUp)) >= 9
}

func saplingShouldUproot(pos cube.Pos, tx *world.Tx) bool {
	return tx.Light(pos) <= 7 && tx.SkyLight(pos) < 15
}

func saplingSquare(pos cube.Pos, tx *world.Tx, typ SaplingType) (cube.Pos, bool) {
	candidates := []cube.Pos{
		pos,
		pos.Add(cube.Pos{0, 0, -1}),
		pos.Add(cube.Pos{-1, 0, 0}),
		pos.Add(cube.Pos{-1, 0, -1}),
	}
	for _, base := range candidates {
		if saplingSquareAt(base, tx, typ) {
			return base, true
		}
	}
	return cube.Pos{}, false
}

func saplingSquareAt(base cube.Pos, tx *world.Tx, typ SaplingType) bool {
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			s, ok := tx.Block(base.Add(cube.Pos{dx, 0, dz})).(Sapling)
			if !ok || s.Type != typ {
				return false
			}
		}
	}
	return true
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

func saplingClearArea(tx *world.Tx, y int, minX, maxX, minZ, maxZ int, allow func(cube.Pos, world.Block) bool) bool {
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			p := cube.Pos{x, y, z}
			b := tx.Block(p)
			if allow != nil && allow(p, b) {
				continue
			}
			if !saplingGrowthReplaceable(tx, p, Sapling{}) {
				return false
			}
		}
	}
	return true
}

func saplingClearVolume(tx *world.Tx, min, max cube.Pos, allow func(cube.Pos, world.Block) bool) bool {
	for x := min[0]; x <= max[0]; x++ {
		for y := min[1]; y <= max[1]; y++ {
			for z := min[2]; z <= max[2]; z++ {
				p := cube.Pos{x, y, z}
				b := tx.Block(p)
				if allow != nil && allow(p, b) {
					continue
				}
				if !saplingGrowthReplaceable(tx, p, Sapling{}) {
					return false
				}
			}
		}
	}
	return true
}

func saplingAllowType(typ SaplingType) func(cube.Pos, world.Block) bool {
	return func(_ cube.Pos, b world.Block) bool {
		s, ok := b.(Sapling)
		return ok && s.Type == typ
	}
}

func (p *saplingTreePlan) addLeafDisc(center cube.Pos, radius int, leaves LeavesType, corners bool) {
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if x*x+z*z > radius*radius+radius/2 {
				continue
			}
			if !corners && abs(x) == radius && abs(z) == radius {
				continue
			}
			p.setLeaves(center.Add(cube.Pos{x, 0, z}), leaves)
		}
	}
}

func (p *saplingTreePlan) addLeafSquare(center cube.Pos, radius int, leaves LeavesType, trimCorners bool) {
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if trimCorners && abs(x) == radius && abs(z) == radius {
				continue
			}
			p.setLeaves(center.Add(cube.Pos{x, 0, z}), leaves)
		}
	}
}

func (p *saplingTreePlan) addLeafBlob(center cube.Pos, radius int, height int, leaves LeavesType) {
	for y := 0; y < height; y++ {
		layerRadius := max(1, radius-abs(y-height/2))
		p.addLeafDisc(center.Add(cube.Pos{0, y, 0}), layerRadius, leaves, false)
	}
}

func (p *saplingTreePlan) addTrunkVines(trunk []cube.Pos, r *rand.Rand) {
	for _, pos := range trunk {
		if r.IntN(4) != 0 {
			continue
		}
		for _, dir := range cube.Directions() {
			if r.IntN(2) == 0 {
				continue
			}
			length := r.IntN(3) + 1
			vinePos := pos.Side(dir.Face())
			for i := 0; i < length; i++ {
				if existing, ok := p.blocks[vinePos]; ok {
					if _, solid := existing.(Log); solid {
						break
					}
				}
				p.setVines(vinePos, dir.Opposite())
				vinePos = vinePos.Side(cube.FaceDown)
			}
		}
	}
}

func directionAxis(dir cube.Direction) cube.Axis {
	switch dir {
	case cube.East, cube.West:
		return cube.X
	default:
		return cube.Z
	}
}

func directionOffset(dir cube.Direction) cube.Pos {
	switch dir {
	case cube.North:
		return cube.Pos{0, 0, -1}
	case cube.South:
		return cube.Pos{0, 0, 1}
	case cube.West:
		return cube.Pos{-1, 0, 0}
	default:
		return cube.Pos{1, 0, 0}
	}
}

func growSaplingTree(pos cube.Pos, tx *world.Tx, typ SaplingType, r *rand.Rand) bool {
	switch typ {
	case OakSapling():
		return growOakTree(pos, tx, r)
	case SpruceSapling():
		return growSpruceTree(pos, tx, r)
	case BirchSapling():
		return growBirchTree(pos, tx, r)
	case JungleSapling():
		return growJungleTree(pos, tx, r)
	case AcaciaSapling():
		return growAcaciaTree(pos, tx, r)
	case DarkOakSapling():
		return growDarkOakTree(pos, tx, r)
	case CherrySapling():
		return growCherryTree(pos, tx, r)
	case PaleOakSapling():
		return growPaleOakTree(pos, tx, r)
	default:
		return false
	}
}

func growOakTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-1, 1, -1}), pos.Add(cube.Pos{1, 4, 1}), nil) {
		return false
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 4, -2}), pos.Add(cube.Pos{2, 6, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(2)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), OakWood(), cube.Y)
	}
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 1, 0}), 2, OakLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height, 0}), 2, OakLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height + 1, 0}), 1, OakLeaves(), true)
	plan.setLeaves(pos.Add(cube.Pos{0, height + 2, 0}), OakLeaves())
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growBirchTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-1, 1, -1}), pos.Add(cube.Pos{1, 3, 1}), nil) {
		return false
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 4, -2}), pos.Add(cube.Pos{2, 7, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(3)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), BirchWood(), cube.Y)
	}
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 2, 0}), 2, BirchLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 1, 0}), 2, BirchLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height, 0}), 2, BirchLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height + 1, 0}), 1, BirchLeaves(), true)
	plan.setLeaves(pos.Add(cube.Pos{0, height + 1, 0}), BirchLeaves())
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growSpruceTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if base, ok := saplingSquare(pos, tx, SpruceSapling()); ok && saplingGrowthSquareValid(base, tx) {
		if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 16, 3}), saplingAllowType(SpruceSapling())) {
			return false
		}
		return growMegaSpruceTree(base, tx, r)
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 1, -2}), pos.Add(cube.Pos{2, 12, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 6 + r.IntN(4)
	var trunk []cube.Pos
	for y := 0; y < height; y++ {
		p := pos.Add(cube.Pos{0, y, 0})
		plan.setLog(p, SpruceWood(), cube.Y)
		trunk = append(trunk, p)
	}
	bare := 2 + r.IntN(2)
	for y := bare; y <= height; y++ {
		d := height - y
		radius := 1
		if d >= 2 {
			radius = 2
		}
		plan.addLeafDisc(pos.Add(cube.Pos{0, y, 0}), radius, SpruceLeaves(), false)
	}
	plan.setLeaves(pos.Add(cube.Pos{0, height + 1, 0}), SpruceLeaves())
	plan.markDirt(pos.Side(cube.FaceDown))
	if r.IntN(6) == 0 {
		plan.addTrunkVines(trunk[:max(0, len(trunk)-2)], r)
	}
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growMegaSpruceTree(base cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 21, 3}), saplingAllowType(SpruceSapling())) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 14 + r.IntN(7)
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				plan.setLog(base.Add(cube.Pos{dx, y, dz}), SpruceWood(), cube.Y)
			}
			plan.markDirt(base.Add(cube.Pos{dx, -1, dz}))
		}
	}
	for y := height - 6; y <= height+1; y++ {
		radius := 2
		if y >= height {
			radius = 1
		}
		for x := -radius; x <= 1+radius; x++ {
			for z := -radius; z <= 1+radius; z++ {
				if abs(x) == radius && abs(z) == radius && y < height {
					continue
				}
				plan.setLeaves(base.Add(cube.Pos{x, y, z}), SpruceLeaves())
			}
		}
	}
	for x := -2; x <= 3; x++ {
		for z := -2; z <= 3; z++ {
			if abs(x-0) <= 1 && abs(z-0) <= 1 {
				continue
			}
			plan.markPodzol(base.Add(cube.Pos{x, -1, z}))
		}
	}
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growJungleTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if base, ok := saplingSquare(pos, tx, JungleSapling()); ok && saplingGrowthSquareValid(base, tx) {
		return growMegaJungleTree(base, tx, r)
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-1, 1, -1}), pos.Add(cube.Pos{1, 3, 1}), nil) {
		return false
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 3, -2}), pos.Add(cube.Pos{2, 7, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(3)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), JungleWood(), cube.Y)
	}
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 1, 0}), 2, JungleLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height, 0}), 2, JungleLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height + 1, 0}), 1, JungleLeaves(), true)
	plan.setLeaves(pos.Add(cube.Pos{0, height + 2, 0}), JungleLeaves())
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growMegaJungleTree(base cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 20, 3}), saplingAllowType(JungleSapling())) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 11 + r.IntN(10)
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				plan.setLog(base.Add(cube.Pos{dx, y, dz}), JungleWood(), cube.Y)
			}
			plan.markDirt(base.Add(cube.Pos{dx, -1, dz}))
		}
	}
	plan.addLeafDisc(base.Add(cube.Pos{0, height - 1, 0}), 3, JungleLeaves(), false)
	plan.addLeafDisc(base.Add(cube.Pos{0, height, 0}), 3, JungleLeaves(), false)
	plan.addLeafDisc(base.Add(cube.Pos{0, height + 1, 0}), 2, JungleLeaves(), false)
	for i := 0; i < 1+r.IntN(4); i++ {
		dir := cube.Directions()[r.IntN(len(cube.Directions()))]
		branchStart := base.Add(cube.Pos{0, height - 3 - i, 0})
		branchEnd := branchStart
		for j := 0; j < 2+r.IntN(3); j++ {
			branchEnd = branchEnd.Side(dir.Face())
			plan.setLog(branchEnd, JungleWood(), directionAxis(dir))
			if j%2 == 1 && branchEnd[1] < base[1]+height+1 {
				branchEnd = branchEnd.Side(cube.FaceUp)
				plan.setLog(branchEnd, JungleWood(), cube.Y)
			}
		}
		plan.addLeafDisc(branchEnd, 2, JungleLeaves(), false)
		plan.addLeafDisc(branchEnd.Side(cube.FaceUp), 1, JungleLeaves(), true)
	}
	trunk := make([]cube.Pos, 0, height*4)
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height-1; y++ {
				trunk = append(trunk, base.Add(cube.Pos{dx, y, dz}))
			}
		}
	}
	plan.addTrunkVines(trunk, r)
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growAcaciaTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-1, 1, -1}), pos.Add(cube.Pos{1, 4, 1}), nil) {
		return false
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 4, -2}), pos.Add(cube.Pos{2, 9, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 7 + r.IntN(3)
	dir := cube.Directions()[r.IntN(len(cube.Directions()))]
	turnAt := 2 + r.IntN(2)
	variant := r.IntN(3)
	for y := 0; y < turnAt; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), AcaciaWood(), cube.Y)
	}
	branchBase := pos.Add(cube.Pos{0, turnAt, 0})
	branchEnd := branchBase
	for i := 0; i < height-turnAt; i++ {
		branchEnd = branchEnd.Side(dir.Face())
		plan.setLog(branchEnd.Add(cube.Pos{0, i, 0}), AcaciaWood(), directionAxis(dir))
	}
	top := branchEnd.Add(cube.Pos{0, height - turnAt, 0})
	plan.setLog(top, AcaciaWood(), cube.Y)
	plan.addLeafSquare(top, 2, AcaciaLeaves(), true)
	plan.addLeafSquare(top.Side(cube.FaceUp), 1, AcaciaLeaves(), false)
	if variant != 0 {
		otherDir := dir.RotateLeft()
		if variant == 2 {
			otherDir = dir.RotateRight()
		}
		otherBase := branchBase.Add(cube.Pos{0, 1 + r.IntN(2), 0})
		otherEnd := otherBase.Side(otherDir.Face())
		plan.setLog(otherEnd, AcaciaWood(), directionAxis(otherDir))
		if variant == 1 {
			plan.setLog(otherEnd.Side(cube.FaceUp), AcaciaWood(), cube.Y)
			plan.addLeafSquare(otherEnd.Side(cube.FaceUp), 2, AcaciaLeaves(), true)
		} else {
			mid := top.Side(cube.FaceUp)
			plan.setLog(mid, AcaciaWood(), cube.Y)
			plan.addLeafSquare(otherEnd, 1, AcaciaLeaves(), false)
			plan.addLeafSquare(mid, 1, AcaciaLeaves(), false)
		}
	}
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growDarkOakTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	base, ok := saplingSquare(pos, tx, DarkOakSapling())
	if !ok || !saplingGrowthSquareValid(base, tx) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{0, 1, 0}), base.Add(cube.Pos{2, 4, 2}), saplingAllowType(DarkOakSapling())) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 4, -1}), base.Add(cube.Pos{3, 9, 3}), saplingAllowType(DarkOakSapling())) {
		return false
	}
	plan := newSaplingTreePlan()
	height := 6 + r.IntN(3)
	bendDir := cube.Directions()[r.IntN(len(cube.Directions()))]
	bend := cube.Pos{}
	if r.IntN(2) == 0 {
		bend = directionOffset(bendDir)
	}
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				offset := cube.Pos{}
				if y >= height-2 {
					offset = bend
				}
				plan.setLog(base.Add(cube.Pos{dx, y, dz}).Add(offset), DarkOakWood(), cube.Y)
			}
			plan.markDirt(base.Add(cube.Pos{dx, -1, dz}))
		}
	}
	crownBase := base.Add(cube.Pos{bend[0], height - 1, bend[2]})
	for y := 0; y < 3; y++ {
		radius := 3 - y
		for x := -radius; x <= 1+radius; x++ {
			for z := -radius; z <= 1+radius; z++ {
				if abs(x) == radius && abs(z) == radius && y != 0 {
					continue
				}
				plan.setLeaves(crownBase.Add(cube.Pos{x, y, z}), DarkOakLeaves())
			}
		}
	}
	branchDir := cube.Directions()[r.IntN(len(cube.Directions()))]
	branchStart := crownBase.Add(cube.Pos{0, 1, 0})
	branchMid := branchStart.Side(branchDir.Face())
	plan.setLog(branchMid, DarkOakWood(), directionAxis(branchDir))
	plan.setLog(branchMid.Side(cube.FaceUp), DarkOakWood(), cube.Y)
	plan.addLeafDisc(branchMid.Side(cube.FaceUp), 2, DarkOakLeaves(), false)
	plan.addLeafDisc(crownBase.Add(cube.Pos{0, 3, 0}), 1, DarkOakLeaves(), true)
	if r.IntN(4) == 0 {
		trunk := make([]cube.Pos, 0, height*4)
		for dx := 0; dx < 2; dx++ {
			for dz := 0; dz < 2; dz++ {
				for y := 0; y < height-1; y++ {
					trunk = append(trunk, base.Add(cube.Pos{dx, y, dz}))
				}
			}
		}
		plan.addTrunkVines(trunk, r)
	}
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growCherryTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 1, -2}), pos.Add(cube.Pos{2, 9, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 7 + r.IntN(3)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), CherryWood(), cube.Y)
	}
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 1, 0}), 3, CherryLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{-1, height, 0}), 3, CherryLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{1, height, 0}), 3, CherryLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height + 1, 0}), 2, CherryLeaves(), false)
	plan.setLeaves(pos.Add(cube.Pos{0, height + 2, 0}), CherryLeaves())
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growPaleOakTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	base, ok := saplingSquare(pos, tx, PaleOakSapling())
	if !ok || !saplingGrowthSquareValid(base, tx) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{0, 1, 0}), base.Add(cube.Pos{2, 4, 2}), saplingAllowType(PaleOakSapling())) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 4, -1}), base.Add(cube.Pos{3, 9, 3}), saplingAllowType(PaleOakSapling())) {
		return false
	}
	plan := newSaplingTreePlan()
	height := 6 + r.IntN(3)
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				plan.setLog(base.Add(cube.Pos{dx, y, dz}), PaleOakWood(), cube.Y)
			}
			plan.markDirt(base.Add(cube.Pos{dx, -1, dz}))
		}
	}
	crownBase := base.Add(cube.Pos{0, height - 1, 0})
	for y := 0; y < 3; y++ {
		radius := 2 + (1 - y/2)
		for x := -radius; x <= 1+radius; x++ {
			for z := -radius; z <= 1+radius; z++ {
				if abs(x) == radius && abs(z) == radius && y > 0 {
					continue
				}
				plan.setLeaves(crownBase.Add(cube.Pos{x, y, z}), PaleOakLeaves())
			}
		}
	}
	branchDir := cube.Directions()[r.IntN(len(cube.Directions()))]
	branch := crownBase.Side(branchDir.Face())
	plan.setLog(branch, PaleOakWood(), directionAxis(branchDir))
	plan.setLeaves(branch.Side(cube.FaceUp), PaleOakLeaves())
	plan.addLeafDisc(crownBase.Add(cube.Pos{0, 3, 0}), 1, PaleOakLeaves(), true)
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}
