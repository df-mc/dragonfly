package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// Bamboo is a non-solid plant block that can be placed on vegetation-supporting blocks.
type Bamboo struct {
	empty
	transparent
	Age      bool
	LeafSize int
	Thick    bool
}

var _ item.BoneMealAffected = Bamboo{}

// UseOnBlock ...
func (b Bamboo) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Bamboo); ok {
		// Vanilla only allows extending a manually placed stalk up to two blocks.
		base := bambooBase(below, tx)
		if bambooHeightFromBase(base, tx) >= 2 {
			return false
		}
	} else if !supportsVegetation(b, tx.Block(below)) {
		return false
	}
	b.Age = false
	b.LeafSize = SmallLeaves
	b.Thick = false
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...

// BoneMeal grows a bamboo stalk by 1-2 blocks if there is enough room.
func (b Bamboo) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	top, ok := bambooTop(pos, tx)
	if !ok {
		return false
	}
	growth := rand.IntN(2) + 1
	applied := false
	for range growth {
		nextTop, ok := growBamboo(top, tx)
		if !ok {
			break
		}
		top = nextTop
		applied = true
	}
	return applied
}

// NeighbourUpdateTick ...
func (b Bamboo) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
}

// RandomTick ...
func (b Bamboo) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		return
	}
	if tx.Light(pos) < 9 || r.IntN(3) != 0 {
		return
	}
	top, ok := bambooTop(pos, tx)
	if !ok {
		return
	}
	_, _ = growBamboo(top, tx)
}

// HasLiquidDrops ...
func (Bamboo) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (Bamboo) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (b Bamboo) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       oneOf(b),
		BreakHandler: func(pos cube.Pos, tx *world.Tx, u item.User) {
			tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
		},
	}
}

// CompostChance ...
func (Bamboo) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (Bamboo) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo", 0
}

// EncodeBlock ...
func (b Bamboo) EncodeBlock() (string, map[string]any) {
	thickness := "thin"
	if b.Thick {
		thickness = "thick"
	}
	return "minecraft:bamboo", map[string]any{
		"age_bit":                boolByte(b.Age),
		"bamboo_leaf_size":       bambooLeafSizeString(b.LeafSize),
		"bamboo_stalk_thickness": thickness,
	}
}

// allBamboo returns all bamboo block states.
func allBamboo() (blocks []world.Block) {
	for _, age := range []bool{false, true} {
		for _, leafSize := range bambooLeafSizes() {
			for _, thick := range []bool{false, true} {
				blocks = append(blocks, Bamboo{Age: age, LeafSize: leafSize, Thick: thick})
			}
		}
	}
	return
}

const (
	bambooNoLeaves = iota
	SmallLeaves
	LargeLeaves
)

func bambooLeafSizes() []int {
	return []int{bambooNoLeaves, SmallLeaves, LargeLeaves}
}

func bambooLeafSizeString(size int) string {
	switch size {
	case SmallLeaves:
		return "small_leaves"
	case LargeLeaves:
		return "large_leaves"
	default:
		return "no_leaves"
	}
}

func canSurviveBamboo(pos cube.Pos, tx *world.Tx) bool {
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Bamboo); ok {
		return canSurviveBamboo(below, tx)
	}
	return supportsVegetation(Bamboo{}, tx.Block(below))
}

func bambooTop(pos cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	if _, ok := tx.Block(pos).(Bamboo); !ok {
		return cube.Pos{}, false
	}
	for {
		next := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(next).(Bamboo); !ok {
			return pos, true
		}
		pos = next
	}
}

func bambooBase(pos cube.Pos, tx *world.Tx) cube.Pos {
	for {
		next := pos.Side(cube.FaceDown)
		if _, ok := tx.Block(next).(Bamboo); !ok {
			return pos
		}
		pos = next
	}
}

func bambooHeightFromBase(base cube.Pos, tx *world.Tx) int {
	height := 1
	for curr := base.Side(cube.FaceUp); ; curr = curr.Side(cube.FaceUp) {
		if _, ok := tx.Block(curr).(Bamboo); !ok {
			return height
		}
		height++
	}
}

func growBamboo(top cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	above := top.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return cube.Pos{}, false
	}
	base := bambooBase(top, tx)
	if bambooHeightFromBase(base, tx) >= 16 {
		return cube.Pos{}, false
	}

	topBlock, _ := tx.Block(top).(Bamboo)
	newTop := Bamboo{Age: topBlock.Age, LeafSize: LargeLeaves, Thick: true}
	tx.SetBlock(above, newTop, nil)
	refreshBambooTop(above, tx)
	return above, true
}

// refreshBambooTop updates the top section so the last three blocks carry leaves.
func refreshBambooTop(top cube.Pos, tx *world.Tx) {
	if b, ok := tx.Block(top).(Bamboo); ok {
		b.LeafSize = LargeLeaves
		b.Thick = true
		tx.SetBlock(top, b, nil)
	}
	if b, ok := tx.Block(top.Side(cube.FaceDown)).(Bamboo); ok {
		b.LeafSize = LargeLeaves
		b.Thick = true
		tx.SetBlock(top.Side(cube.FaceDown), b, nil)
	}
	if b, ok := tx.Block(top.Side(cube.FaceDown).Side(cube.FaceDown)).(Bamboo); ok {
		b.LeafSize = SmallLeaves
		b.Thick = true
		tx.SetBlock(top.Side(cube.FaceDown).Side(cube.FaceDown), b, nil)
	}
	if b, ok := tx.Block(top.Side(cube.FaceDown).Side(cube.FaceDown).Side(cube.FaceDown)).(Bamboo); ok {
		b.LeafSize = bambooNoLeaves
		b.Thick = true
		tx.SetBlock(top.Side(cube.FaceDown).Side(cube.FaceDown).Side(cube.FaceDown), b, nil)
	}
}
