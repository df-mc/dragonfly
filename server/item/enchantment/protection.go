package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/armour"
)

// BlastProtection is an armour enchantment that decreases explosion damage.
type BlastProtection struct {
	enchantment
}

// Name ...
func (e BlastProtection) Name() string {
	return "Blast Protection"
}

// MaxLevel ...
func (e BlastProtection) MaxLevel() int {
	return 4
}

// WithLevel ...
func (e BlastProtection) WithLevel(level int) item.Enchantment {
	return BlastProtection{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e BlastProtection) CompatibleWith(s item.Stack) bool {
	_, armour := s.Item().(armour.Armour)

	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})

	return armour && !fireProt && !projectileProt && !prot
}

// FireProtection is an armour enchantment that decreases fire damage.
type FireProtection struct {
	enchantment
}

// Name ...
func (e FireProtection) Name() string {
	return "Fire Protection"
}

// MaxLevel ...
func (e FireProtection) MaxLevel() int {
	return 4
}

// WithLevel ...
func (e FireProtection) WithLevel(level int) item.Enchantment {
	return FireProtection{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e FireProtection) CompatibleWith(s item.Stack) bool {
	_, armour := s.Item().(armour.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})

	return armour && !blastProt && !projectileProt && !prot
}

// ProjectileProtection is an armour enchantment that reduces damage from projectiles.
type ProjectileProtection struct {
	enchantment
}

// Name ...
func (e ProjectileProtection) Name() string {
	return "Projectile Protection"
}

// MaxLevel ...
func (e ProjectileProtection) MaxLevel() int {
	return 4
}

// WithLevel ...
func (e ProjectileProtection) WithLevel(level int) item.Enchantment {
	return ProjectileProtection{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e ProjectileProtection) CompatibleWith(s item.Stack) bool {
	_, armour := s.Item().(armour.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, prot := s.Enchantment(Protection{})

	return armour && !blastProt && !fireProt && !prot
}

// Protection is an armour enchantment which increases the damage reduction.
type Protection struct {
	enchantment
}

// Name ...
func (e Protection) Name() string {
	return "Protection"
}

// MaxLevel ...
func (e Protection) MaxLevel() int {
	return 4
}

// WithLevel ...
func (e Protection) WithLevel(level int) item.Enchantment {
	return Protection{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e Protection) CompatibleWith(s item.Stack) bool {
	_, armour := s.Item().(armour.Armour)

	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})

	return armour && !blastProt && !fireProt && !projectileProt
}
