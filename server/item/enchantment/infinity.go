package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Infinity is an enchantment to bows that prevents regular arrows from being
// consumed when shot.
var Infinity infinity

type infinity struct{}

func (infinity) Name() string {
	return "Infinity"
}

func (infinity) MaxLevel() int {
	return 1
}

func (infinity) Cost(int) (int, int) {
	return 20, 50
}

func (infinity) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// ConsumesArrows always returns false.
func (infinity) ConsumesArrows() bool {
	return false
}

func (infinity) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Mending
}

func (infinity) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
