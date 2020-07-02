package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Hoe is a tool generally used to till dirt and grass blocks into FarmLand blocks for planting crops.
// Additionally a Hoe can be used to break certain types of blocks such as Crimson and Hay Blocks.
type Hoe struct {
	Tier tool.Tier
}

// UseOnBlock will turn a dirt or grass block into a farmland if the necessary properties are met.
func (h Hoe) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool {
	if grass := w.Block(pos); grass == item_internal.Grass || grass == item_internal.Dirt {
		if face == world.FaceDown {
			// Tilled land isn't created when the bottom face is clicked.
			return false
		}
		if w.Block(pos.Add(world.BlockPos{0, 1})) != item_internal.Air {
			// Tilled land can only be created if air is above the grass block.
			return false
		}
		w.SetBlock(pos, item_internal.FarmLand)
		w.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: item_internal.FarmLand})
		ctx.DamageItem(1)
		return true
	}
	return false
}

// MaxCount will always return 1.
func (h Hoe) MaxCount() int {
	return 1
}

// AttackDamage returns the damage that the hoe will do when attacking.
func (h Hoe) AttackDamage() float64 {
	return h.Tier.BaseAttackDamage + 1
}

// ToolType returns the Hoe ToolType.
func (h Hoe) ToolType() tool.Type {
	return tool.TypeHoe
}

// BaseMiningEfficiency returns how fast the hoe will mine it's respective blocks.
func (h Hoe) BaseMiningEfficiency(world.Block) float64 {
	return h.Tier.BaseMiningEfficiency
}

// DurabilityInfo returns the durability of the Hoe after breaking or attacking, along with the max durability of that tier.
func (h Hoe) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    h.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// EncodeItem will return the ID of the Hoe based on it's tier.
func (h Hoe) EncodeItem() (id int32, meta int16) {
	switch h.Tier {
	case tool.TierWood:
		return 290, 0
	case tool.TierGold:
		return 294, 0
	case tool.TierStone:
		return 291, 0
	case tool.TierIron:
		return 292, 0
	case tool.TierDiamond:
		return 293, 0
	case tool.TierNetherite:
		return 747, 0
	}
	panic("invalid hoe tier")
}
