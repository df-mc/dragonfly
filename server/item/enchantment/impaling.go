package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Impaling is a trident enchantment that increases damage dealt to entities
// that are in contact with water or rain.
var Impaling impaling

type impaling struct{}

// Name ...
func (impaling) Name() string {
	return "Impaling"
}

// MaxLevel ...
func (impaling) MaxLevel() int {
	return 5
}

// Cost ...
func (impaling) Cost(level int) (int, int) {
	minCost := 1 + (level-1)*8
	return minCost, minCost + 20
}

// Rarity ...
func (impaling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Addend is the extra amount of damage the Impaling enchantment adds when
// attacking mobs that are touching water
func (impaling) Addend(level int) float64 {
	return float64(level) * 2.5
}

// CompatibleWithEnchantment ...
func (impaling) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (impaling) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Trident)
	return ok
}
