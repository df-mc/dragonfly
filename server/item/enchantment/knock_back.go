package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// KnockBack is an enchantment to a sword that increases the sword's knock-back.
type KnockBack struct{}

// Name ...
func (KnockBack) Name() string {
	return "Knockback"
}

// MaxLevel ...
func (KnockBack) MaxLevel() int {
	return 2
}

// Rarity ...
func (KnockBack) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Force returns the increase in knock-back force from the enchantment.
func (KnockBack) Force(level int) float64 {
	return float64(level) / 2
}

// CompatibleWithEnchantment ...
func (KnockBack) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (KnockBack) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
