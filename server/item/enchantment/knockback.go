package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Knockback is an enchantment to a sword that increases the sword's knock-back.
var Knockback knockback

type knockback struct{}

func (knockback) Name() string {
	return "Knockback"
}

func (knockback) MaxLevel() int {
	return 2
}

func (knockback) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*20
	return minCost, minCost + 50
}

func (knockback) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Force returns the increase in knock-back force from the enchantment.
func (knockback) Force(level int) float64 {
	return float64(level) / 2
}

func (knockback) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (knockback) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
