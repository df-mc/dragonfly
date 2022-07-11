package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Sword is a tool generally used to attack enemies. In addition, it may be used to mine any block slightly
// faster than without tool and to break cobwebs rapidly.
type Sword struct {
	// Tier is the tier of the sword.
	Tier ToolTier
}

// AttackDamage returns the attack damage to the sword.
func (s Sword) AttackDamage() float64 {
	return s.Tier.BaseAttackDamage + 3
}

// MaxCount always returns 1.
func (s Sword) MaxCount() int {
	return 1
}

// ToolType returns the tool type for swords.
func (s Sword) ToolType() ToolType {
	return TypeSword
}

// HarvestLevel returns the harvest level of the sword tier.
func (s Sword) HarvestLevel() int {
	return s.Tier.HarvestLevel
}

// BaseMiningEfficiency always returns 1.5, unless the block passed is cobweb, in which case 15 is returned.
func (s Sword) BaseMiningEfficiency(world.Block) float64 {
	// TODO: Implement cobwebs and return 15 here.
	return 1.5
}

// DurabilityInfo ...
func (s Sword) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    s.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 1,
		BreakDurability:  2,
	}
}

// RepairableBy ...
func (s Sword) RepairableBy(i Stack) bool {
	return toolTierRepairable(s.Tier)(i)
}

// EncodeItem ...
func (s Sword) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Tier.Name + "_sword", 0
}
