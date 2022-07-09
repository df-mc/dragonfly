package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
type FireAspect struct{}

// Duration returns how long the fire from fire aspect will last.
func (FireAspect) Duration(lvl int) time.Duration {
	return time.Second * 4 * time.Duration(lvl)
}

// Name ...
func (FireAspect) Name() string {
	return "Fire Aspect"
}

// MaxLevel ...
func (FireAspect) MaxLevel() int {
	return 2
}

// MinCost ...
func (FireAspect) MinCost(level int) int {
	return 10 + (level-1)*20
}

// MaxCost ...
func (f FireAspect) MaxCost(level int) int {
	return f.MinCost(level) + 50
}

// Rarity ...
func (FireAspect) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithOther ...
func (FireAspect) CompatibleWithOther(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (FireAspect) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
