package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// CurseOfVanishing is an enchantment that causes the item to disappear on death.
type CurseOfVanishing struct{}

// Name ...
func (CurseOfVanishing) Name() string {
	return "Curse of Vanishing"
}

// MaxLevel ...
func (CurseOfVanishing) MaxLevel() int {
	return 1
}

// Cost ...
func (CurseOfVanishing) Cost(int) (int, int) {
	return 25, 50
}

// Rarity ...
func (CurseOfVanishing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (CurseOfVanishing) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (CurseOfVanishing) CompatibleWithItem(i world.Item) bool {
	_, arm := i.(item.Armour)
	_, com := i.(item.Compass)
	_, dur := i.(item.Durable)
	_, rec := i.(item.RecoveryCompass)
	// TODO: Carrot on a Stick
	// TODO: Warped Fungus on a Stick
	return arm || com || dur || rec
}

// Treasure ...
func (CurseOfVanishing) Treasure() bool {
	return true
}

// Curse ...
func (CurseOfVanishing) Curse() bool {
	return true
}
