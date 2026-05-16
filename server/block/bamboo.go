package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bamboo is a non-solid plant block that can be placed on vegetation-supporting blocks.
type Bamboo struct {
	empty
	transparent

	// Age specifies the bamboo age bit block state.
	Age bool
	// LeafSize specifies the leaf size block state.
	LeafSize int
	// Thick specifies the bamboo stalk thickness block state.
	Thick bool
}

// UseOnBlock ...
func (b Bamboo) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used || !supportsVegetation(b, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, b, user, ctx)
	return placed(ctx)
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
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(b))
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
