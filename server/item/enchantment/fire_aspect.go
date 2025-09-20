package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
var FireAspect fireAspect

type fireAspect struct{}

func (fireAspect) Name() string {
	return "Fire Aspect"
}

func (fireAspect) MaxLevel() int {
	return 2
}

func (fireAspect) Cost(level int) (int, int) {
	minCost := 10 + (level-1)*20
	return minCost, minCost + 50
}

func (fireAspect) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Duration returns how long the fire from fire aspect will last.
func (fireAspect) Duration(level int) time.Duration {
	return time.Second * 4 * time.Duration(level)
}

func (fireAspect) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (fireAspect) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
