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

// Cost ...
func (Protection) Cost(level int) (int, int) {
	min := 1 + (level-1)*11
	return min, min + 11
}

// Rarity ...
func (Protection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Affects ...
func (Protection) Affects(src damage.Source) bool {
	_, projectile := src.(damage.SourceProjectile)
	_, attack := src.(damage.SourceEntityAttack)
	_, fall := src.(damage.SourceFall)
	return projectile || attack || fall || src.Fire()
}

// Modifier returns the base protection modifier for the enchantment.
func (Protection) Modifier() float64 {
	return 0.75
}

// CompatibleWithEnchantment ...
func (Protection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProtection := t.(BlastProtection)
	_, fireProtection := t.(FireProtection)
	_, projectileProtection := t.(ProjectileProtection)
	return !blastProtection && !fireProtection && !projectileProtection
}

// CompatibleWithItem ...
func (Protection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
