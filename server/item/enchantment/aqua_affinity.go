package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// AquaAffinity is a helmet enchantment that increases underwater mining speed.
type AquaAffinity struct{}

// Name ...
func (e AquaAffinity) Name() string {
	return "Aqua Affinity"
}

// MaxLevel ...
func (e AquaAffinity) MaxLevel() int {
	return 1
}

// Rarity ...
func (e AquaAffinity) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithOther ...
func (e AquaAffinity) CompatibleWithOther(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e AquaAffinity) CompatibleWithItem(i world.Item) bool {
	h, ok := i.(item.HelmetType)
	return ok && h.Helmet()
}
