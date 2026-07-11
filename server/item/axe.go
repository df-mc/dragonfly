package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Axe is a tool generally used for mining wood-like blocks. It may also be used to break some plant-like
// blocks at a faster pace such as pumpkins.
type Axe struct {
	// Tier is the tier of the axe.
	Tier ToolTier
}

// UseOnBlock handles the stripping of logs when a player clicks a log with an axe.
func (a Axe) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	if s, ok := tx.Block(pos).(Strippable); ok {
		if res, so, ok := s.Strip(); ok {
			tx.SetBlock(pos, res, nil)
			tx.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: res})
			if so != nil {
				tx.PlaySound(pos.Vec3(), so)
			}

			ctx.DamageItem(1)
			return true
		}
	}
	return false
}

// Strippable represents a block that can be stripped by right-clicking it with
// an axe.
type Strippable interface {
	// Strip returns a block that is the result of stripping it. Alternatively,
	// the bool returned may be false to indicate the block couldn't be
	// stripped.
	Strip() (world.Block, world.Sound, bool)
}

// MaxCount always returns 1.
func (a Axe) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (a Axe) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    a.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// SmeltInfo ...
func (a Axe) SmeltInfo() SmeltInfo {
	switch a.Tier {
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
func (a Axe) FuelInfo() FuelInfo {
	if a.Tier == ToolTierWood {
		return newFuelInfo(time.Second * 10)
	}
	return FuelInfo{}
}

// AttackDamage ...
func (a Axe) AttackDamage() float64 {
	return a.Tier.BaseAttackDamage + 2
}

// ToolType ...
func (a Axe) ToolType() ToolType {
	return TypeAxe
}

// HarvestLevel ...
func (a Axe) HarvestLevel() int {
	return a.Tier.HarvestLevel
}

// BaseMiningEfficiency ...
func (a Axe) BaseMiningEfficiency(world.Block) float64 {
	return a.Tier.BaseMiningEfficiency
}

// RepairableBy ...
func (a Axe) RepairableBy(i Stack) bool {
	return toolTierRepairable(a.Tier)(i)
}

// EnchantmentValue ...
func (a Axe) EnchantmentValue() int {
	return a.Tier.EnchantmentValue
}

// EncodeItem ...
func (a Axe) EncodeItem() (name string, meta int16) {
	return "minecraft:" + a.Tier.Name + "_axe", 0
}
