package item

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Hoe is a tool generally used to till dirt and grass blocks into farmland blocks for planting crops.
// Additionally a Hoe can be used to break certain types of blocks such as Crimson and Hay Blocks.
type Hoe struct {
	Tier tool.Tier
}

// UseOnBlock will turn a dirt or grass block into a farmland if the necessary properties are met.
func (h Hoe) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool {
	if b, ok := w.Block(pos).(tillable); ok {
		if res, ok := b.Till(); ok {
			if face == cube.FaceDown {
				// Tilled land isn't created when the bottom face is clicked.
				return false
			}
			if w.Block(pos.Add(cube.Pos{0, 1})) != item_internal.Air {
				// Tilled land can only be created if air is above the grass block.
				return false
			}
			w.PlaceBlock(pos, res)
			w.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: res})
			ctx.DamageItem(1)
			return true
		}
	}
	return false
}

// tillable represents a block that can be tilled by using a hoe on it.
type tillable interface {
	// Till returns a block that results from tilling it. If tilling it does not have a result, the bool returned
	// is false.
	Till() (world.Block, bool)
}

// MaxCount ...
func (h Hoe) MaxCount() int {
	return 1
}

// AttackDamage ...
func (h Hoe) AttackDamage() float64 {
	return h.Tier.BaseAttackDamage + 1
}

// ToolType ...
func (h Hoe) ToolType() tool.Type {
	return tool.TypeHoe
}

// BaseMiningEfficiency ...
func (h Hoe) BaseMiningEfficiency(world.Block) float64 {
	return h.Tier.BaseMiningEfficiency
}

// DurabilityInfo ...
func (h Hoe) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    h.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// EncodeItem ...
func (h Hoe) EncodeItem() (id int32, name string, meta int16) {
	name = "minecraft:" + h.Tier.Name + "_hoe"
	switch h.Tier {
	case tool.TierWood:
		return 290, name, 0
	case tool.TierGold:
		return 294, name, 0
	case tool.TierStone:
		return 291, name, 0
	case tool.TierIron:
		return 292, name, 0
	case tool.TierDiamond:
		return 293, name, 0
	case tool.TierNetherite:
		return 747, name, 0
	}
	panic("invalid hoe tier")
}
