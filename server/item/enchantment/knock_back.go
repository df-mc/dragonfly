package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// KnockBack is an enchantment to a sword that increases the sword's knock-back.
type KnockBack struct{}

// Force returns the increase in knock-back force from the enchantment.
func (e KnockBack) Force(lvl int) float64 {
	return float64(lvl) / 2
}

// Name ...
func (e KnockBack) Name() string {
	return "Knockback"
}

// MaxLevel ...
func (e KnockBack) MaxLevel() int {
	return 2
}

// Rarity ...
func (e KnockBack) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (e KnockBack) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e KnockBack) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
