package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Sword is a tool generally used to attack enemies. In addition, it may be used to mine any block slightly
// faster than without tool and to break cobwebs rapidly.
type Sword struct {
	// Tier is the tier of the sword.
	Tier tool.Tier
}

// AttackDamage returns the attack damage of the sword.
func (s Sword) AttackDamage() float64 {
	return s.Tier.BaseAttackDamage + 3
}

// MaxCount always returns 1.
func (s Sword) MaxCount() int {
	return 1
}

// ToolType returns the tool type for swords.
func (s Sword) ToolType() tool.Type {
	return tool.TypeSword
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

// EncodeItem ...
func (s Sword) EncodeItem() (id int32, meta int16) {
	switch s.Tier {
	case tool.TierWood:
		return 268, 0
	case tool.TierGold:
		return 283, 0
	case tool.TierStone:
		return 272, 0
	case tool.TierIron:
		return 267, 0
	case tool.TierDiamond:
		return 276, 0
	}
	panic("invalid sword tier")
}
