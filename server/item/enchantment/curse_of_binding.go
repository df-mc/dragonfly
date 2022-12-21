package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Curse of Binding is an enchantment that prevents removal of a cursed item from its armour slot.
type CurseOfBinding struct{}

// Name ...
func (CurseOfBinding) Name() string {
	return "Curse of Binding"
}

// MaxLevel ...
func (CurseOfBinding) MaxLevel() int {
	return 1
}

// Cost ...
func (CurseOfBinding) Cost(level int) (int, int) {
	return 0, 0
}

// Rarity ...
func (CurseOfBinding) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// CompatibleWithEnchantment ...
func (CurseOfBinding) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (CurseOfBinding) CompatibleWithItem(i world.Item) bool {
	_, isArmour := i.(item.Armour)

	return isArmour
}
