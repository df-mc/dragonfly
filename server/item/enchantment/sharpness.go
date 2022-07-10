package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Sharpness is an enchantment applied to a sword or axe that increases melee damage.
type Sharpness struct{}

// Addend returns the additional damage when attacking with sharpness.
func (e Sharpness) Addend(level int) float64 {
	return float64(level) * 1.25
}

// Name ...
func (e Sharpness) Name() string {
	return "Sharpness"
}

// MaxLevel ...
func (e Sharpness) MaxLevel() int {
	return 5
}

// Rarity ...
func (e Sharpness) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// CompatibleWithEnchantment ...
func (e Sharpness) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e Sharpness) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() == item.TypeSword || t.ToolType() == item.TypeAxe)
}
