package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Pickaxe is a tool generally used for mining stone-like blocks and ores at a higher speed and to obtain
// their drops.
type Pickaxe struct {
	// Tier is the tier of the pickaxe.
	Tier ToolTier
}

// ToolType returns the type for pickaxes.
func (p Pickaxe) ToolType() ToolType {
	return TypePickaxe
}

// HarvestLevel returns the level that this pickaxe is able to harvest. If a block has a harvest level above
// this one, this pickaxe won't be able to harvest it.
func (p Pickaxe) HarvestLevel() int {
	return p.Tier.HarvestLevel
}

// BaseMiningEfficiency is the base efficiency of the pickaxe, when it comes to mining blocks. This decides
// the speed with which blocks can be mined.
func (p Pickaxe) BaseMiningEfficiency(world.Block) float64 {
	return p.Tier.BaseMiningEfficiency
}

// MaxCount returns 1.
func (p Pickaxe) MaxCount() int {
	return 1
}

// AttackDamage returns the attack damage to the pickaxe.
func (p Pickaxe) AttackDamage() float64 {
	return p.Tier.BaseAttackDamage + 1
}

// DurabilityInfo ...
func (p Pickaxe) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    p.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// RepairableBy ...
func (p Pickaxe) RepairableBy(i Stack) bool {
	return toolTierRepairable(p.Tier)(i)
}

// EncodeItem ...
func (p Pickaxe) EncodeItem() (name string, meta int16) {
	return "minecraft:" + p.Tier.Name + "_pickaxe", 0
}
