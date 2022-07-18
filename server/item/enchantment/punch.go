package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Punch increases the knock-back dealt when hitting a player or mob with a bow.
type Punch struct{}

// Name ...
func (Punch) Name() string {
	return "Punch"
}

// MaxLevel ...
func (Punch) MaxLevel() int {
	return 2
}

// Cost ...
func (Punch) Cost(level int) (int, int) {
	min := 12 + (level-1)*20
	return min, min + 25
}

// Rarity ...
func (Punch) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// PunchMultiplier returns the punch multiplier for the level and horizontal speed.
func (Punch) PunchMultiplier(level int, horizontalSpeed float64) float64 {
	return float64(level) * 1.25 / horizontalSpeed
}

// CompatibleWithEnchantment ...
func (Punch) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (Punch) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
