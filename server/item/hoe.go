package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Hoe is a tool generally used to till dirt and grass blocks into farmland blocks for planting crops.
// Additionally, a Hoe can be used to break certain types of blocks such as Crimson and Hay Blocks.
type Hoe struct {
	Tier ToolTier
}

// UseOnBlock will turn a dirt or grass block into a farmland if the necessary properties are met.
func (h Hoe) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	if b, ok := tx.Block(pos).(tillable); ok {
		if res, ok := b.Till(); ok {
			if face == cube.FaceDown {
				// Tilled land isn't created when the bottom face is clicked.
				return false
			}
			if tx.Block(pos.Side(cube.FaceUp)) != air() {
				// Tilled land can only be created if air is above the grass block.
				return false
			}
			tx.SetBlock(pos, res, nil)
			tx.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: res})
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
func (h Hoe) ToolType() ToolType {
	return TypeHoe
}

// HarvestLevel returns the level that this hoe is able to harvest. If a block has a harvest level above
// this one, this hoe won't be able to harvest it.
func (h Hoe) HarvestLevel() int {
	return h.Tier.HarvestLevel
}

// BaseMiningEfficiency ...
func (h Hoe) BaseMiningEfficiency(world.Block) float64 {
	return h.Tier.BaseMiningEfficiency
}

// EnchantmentValue ...
func (h Hoe) EnchantmentValue() int {
	return h.Tier.EnchantmentValue
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

// SmeltInfo ...
func (h Hoe) SmeltInfo() SmeltInfo {
	switch h.Tier {
	case ToolTierIron:
		return newOreSmeltInfo(NewStack(IronNugget{}, 1), 0.1)
	case ToolTierGold:
		return newOreSmeltInfo(NewStack(GoldNugget{}, 1), 0.1)
	case ToolTierCopper:
		return newOreSmeltInfo(NewStack(CopperNugget{}, 1), 0.1)
	}
	return SmeltInfo{}
}

// FuelInfo ...
func (h Hoe) FuelInfo() FuelInfo {
	if h.Tier == ToolTierWood {
		return newFuelInfo(time.Second * 10)
	}
	return FuelInfo{}
}

// RepairableBy ...
func (h Hoe) RepairableBy(i Stack) bool {
	return toolTierRepairable(h.Tier)(i)
}

// EncodeItem ...
func (h Hoe) EncodeItem() (name string, meta int16) {
	return "minecraft:" + h.Tier.Name + "_hoe", 0
}
