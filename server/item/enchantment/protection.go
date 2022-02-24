package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
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

// CompatibleWith ...
func (e BlastProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)

	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})

	return ok && !fireProt && !projectileProt && !prot
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

// CompatibleWith ...
func (e FireProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})

	return ok && !blastProt && !projectileProt && !prot
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

// CompatibleWith ...
func (e ProjectileProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, prot := s.Enchantment(Protection{})

	return ok && !blastProt && !fireProt && !prot
}

// Protection is an armour enchantment which increases the damage reduction.
type Protection struct{}

// Affects ...
func (e Protection) Affects(src damage.Source) bool {
	return src == damage.SourceEntityAttack{} || src == damage.SourceFall{} || src == damage.SourceFire{} || src == damage.SourceFireTick{} || src == damage.SourceLava{}
}

// Subtrahend returns the amount of damage that should be subtracted with protection.
func (e Protection) Subtrahend(level int) float64 {
	return float64(level) / 20
}

// Name ...
func (e Protection) Name() string {
	return "Protection"
}

// MaxLevel ...
func (e Protection) MaxLevel() int {
	return 4
}

// CompatibleWith ...
func (e Protection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})

	return ok && !blastProt && !fireProt && !projectileProt
}
