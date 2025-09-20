package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Shears is a tool used to shear sheep, mine a few types of blocks, and carve pumpkins.
type Shears struct{}

func (s Shears) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	if face == cube.FaceUp || face == cube.FaceDown {
		// Pumpkins can only be carved when one of the horizontal faces is clicked.
		return false
	}
	if c, ok := tx.Block(pos).(carvable); ok {
		if res, ok := c.Carve(face); ok {
			// TODO: Drop pumpkin seeds.
			tx.SetBlock(pos, res, nil)

			ctx.DamageItem(1)
			return true
		}
	}
	return false
}

// carvable represents a block that may be carved by using shears on it.
type carvable interface {
	// Carve returns the resulting block of carving this block. If carving it has no result, Carve returns false.
	Carve(f cube.Face) (world.Block, bool)
}

func (s Shears) ToolType() ToolType {
	return TypeShears
}

func (s Shears) HarvestLevel() int {
	return 1
}

func (s Shears) BaseMiningEfficiency(world.Block) float64 {
	return 1.5
}

func (s Shears) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    238,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 0,
		BreakDurability:  1,
	}
}

func (s Shears) MaxCount() int {
	return 1
}

func (s Shears) EncodeItem() (name string, meta int16) {
	return "minecraft:shears", 0
}
