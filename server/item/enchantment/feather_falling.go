package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Multiplier returns the damage multiplier of feather falling.
func (FeatherFalling) Multiplier(lvl int) float64 {
	return 1 - 0.12*float64(lvl)
}

// Name ...
func (FeatherFalling) Name() string {
	return "Feather Falling"
}

// Rarity ...
func (FeatherFalling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// MaxLevel ...
func (FeatherFalling) MaxLevel() int {
	return 4
}

// MinCost ...
func (FeatherFalling) MinCost(level int) int {
	return 5 + (level-1)*6
}

// MaxCost ...
func (f FeatherFalling) MaxCost(level int) int {
	return f.MinCost(level) + 10
}

// CompatibleWithOther ...
func (FeatherFalling) CompatibleWithOther(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (FeatherFalling) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
