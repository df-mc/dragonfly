package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type endCrystalBehaviour struct {
	showBase      bool
	beamTarget    cube.Pos
	hasBeamTarget bool
	explosionSize float64
}

func (b endCrystalBehaviour) Apply(data *world.EntityData) {
	data.Data = b
}

// Tick continuously generates fire at the End crystal's position while in the
// End, if the block at that position is air.
func (endCrystalBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if tx.World().Dimension() == world.End {
		block.Fire{}.Start(tx, cube.PosFromVec3(e.Position()))
	}
	return nil
}

// Explode makes the End crystal explode itself when hit by another explosion,
// causing a chain reaction.
func (b endCrystalBehaviour) Explode(e *Ent, _ world.ExplosionSource, impact float64) {
	if impact <= 0 {
		return
	}
	explodeEndCrystal(e, b.explosionSize)
}

// Hurt makes the End crystal explode when damaged by any source, even by
// damage that deals no health. Void damage removes it without an explosion.
func (b endCrystalBehaviour) Hurt(e *Ent, damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	if _, ok := src.(VoidDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	explodeEndCrystal(e, b.explosionSize)
	return damage, true
}

func (endCrystalBehaviour) Immobile() bool {
	return true
}

func (b endCrystalBehaviour) ShowBase() bool {
	return b.showBase
}

func (b endCrystalBehaviour) BeamTarget() (cube.Pos, bool) {
	return b.beamTarget, b.hasBeamTarget
}

// explodeEndCrystal closes the End crystal and creates a non-incendiary
// explosion at its base, if the crystal was not closed yet.
func explodeEndCrystal(e *Ent, explosionSize float64) {
	if _, ok := e.H().Entity(e.tx); !ok {
		return
	}
	_ = e.Close()
	block.ExplosionConfig{
		SuppressUnderwaterImpact: true,
	}.Explode(e.tx, world.EntityExplosionSource{
		Entity:        e,
		ExplosionSize: explosionSize,
	})
}
