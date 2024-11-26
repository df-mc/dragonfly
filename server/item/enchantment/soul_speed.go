package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SoulSpeed is an enchantment that can be applied on boots and allows the
// player to walk more quickly on soul sand or soul soil.
var SoulSpeed soulSpeed

type soulSpeed struct{}

// Name ...
func (soulSpeed) Name() string {
	return "Soul Speed"
}

// MaxLevel ...
func (soulSpeed) MaxLevel() int {
	return 3
}

// Cost ...
func (soulSpeed) Cost(level int) (int, int) {
	minCost := level * 10
	return minCost, minCost + 15
}

// Rarity ...
func (soulSpeed) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// Treasure ...
func (soulSpeed) Treasure() bool {
	return true
}

// CompatibleWithEnchantment ...
func (soulSpeed) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (soulSpeed) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
