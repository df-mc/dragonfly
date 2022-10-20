package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Name ...
func (FeatherFalling) Name() string {
	return "Feather Falling"
}

// MaxLevel ...
func (FeatherFalling) MaxLevel() int {
	return 4
}

// Cost ...
func (FeatherFalling) Cost(level int) (int, int) {
	min := 5 + (level-1)*6
	return min, min + 6
}

// Rarity ...
func (FeatherFalling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Modifier returns the base protection modifier for the enchantment.
func (FeatherFalling) Modifier() float64 {
	return 2.5
}

// CompatibleWithEnchantment ...
func (FeatherFalling) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (FeatherFalling) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
