package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Protection is an armour enchantment which increases the damage reduction.
type Protection struct{}

// Name ...
func (Protection) Name() string {
	return "Protection"
}

// MaxLevel ...
func (Protection) MaxLevel() int {
	return 4
}

// MinCost ...
func (Protection) MinCost(level int) int {
	return 1 + (level-1)*11
}

// MaxCost ...
func (p Protection) MaxCost(level int) int {
	return p.MinCost(level) + 11
}

// Rarity ...
func (Protection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Affects ...
func (Protection) Affects(src damage.Source) bool {
	_, ok := src.(damage.SourceEntityAttack)
	return ok || src == damage.SourceFall{} || src == damage.SourceFire{} || src == damage.SourceFireTick{} || src == damage.SourceLava{}
}

// Multiplier returns the damage multiplier of protection.
func (Protection) Multiplier(lvl int) float64 {
	if lvl > 20 {
		lvl = 20
	}
	return 1 - float64(lvl)/25
}

// CompatibleWithOther ...
func (Protection) CompatibleWithOther(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, fireProt := t.(FireProtection)
	_, prot := t.(ProjectileProtection)
	return !blastProt && !fireProt && !prot
}

// CompatibleWithItem ...
func (Protection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
