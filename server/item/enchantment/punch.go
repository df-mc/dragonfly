package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Punch increases the knock-back dealt when hitting a player or mob with a bow.
var Punch punch

type punch struct{}

func (punch) Name() string {
	return "Punch"
}

func (punch) MaxLevel() int {
	return 2
}

func (punch) Cost(level int) (int, int) {
	minCost := 12 + (level-1)*20
	return minCost, minCost + 25
}

func (punch) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// KnockBackMultiplier returns the punch multiplier for the level and horizontal speed.
func (punch) KnockBackMultiplier() float64 {
	return 0.25
}

func (punch) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (punch) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
