package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
type FireAspect struct{}

// Duration returns how long the fire from fire aspect will last.
func (e FireAspect) Duration(lvl int) time.Duration {
	return time.Second * 4 * time.Duration(lvl)
}

// Name ...
func (e FireAspect) Name() string {
	return "Fire Aspect"
}

// MaxLevel ...
func (e FireAspect) MaxLevel() int {
	return 2
}

// Rarity ...
func (e FireAspect) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (e FireAspect) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e FireAspect) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
