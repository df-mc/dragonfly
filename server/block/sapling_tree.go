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
		switch tx.Block(pos).(type) {
		case Grass, Dirt, Mud, MuddyMangroveRoots:
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

func saplingBaseClear(tx *world.Tx, base cube.Pos, typ SaplingType) bool {
	return saplingSameLevelClear(tx, base, func(_ cube.Pos, b world.Block) bool {
		s, ok := b.(Sapling)
		return ok && s.Type == typ
	})
}

func jungleBaseClear(tx *world.Tx, base cube.Pos) bool {
	return saplingSameLevelClear(tx, base, func(pos cube.Pos, b world.Block) bool {
		if s, ok := b.(Sapling); ok && s.Type == JungleSapling() {
			return true
		}
		if pos[0] <= base[0] || pos[2] <= base[2] {
			switch b.(type) {
			case Log, Wood, Leaves:
				return true
			}
		}
		return false
	})
}

func saplingSameLevelClear(tx *world.Tx, base cube.Pos, allow func(cube.Pos, world.Block) bool) bool {
	for x := base[0] - 1; x <= base[0]+1; x++ {
		for z := base[2] - 1; z <= base[2]+1; z++ {
			pos := cube.Pos{x, base[1], z}
			if _, ok := tx.Liquid(pos); ok {
				return false
			}
			b := tx.Block(pos)
			if allow != nil && allow(pos, b) {
				continue
			}
			if _, ok := b.(Air); !ok {
				return false
			}
		}
	}
	return true
}

func (p *saplingTreePlan) addLeafPlus(center cube.Pos, leaves LeavesType) {
	p.setLeaves(center, leaves)
	for _, face := range cube.HorizontalFaces() {
		p.setLeaves(center.Side(face), leaves)
	}
}

func (p *saplingTreePlan) addSmallTreeCanopy(pos cube.Pos, trunkHeight int, leaves LeavesType, r *rand.Rand) {
	top := pos.Add(cube.Pos{0, trunkHeight, 0})
	p.addLeafPlus(top, leaves)

	second := pos.Add(cube.Pos{0, trunkHeight - 1, 0})
	p.addLeafPlus(second, leaves)
	corners := []cube.Pos{{-1, 0, -1}, {-1, 0, 1}, {1, 0, -1}, {1, 0, 1}}
	cornerCount := 1 + r.IntN(3)
	for i, c := range r.Perm(len(corners)) {
		if i >= cornerCount {
			break
		}
		p.setLeaves(second.Add(corners[c]), leaves)
	}

	for y := trunkHeight - 2; y >= trunkHeight-3; y-- {
		layer := pos.Add(cube.Pos{0, y, 0})
		p.addLeafSquare(layer, 2, leaves, true)
		for _, c := range corners {
			if r.IntN(4) == 0 {
				p.setLeaves(layer.Add(cube.Pos{c[0] * 2, c[1] * 2, c[2] * 2}), leaves)
			}
		}
	}
}

func (p *saplingTreePlan) addSpruceCanopy(pos cube.Pos, height int, leaves LeavesType, r *rand.Rand) {
	// Bedrock spruce crowns are layered and conical, with discrete rows.
	top := pos.Add(cube.Pos{0, height, 0})
	p.setLeaves(top, leaves)
	p.addLeafDisc(top.Side(cube.FaceDown), 1, leaves, false)
	p.addLeafDisc(top.Side(cube.FaceDown).Side(cube.FaceDown), 2, leaves, true)

	rowRadius := 1 + r.IntN(2)
	for y := height - 3; y >= 2+r.IntN(2); y-- {
		p.addLeafDisc(pos.Add(cube.Pos{0, y, 0}), rowRadius, leaves, true)
		if rowRadius == 1 {
			rowRadius = 2
			continue
		}
		if r.IntN(3) != 0 {
			rowRadius = 1
		}
	}
}

func (p *saplingTreePlan) addMegaSpruceCanopy(base cube.Pos, height int, leaves LeavesType, r *rand.Rand) {
	// 2x2 spruce in Bedrock has a broad lower crown and tighter top layers.
	for y := height - 9; y <= height+1; y++ {
		d := height - y
		radius := 1
		switch {
		case d >= 7:
			radius = 3
		case d >= 4:
			radius = 2
		case d >= 1:
			radius = 2
		}
		for x := -radius; x <= 1+radius; x++ {
			for z := -radius; z <= 1+radius; z++ {
				edgeX, edgeZ := abs(x) == radius, abs(z) == radius
				if edgeX && edgeZ {
					// Keep lower layers rougher and the top tighter.
					skipChance := 2
					if d <= 2 {
						skipChance = 3
					}
					if r.IntN(skipChance) == 0 {
						continue
					}
				}
				p.setLeaves(base.Add(cube.Pos{x, y, z}), leaves)
			}
		}
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
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 3, -2}), pos.Add(cube.Pos{2, 7, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(2)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), OakWood(), cube.Y)
	}
	plan.addSmallTreeCanopy(pos, height, OakLeaves(), r)
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
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 4, -2}), pos.Add(cube.Pos{2, 8, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(3)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), BirchWood(), cube.Y)
	}
	plan.addSmallTreeCanopy(pos, height, BirchLeaves(), r)
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growSpruceTree(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if base, ok := saplingSquare(pos, tx, SpruceSapling()); ok {
		if !saplingGrowthSquareValid(base, tx) || !saplingBaseClear(tx, base, SpruceSapling()) {
			return false
		}
		if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 30, 3}), saplingAllowType(SpruceSapling())) {
			return false
		}
		return growMegaSpruceTree(base, tx, r)
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 1, -2}), pos.Add(cube.Pos{2, 12, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 6 + r.IntN(4)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), SpruceWood(), cube.Y)
	}
	plan.addSpruceCanopy(pos, height, SpruceLeaves(), r)
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growMegaSpruceTree(base cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !saplingBaseClear(tx, base, SpruceSapling()) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 30, 3}), saplingAllowType(SpruceSapling())) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 14 + r.IntN(17)
	for dx := 0; dx < 2; dx++ {
		for dz := 0; dz < 2; dz++ {
			for y := 0; y < height; y++ {
				plan.setLog(base.Add(cube.Pos{dx, y, dz}), SpruceWood(), cube.Y)
			}
			plan.markDirt(base.Add(cube.Pos{dx, -1, dz}))
		}
	}
	plan.addMegaSpruceCanopy(base, height, SpruceLeaves(), r)
	for x := -6; x <= 7; x++ {
		for z := -6; z <= 7; z++ {
			if x*x+z*z > 49 {
				continue
			}
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
	if base, ok := saplingSquare(pos, tx, JungleSapling()); ok {
		if !saplingGrowthSquareValid(base, tx) || !jungleBaseClear(tx, base) {
			return false
		}
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
	plan.addSmallTreeCanopy(pos, height, JungleLeaves(), r)
	plan.markDirt(pos.Side(cube.FaceDown))
	if !plan.fits(tx) {
		return false
	}
	plan.apply(tx)
	return true
}

func growMegaJungleTree(base cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if !jungleBaseClear(tx, base) {
		return false
	}
	if !saplingClearVolume(tx, base.Add(cube.Pos{-1, 1, -1}), base.Add(cube.Pos{3, 32, 3}), saplingAllowType(JungleSapling())) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 11 + r.IntN(21)
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
	for i := 0; i < 1+r.IntN(6); i++ {
		dir := cube.Directions()[r.IntN(len(cube.Directions()))]
		branchStart := base.Add(cube.Pos{0, height - 3 - i, 0})
		branchEnd := branchStart
		for j := 0; j < 1+r.IntN(6); j++ {
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
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-1, 1, -1}), pos.Add(cube.Pos{1, 3, 1}), nil) {
		return false
	}
	if !saplingClearVolume(tx, pos.Add(cube.Pos{-2, 3, -2}), pos.Add(cube.Pos{2, 10, 2}), nil) {
		return false
	}

	plan := newSaplingTreePlan()
	height := 5 + r.IntN(4)
	dir := cube.Directions()[r.IntN(len(cube.Directions()))]
	leanStart := 2 + r.IntN(2)
	leanSteps := 1 + r.IntN(2)

	trunk := pos
	for y := 0; y < height; y++ {
		if y >= leanStart && y < leanStart+leanSteps {
			trunk = trunk.Side(dir.Face())
		}
		plan.setLog(cube.Pos{trunk[0], pos[1] + y, trunk[2]}, AcaciaWood(), cube.Y)
	}

	mainTop := cube.Pos{trunk[0], pos[1] + height, trunk[2]}
	plan.setLog(mainTop, AcaciaWood(), cube.Y)
	plan.addLeafSquare(mainTop, 2, AcaciaLeaves(), true)
	plan.addLeafSquare(mainTop.Side(cube.FaceUp), 1, AcaciaLeaves(), false)
	for _, side := range cube.HorizontalFaces() {
		if r.IntN(3) == 0 {
			plan.setLeaves(mainTop.Side(side).Side(cube.FaceUp), AcaciaLeaves())
		}
	}

	// Bedrock acacia can occasionally fork into a second canopy.
	if r.IntN(2) == 0 {
		secondDir := dir.RotateLeft()
		if r.IntN(2) == 0 {
			secondDir = dir.RotateRight()
		}
		branchStartY := max(2, height-2-r.IntN(2))
		second := cube.Pos{pos[0], pos[1] + branchStartY, pos[2]}
		length := 1 + r.IntN(2)
		for i := 0; i < length; i++ {
			second = second.Side(secondDir.Face())
			axis := directionAxis(secondDir)
			if i == length-1 {
				axis = cube.Y
			}
			plan.setLog(second, AcaciaWood(), axis)
		}
		plan.addLeafSquare(second, 2, AcaciaLeaves(), true)
		plan.addLeafSquare(second.Side(cube.FaceUp), 1, AcaciaLeaves(), false)
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
	height := 5 + r.IntN(4)
	branchStart := 2 + r.IntN(3)
	for y := 0; y < height; y++ {
		plan.setLog(pos.Add(cube.Pos{0, y, 0}), CherryWood(), cube.Y)
	}
	for _, dir := range cube.Directions() {
		if r.IntN(4) == 0 {
			continue
		}
		branch := pos.Add(cube.Pos{0, branchStart + r.IntN(max(1, height-branchStart)), 0})
		length := 2 + r.IntN(3)
		for i := 0; i < length; i++ {
			branch = branch.Side(dir.Face())
			plan.setLog(branch, CherryWood(), directionAxis(dir))
			if i == length-1 || r.IntN(3) == 0 {
				plan.addLeafDisc(branch, 3, CherryLeaves(), false)
				plan.addLeafDisc(branch.Side(cube.FaceUp), 2, CherryLeaves(), false)
				if r.IntN(2) == 0 {
					plan.addLeafDisc(branch.Side(cube.FaceDown), 2, CherryLeaves(), false)
				}
			}
		}
	}
	plan.addLeafDisc(pos.Add(cube.Pos{0, height - 1, 0}), 4, CherryLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height, 0}), 4, CherryLeaves(), false)
	plan.addLeafDisc(pos.Add(cube.Pos{0, height + 1, 0}), 2, CherryLeaves(), false)
	for x := -3; x <= 3; x++ {
		for z := -3; z <= 3; z++ {
			if abs(x)+abs(z) < 3 || r.IntN(3) != 0 {
				continue
			}
			plan.setLeaves(pos.Add(cube.Pos{x, height - 2, z}), CherryLeaves())
		}
	}
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
