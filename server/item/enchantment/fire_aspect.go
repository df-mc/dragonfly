package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
type FireAspect struct{}

// Name ...
func (FireAspect) Name() string {
	return "Fire Aspect"
}

// MaxLevel ...
func (FireAspect) MaxLevel() int {
	return 2
}

// Rarity ...
func (FireAspect) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Duration returns how long the fire from fire aspect will last.
func (FireAspect) Duration(level int) time.Duration {
	return time.Second * 4 * time.Duration(level)
}

// CompatibleWithEnchantment ...
func (FireAspect) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (FireAspect) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
