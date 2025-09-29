package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Power is a bow enchantment which increases arrow damage.
var Power power

type power struct{}

func (power) Name() string {
	return "Power"
}

func (power) MaxLevel() int {
	return 5
}

func (power) Cost(level int) (int, int) {
	minCost := 1 + (level-1)*10
	return minCost, minCost + 15
}

func (power) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// PowerDamage returns the extra base damage dealt by the enchantment and level.
func (power) PowerDamage(level int) float64 {
	return float64(level+1) * 0.5
}

func (power) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (power) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
