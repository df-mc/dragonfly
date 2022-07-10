package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// KnockBack is an enchantment to a sword that increases the sword's knock-back.
type KnockBack struct{}

// Force returns the increase in knock-back force from the enchantment.
func (KnockBack) Force(lvl int) float64 {
	return float64(lvl) / 2
}

// Name ...
func (KnockBack) Name() string {
	return "Knockback"
}

// MaxLevel ...
func (KnockBack) MaxLevel() int {
	return 2
}

// MinCost ...
func (KnockBack) MinCost(level int) int {
	return 5 + (level-1)*20
}

// MaxCost ...
func (k KnockBack) MaxCost(level int) int {
	return k.MinCost(level) + 50
}

// Rarity ...
func (KnockBack) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
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
