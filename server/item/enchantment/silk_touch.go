package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SilkTouch is an enchantment that allows many blocks to drop themselves instead of their usual items when mined.
type SilkTouch struct{}

// Name ...
func (SilkTouch) Name() string {
	return "Silk Touch"
}

// MaxLevel ...
func (SilkTouch) MaxLevel() int {
	return 1
}

// MinCost ...
func (SilkTouch) MinCost(int) int {
	return 15
}

// MaxCost ...
func (SilkTouch) MaxCost(int) int {
	return 65
}

// Rarity ...
func (SilkTouch) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithOther ...
func (SilkTouch) CompatibleWithOther(item.EnchantmentType) bool {
	// TODO: Fortune.
	return true
}

// CompatibleWithItem ...
func (SilkTouch) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() != item.TypeSword && t.ToolType() != item.TypeNone)
}
