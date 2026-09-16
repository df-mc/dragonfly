package entity

import (
	"iter"

	"github.com/df-mc/dragonfly/server/world"
)

// Hurtable is a world.Entity that may be hurt directly. It is the least the
// shared damage helpers need of an entity, and is deliberately smaller than
// Damageable: a falling or burning entity need not carry health of its own.
type Hurtable interface {
	world.Entity
	// Hurt hurts the entity for the damage passed. It returns the damage dealt
	// and whether the entity was vulnerable to it.
	Hurt(damage float64, src world.DamageSource) (n float64, vulnerable bool)
}

// DamageableBehaviour may be implemented by a Behaviour to let its entity take damage without implementing
// the full Living interface. HurtEntity and Ent.Hurt dispatch to it.
type DamageableBehaviour interface {
	// Hurt hurts the entity of the Behaviour. It returns the damage dealt and whether the entity was
	// vulnerable to it.
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
	if w, ok := e.(wrappedEnt); ok {
		ent := w.Base()
		if _, ok := ent.Behaviour().(DamageableBehaviour); ok {
			n, vulnerable = ent.Hurt(damage, src)
			return n, vulnerable, true
		}
	}
	return 0, false, false
}

// DamageableEntity checks if an entity may be damaged.
func DamageableEntity(e world.Entity) bool {
	if _, ok := e.(Living); ok {
		return true
	}
	if w, ok := e.(wrappedEnt); ok {
		_, ok = w.Base().Behaviour().(DamageableBehaviour)
		return ok
	}
	return false
}

// filterDamageable filters an entity sequence down to the entities that may be damaged, as reported by
// DamageableEntity.
func filterDamageable(seq iter.Seq[world.Entity]) iter.Seq[world.Entity] {
	return func(yield func(world.Entity) bool) {
		for e := range seq {
			if !DamageableEntity(e) {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}
