package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Shears is a tool used to shear sheep, mine a few types of blocks, and carve pumpkins.
type Shears struct {
}

// UseOnBlock ...
func (s Shears) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool {
	if b := w.Block(pos); item_internal.IsUncarvedPumpkin(b) {
		carvedPumpkin := item_internal.CarvePumpkin(b)
		w.SetBlock(pos, carvedPumpkin)

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
