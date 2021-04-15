package item

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Shears is a tool used to shear sheep, mine a few types of blocks, and carve pumpkins.
type Shears struct{}

// UseOnBlock ...
func (s Shears) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if face == cube.FaceUp || face == cube.FaceDown {
		// Pumpkins can only be carved when once of the horizontal faces is clicked.
		return false
	}
	if b := w.Block(pos); item_internal.IsUncarvedPumpkin(b) {
		// TODO: Drop pumpkin seeds.
		carvedPumpkin := item_internal.CarvePumpkin(b, face)
		w.PlaceBlock(pos, carvedPumpkin)

		ctx.DamageItem(1)
		return true
	}
	return false
}

// ToolType ...
func (s Shears) ToolType() tool.Type {
	return tool.TypeShears
}

// HarvestLevel ...
func (s Shears) HarvestLevel() int {
	return 1
}

// BaseMiningEfficiency ...
func (s Shears) BaseMiningEfficiency(b world.Block) float64 {
	return 1.5
}

// DurabilityInfo ...
func (s Shears) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    238,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 0,
		BreakDurability:  1,
	}
}

// MaxCount ...
func (s Shears) MaxCount() int {
	return 1
}

// EncodeItem ...
func (s Shears) EncodeItem() (id int32, meta int16) {
	return 359, 0
}
