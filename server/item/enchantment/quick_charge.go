package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// QuickCharge is an enchantment for quickly reloading a crossbow.
var QuickCharge quickCharge

type quickCharge struct{}

func (quickCharge) Name() string {
	return "Quick Charge"
}

func (quickCharge) MaxLevel() int {
	return 3
}

func (quickCharge) Cost(level int) (int, int) {
	minCost := 12 + (level-1)*20
	return minCost, 50
}

func (quickCharge) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// ChargeDuration returns the charge duration.
func (quickCharge) ChargeDuration(level int) time.Duration {
	return time.Duration((1.25 - 0.25*float64(level)) * float64(time.Second))
}

func (quickCharge) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (quickCharge) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Crossbow)
	return ok
}
