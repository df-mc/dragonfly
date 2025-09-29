package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SoulSpeed is an enchantment that can be applied on boots and allows the
// player to walk more quickly on soul sand or soul soil.
var SoulSpeed soulSpeed

type soulSpeed struct{}

func (soulSpeed) Name() string {
	return "Soul Speed"
}

func (soulSpeed) MaxLevel() int {
	return 3
}

func (soulSpeed) Cost(level int) (int, int) {
	minCost := level * 10
	return minCost, minCost + 15
}

func (soulSpeed) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

func (soulSpeed) Treasure() bool {
	return true
}

func (soulSpeed) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (soulSpeed) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
