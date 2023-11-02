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

// KnockBackMultiplier returns the punch multiplier for the level and horizontal speed.
func (Punch) KnockBackMultiplier() float64 {
	return 0.25
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
