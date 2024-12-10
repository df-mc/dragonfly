package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

var QuickCharge quickCharge

type quickCharge struct{}

// Name ...
func (quickCharge) Name() string {
	return "Quick Charge"
}

// MaxLevel ...
func (quickCharge) MaxLevel() int {
	return 3
}

// Cost ...
func (quickCharge) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*10
	return minCost, minCost + 20
}

// Rarity ...
func (quickCharge) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

func (quickCharge) ChargeDuration(level int) time.Duration {
	return time.Duration(1.25-0.25*float64(level)) * time.Second
}

// CompatibleWithEnchantment ...
func (quickCharge) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (quickCharge) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Crossbow)
	return ok
}
