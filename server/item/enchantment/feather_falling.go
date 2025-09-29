package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does
// not affect falling speed.
var FeatherFalling featherFalling

type featherFalling struct{}

func (featherFalling) Name() string {
	return "Feather Falling"
}

func (featherFalling) MaxLevel() int {
	return 4
}

func (featherFalling) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*6
	return minCost, minCost + 6
}

func (featherFalling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Modifier returns the base protection modifier for the enchantment.
func (featherFalling) Modifier() float64 {
	return 0.12
}

func (featherFalling) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (featherFalling) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.BootsType)
	return ok && b.Boots()
}
