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

// Rarity ...
func (Protection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Affects ...
func (Protection) Affects(src damage.Source) bool {
	_, projectile := src.(damage.SourceProjectile)
	_, attack := src.(damage.SourceEntityAttack)
	_, fireTick := src.(damage.SourceFireTick)
	_, fall := src.(damage.SourceFall)
	_, fire := src.(damage.SourceFire)
	_, lava := src.(damage.SourceLava)
	return projectile || attack || fireTick || fall || fire || lava
}

// Modifier returns the base protection modifier for the enchantment.
func (Protection) Modifier() float64 {
	return 0.75
}

// CompatibleWithEnchantment ...
func (Protection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	// TODO: Ensure that the armour does not have blast protection.
	_, fireProt := t.(FireProtection)
	_, projectileProt := t.(ProjectileProtection)
	return !fireProt && !projectileProt
}

// CompatibleWithItem ...
func (Protection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
