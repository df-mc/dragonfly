package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// KnockBackVector returns the velocity that knocks back an entity at pos away from src with the force and
// height passed.
func KnockBackVector(pos, src mgl64.Vec3, force, height float64) mgl64.Vec3 {
	velocity := pos.Sub(src)
	velocity[1] = 0

	if velocity.Len() != 0 {
		velocity = velocity.Normalize().Mul(force)
	}
	velocity[1] = height
	return velocity
}

// ExplosionDamage returns the damage an explosion of the size passed deals to an entity exposed to it with
// the impact passed, following the vanilla formula.
func ExplosionDamage(size, impact float64) float64 {
	return math.Floor((impact*impact+impact)*3.5*size*2 + 1)
}

// FinalDamage returns the damage dealt to an entity after the armour worn and an active resistance effect
// reduced the damage passed.
func FinalDamage(dmg float64, src world.DamageSource, armour *inventory.Armour, effects *EffectManager) float64 {
	dmg = max(dmg, 0)
	dmg -= armour.DamageReduction(dmg, src)
	if res, ok := effects.Effect(effect.Resistance); ok {
		dmg *= effect.Resistance.Multiplier(src, res.Level())
	}
	return dmg
}

// Fall lands an entity that fell the distance passed. The block landed on may soften the distance, a jump
// boost effect reduces it further and the remainder is dealt as fall damage.
func Fall(e Hurtable, tx *world.Tx, effects *EffectManager, distance float64) {
	distance = CheckEntityLanders(tx, e, e.H().Type().BBox(e).Translate(e.Position()), distance)

	dmg := distance - 3
	if boost, ok := effects.Effect(effect.JumpBoost); ok {
		dmg -= float64(boost.Level())
	}
	if dmg < 0.5 {
		return
	}
	e.Hurt(math.Ceil(dmg), FallDamageSource{})
}
