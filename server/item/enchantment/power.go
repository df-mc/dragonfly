package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Power is a bow enchantment which increases arrow damage.
type Power struct{}

// Name ...
func (Power) Name() string {
	return "Power"
}

// MaxLevel ...
func (Power) MaxLevel() int {
	return 5
}

// Cost ...
func (Power) Cost(level int) (int, int) {
	min := 1 + (level-1)*10
	return min, min + 15
}

// Rarity ...
func (Power) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// PowerDamage returns the extra base damage dealt by the enchantment and level.
func (Power) PowerDamage(level int) float64 {
	return float64(level+1) * 0.5
}

// CompatibleWithEnchantment ...
func (Power) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (Power) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
