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

// UseOnBlock places bamboo. When placed on soil it starts as a thin Bamboo Shoot (no leaves).
// When placed on top of an existing bamboo stalk it extends it and updates the whole stalk shape.
func (b Bamboo) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Bamboo); !ok {
		if _, ok := tx.Block(below).(BambooSapling); !ok && !supportsVegetation(b, tx.Block(below)) {
			return false
		}
	}

	if _, ok := tx.Block(below).(Bamboo); ok {
		// Extending an existing stalk: the new top block is aged and leafy.
		b.Age = true
		b.LeafSize = LargeLeaves
		b.Thick = true
		place(tx, pos, b, user, ctx)
		base := bambooBase(pos, tx)
		updateBambooStalk(base, tx)
	} else if _, ok := tx.Block(below).(BambooSapling); ok {
		// Placing on top of a sapling: convert sapling to bottom bamboo and extend stalk.
		b.Age = true
		b.LeafSize = LargeLeaves
		b.Thick = true
		place(tx, pos, b, user, ctx)
		// Convert the sapling below to a proper bamboo bottom block.
		tx.SetBlock(below, Bamboo{Age: false, LeafSize: bambooNoLeaves, Thick: false}, nil)
		updateBambooStalk(below, tx)
	} else {
		// Planting a new shoot on the ground: use BambooSapling.
		sapling := BambooSapling{Age: false}
		place(tx, pos, sapling, user, ctx)
	}
	return placed(ctx)
}

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

// NeighbourUpdateTick breaks the bamboo if it loses support.
func (b Bamboo) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
}

// RandomTick handles natural bamboo growth.
// In vanilla, only the top block (age_bit=1) can grow further with a 1/3 chance.
func (b Bamboo) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		return
	}
	if tx.Light(pos) < 9 {
		return
	}
	// Only the top block (age=1) can grow further.
	if !b.Age {
		return
	}
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return
	}
	if r.IntN(3) != 0 {
		return
	}
	_, _ = growBamboo(pos, tx)
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
	if bambooHeightFromBase(base, tx) >= 12+rand.IntN(5) {
		return cube.Pos{}, false
	}

	// If the base is still a sapling, convert it to bamboo first.
	// The base becomes the bottom block (age=0, no leaves).
	if _, ok := tx.Block(base).(BambooSapling); ok {
		tx.SetBlock(base, Bamboo{Age: false, LeafSize: bambooNoLeaves, Thick: false}, nil)
	}

	// Grow a new top block (age=1, small leaves).
	tx.SetBlock(above, Bamboo{Age: true, LeafSize: SmallLeaves, Thick: false}, nil)
	updateBambooStalk(base, tx)
	return above, true
}

// updateBambooStalk updates the entire bamboo stalk to match vanilla Bedrock visuals.
// Height 1   = thin shoot, no leaves.
// Height 2-3 = thin stalk, top 2 blocks have large leaves.
// Height >=4 = thick stalk, top 2 blocks have large leaves, 3rd-from-top has small leaves.
func updateBambooStalk(base cube.Pos, tx *world.Tx) {
	height := bambooHeightFromBase(base, tx)
	if height == 0 {
		return
	}

	for i := 0; i < height; i++ {
		pos := base.Add(cube.Pos{0, i, 0})
		b, ok := tx.Block(pos).(Bamboo)
		if !ok {
			continue
		}

		// Thickness: thin while short, thick once mature (height >= 4).
		b.Thick = height >= 4

		// Leaf distribution based on distance from the top.
		distFromTop := height - 1 - i
		switch height {
		case 1:
			b.LeafSize = bambooNoLeaves
		case 2:
			if distFromTop == 0 {
				b.LeafSize = LargeLeaves
			} else {
				b.LeafSize = bambooNoLeaves
			}
		case 3, 4:
			if distFromTop <= 1 {
				b.LeafSize = LargeLeaves
			} else {
				b.LeafSize = bambooNoLeaves
			}
		default: // height >= 5
			switch {
			case distFromTop <= 1:
				b.LeafSize = LargeLeaves
			case distFromTop == 2:
				b.LeafSize = SmallLeaves
			default:
				b.LeafSize = bambooNoLeaves
			}
		}
		tx.SetBlock(pos, b, nil)
	}
}
