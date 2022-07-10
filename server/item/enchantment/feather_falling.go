package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Multiplier returns the damage multiplier of feather falling.
func (e FeatherFalling) Multiplier(lvl int) float64 {
	return 1 - 0.12*float64(lvl)
}

// Name ...
func (e FeatherFalling) Name() string {
	return "Feather Falling"
}

// Rarity ...
func (e FeatherFalling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// MaxLevel ...
func (e FeatherFalling) MaxLevel() int {
	return 4
}

// CompatibleWithEnchantment ...
func (e FeatherFalling) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e FeatherFalling) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
