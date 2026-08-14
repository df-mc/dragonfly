package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Piercing is an enchantment that allows arrows to damage and pierce through multiple entities, including shields.
var Piercing piercing

type piercing struct{}

func (p piercing) Name() string {
	return "Piercing"
}

func (p piercing) MaxLevel() int {
	return 4
}

func (p piercing) Cost(level int) (int, int) {
	return 1 + (level-1)*10, 50
}

func (p piercing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

func (p piercing) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Multishot
}

func (p piercing) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Crossbow)
	return ok
}

func (p piercing) Pierces() bool {
	return true
}
