package entity

import "github.com/df-mc/dragonfly/server/world"

// behaviourDamageable represents a Behaviour that may be hurt directly without
// implementing Living.
type behaviourDamageable interface {
	Hurt(e *Ent, damage float64, src world.DamageSource) (n float64, vulnerable bool)
}

// HurtEntity hurts an entity if it is either Living or has a Behaviour that may
// be hurt directly. It returns the damage dealt, whether the entity was
// vulnerable to the damage, and whether the entity could be damaged.
func HurtEntity(e world.Entity, damage float64, src world.DamageSource) (n float64, vulnerable, ok bool) {
	if l, ok := e.(Living); ok {
		n, vulnerable = l.Hurt(damage, src)
		return n, vulnerable, true
	}
	if ent, ok := e.(*Ent); ok {
		if d, ok := ent.Behaviour().(behaviourDamageable); ok {
			n, vulnerable = d.Hurt(ent, damage, src)
			return n, vulnerable, true
		}
	}
	return 0, false, false
}

// damageableEntity checks if an entity may be damaged.
func damageableEntity(e world.Entity) bool {
	if _, ok := e.(Living); ok {
		return true
	}
	if ent, ok := e.(*Ent); ok {
		_, ok = ent.Behaviour().(behaviourDamageable)
		return ok
	}
	return false
}
