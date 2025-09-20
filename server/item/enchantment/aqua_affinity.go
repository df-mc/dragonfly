package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// AquaAffinity is a helmet enchantment that increases underwater mining speed.
var AquaAffinity aquaAffinity

type aquaAffinity struct{}

func (aquaAffinity) Name() string {
	return "Aqua Affinity"
}

func (aquaAffinity) MaxLevel() int {
	return 1
}

func (aquaAffinity) Cost(int) (int, int) {
	return 1, 41
}

func (aquaAffinity) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

func (aquaAffinity) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (aquaAffinity) CompatibleWithItem(i world.Item) bool {
	h, ok := i.(item.HelmetType)
	return ok && h.Helmet()
}
