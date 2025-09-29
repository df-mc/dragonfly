package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// CurseOfVanishing is an enchantment that causes the item to disappear on
// death.
var CurseOfVanishing curseOfVanishing

type curseOfVanishing struct{}

func (curseOfVanishing) Name() string {
	return "Curse of Vanishing"
}

func (curseOfVanishing) MaxLevel() int {
	return 1
}

func (curseOfVanishing) Cost(int) (int, int) {
	return 25, 50
}

func (curseOfVanishing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

func (curseOfVanishing) CompatibleWithEnchantment(_ item.EnchantmentType) bool {
	return true
}

func (curseOfVanishing) CompatibleWithItem(i world.Item) bool {
	_, arm := i.(item.Armour)
	_, com := i.(item.Compass)
	_, dur := i.(item.Durable)
	_, rec := i.(item.RecoveryCompass)
	// TODO: Carrot on a Stick
	// TODO: Warped Fungus on a Stick
	return arm || com || dur || rec
}

func (curseOfVanishing) Treasure() bool {
	return true
}

func (curseOfVanishing) Curse() bool {
	return true
}
