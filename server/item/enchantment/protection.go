package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// BlastProtection is an armour enchantment that decreases explosion damage.
type BlastProtection struct{}

// Name ...
func (e BlastProtection) Name() string {
	return "Blast Protection"
}

// MaxLevel ...
func (e BlastProtection) MaxLevel() int {
	return 4
}

// Rarity ...
func (e BlastProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (e BlastProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, fireProt := t.(FireProtection)
	_, projectileProt := t.(ProjectileProtection)
	_, prot := t.(Protection)
	return !fireProt && !projectileProt && !prot
}

// CompatibleWithItem ...
func (e BlastProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}

// FireProtection is an armour enchantment that decreases fire damage.
type FireProtection struct{}

// Name ...
func (e FireProtection) Name() string {
	return "Fire Protection"
}

// MaxLevel ...
func (e FireProtection) MaxLevel() int {
	return 4
}

// Rarity ...
func (e FireProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (e FireProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, projectileProt := t.(ProjectileProtection)
	_, prot := t.(Protection)
	return !blastProt && !projectileProt && !prot
}

// CompatibleWithItem ...
func (e FireProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}

// ProjectileProtection is an armour enchantment that reduces damage from projectiles.
type ProjectileProtection struct{}

// Name ...
func (e ProjectileProtection) Name() string {
	return "Projectile Protection"
}

// MaxLevel ...
func (e ProjectileProtection) MaxLevel() int {
	return 4
}

// Rarity ...
func (e ProjectileProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (e ProjectileProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, fireProt := t.(FireProtection)
	_, prot := t.(Protection)
	return !blastProt && !fireProt && !prot
}

// CompatibleWithItem ...
func (e ProjectileProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}

// Protection is an armour enchantment which increases the damage reduction.
type Protection struct{}

// Affects ...
func (e Protection) Affects(src damage.Source) bool {
	_, ok := src.(damage.SourceEntityAttack)
	return ok || src == damage.SourceFall{} || src == damage.SourceFire{} || src == damage.SourceFireTick{} || src == damage.SourceLava{}
}

// Multiplier returns the damage multiplier of protection.
func (e Protection) Multiplier(lvl int) float64 {
	if lvl > 20 {
		lvl = 20
	}
	return 1 - float64(lvl)/25
}

// Name ...
func (e Protection) Name() string {
	return "Protection"
}

// MaxLevel ...
func (e Protection) MaxLevel() int {
	return 4
}

// Rarity ...
func (e Protection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// CompatibleWithEnchantment ...
func (e Protection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, fireProt := t.(FireProtection)
	_, prot := t.(ProjectileProtection)
	return !blastProt && !fireProt && !prot
}

// CompatibleWithItem ...
func (e Protection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
