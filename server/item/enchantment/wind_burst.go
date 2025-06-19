package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// WindBurst is enchantment for mace.
var WindBurst windBurst

type windBurst struct{}

// Name ...
func (windBurst) Name() string {
	return "Wind Burst"
}

// MaxLevel ...
func (windBurst) MaxLevel() int {
	return 3
}

// Cost ...
func (windBurst) Cost(level int) (int, int) {
	minCost := 1 + (level-1)*11
	return minCost, minCost + 20
}

// Rarity ...
func (windBurst) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (windBurst) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (windBurst) CompatibleWithItem(i world.Item) bool {
	encodeStr, _ := i.EncodeItem()
	return encodeStr == "minecraft:wind_charge"
}
