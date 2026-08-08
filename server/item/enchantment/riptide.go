package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Riptide is a trident enchantment that launches its user when the trident is
// thrown while the user is in water or rain, instead of throwing the trident.
var Riptide riptide

type riptide struct{}

// Name ...
func (riptide) Name() string {
	return "Riptide"
}

// MaxLevel ...
func (riptide) MaxLevel() int {
	return 3
}

// Cost ...
func (riptide) Cost(level int) (int, int) {
	return 10 + level*7, 50
}

// Rarity ...
func (riptide) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// RiptideForce returns the force with which the user is launched when
// releasing a riptide trident.
func (riptide) RiptideForce(level int) float64 {
	return 3 * float64(1+level) / 4
}

// CompatibleWithEnchantment ...
func (riptide) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Loyalty && t != Channeling
}

// CompatibleWithItem ...
func (riptide) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Trident)
	return ok
}
