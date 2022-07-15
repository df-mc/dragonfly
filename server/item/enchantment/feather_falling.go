package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
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

// Rarity ...
func (FeatherFalling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Affects ...
func (FeatherFalling) Affects(src damage.Source) bool {
	_, fall := src.(damage.SourceFall)
	return fall
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
