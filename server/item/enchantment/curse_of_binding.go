package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// CurseOfBinding is an enchantment that prevents the removal of an equipped item enchanted with the curse of binding from its armour slot.
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
	return 25, 50
}

// Rarity ...
func (CurseOfBinding) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
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

// Treasure ...
func (CurseOfBinding) Treasure() bool {
	return true
}

// Curse ...
func (CurseOfBinding) Curse() bool {
	return true
}
