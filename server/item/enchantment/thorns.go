package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Thorns is an enchantment that inflicts damage on attackers.
var Thorns thorns

type thorns struct{}

func (thorns) Name() string {
	return "Thorns"
}

func (thorns) MaxLevel() int {
	return 3
}

func (thorns) Cost(level int) (int, int) {
	minCost := 10 + 20*(level-1)
	return minCost, minCost + 50
}

func (thorns) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

func (thorns) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (thorns) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}

// ThornsDamageSource is used for damage caused by thorns.
type ThornsDamageSource struct {
	// Owner is the owner of the armour with the thorns enchantment.
	Owner world.Entity
}

func (ThornsDamageSource) ReducedByResistance() bool { return true }
func (ThornsDamageSource) ReducedByArmour() bool     { return false }
func (ThornsDamageSource) Fire() bool                { return false }
func (ThornsDamageSource) IgnoreTotem() bool         { return false }
