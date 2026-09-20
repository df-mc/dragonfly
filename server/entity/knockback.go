package entity

// KnockbackSource is a source of knockback.
type KnockbackSource interface {
	// KnockbackSource ...
	KnockbackSource()
}

func (AttackDamageSource) KnockbackSource()     {}
func (ProjectileDamageSource) KnockbackSource() {}
func (ExplosionDamageSource) KnockbackSource()  {}
