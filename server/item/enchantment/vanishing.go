package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// CurseOfVanishing is an enchantment that causes the item to disappear on
// death.
var CurseOfVanishing curseOfVanishing

type curseOfVanishing struct{}

// Name ...
func (curseOfVanishing) Name() string {
	return "Curse of Vanishing"
}

// MaxLevel ...
func (curseOfVanishing) MaxLevel() int {
	return 1
}

// Cost ...
func (curseOfVanishing) Cost(int) (int, int) {
	return 25, 50
}

// Rarity ...
func (curseOfVanishing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (curseOfVanishing) CompatibleWithEnchantment(_ item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (curseOfVanishing) CompatibleWithItem(i world.Item) bool {
	_, arm := i.(item.Armour)
	_, com := i.(item.Compass)
	_, dur := i.(item.Durable)
	_, rec := i.(item.RecoveryCompass)
	// TODO: Carrot on a Stick
	// TODO: Warped Fungus on a Stick
	return arm || com || dur || rec
}

// Treasure ...
func (curseOfVanishing) Treasure() bool {
	return true
}

// Curse ...
func (curseOfVanishing) Curse() bool {
	return true
}
