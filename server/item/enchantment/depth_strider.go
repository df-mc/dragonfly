package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// DepthStrider is a boot enchantment that increases underwater movement speed.
var DepthStrider depthStrider

type depthStrider struct{}

func (depthStrider) Name() string {
	return "Depth Strider"
}

func (depthStrider) MaxLevel() int {
	return 3
}

func (depthStrider) Cost(level int) (int, int) {
	minCost := level * 10
	return minCost, minCost + 15
}

func (depthStrider) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

func (depthStrider) CompatibleWithEnchantment(item.EnchantmentType) bool {
	// TODO: Frost Walker
	return true
}

func (depthStrider) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
