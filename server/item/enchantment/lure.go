package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Lure is an enchantment for fishing rods that decreases the time it takes for a fish to bite the hook.
var Lure lure

type lure struct{}

// Name ...
func (lure) Name() string {
	return "Lure"
}

// MaxLevel ...
func (lure) MaxLevel() int {
	return 3
}

// Cost ...
func (lure) Cost(level int) (int, int) {
	minCost := 15 + (level-1)*9
	return minCost, minCost + 50
}

// Rarity ...
func (lure) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// LureLevel returns the level of lure applied to a fishing hook cast with a fishing rod with this
// enchantment. Every level reduces the time until a fish bites by 5 seconds.
func (lure) LureLevel(level int) int {
	return level
}

// CompatibleWithEnchantment ...
func (lure) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (lure) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.FishingRod)
	return ok
}
