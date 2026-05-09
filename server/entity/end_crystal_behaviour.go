package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type endCrystalBehaviour struct {
	showBase      bool
	beamTarget    cube.Pos
	hasBeamTarget bool
}

func (b endCrystalBehaviour) Apply(data *world.EntityData) {
	data.Data = b
}

func (endCrystalBehaviour) Tick(*Ent, *world.Tx) *Movement {
	return nil
}

func (endCrystalBehaviour) Explode(e *Ent, _ mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	if impact > 0 {
		explodeEndCrystal(e)
	}
}

func (endCrystalBehaviour) Hurt(e *Ent, damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	if damage == 0 {
		return 0, false
	}
	if _, ok := src.(VoidDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	explodeEndCrystal(e)
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

func explodeEndCrystal(e *Ent) {
	if _, ok := e.H().Entity(e.tx); !ok {
		return
	}
	pos := e.Position()
	conf := block.ExplosionConfig{Size: 6, EndCrystal: true}
	_ = e.Close()
	conf.Explode(e.tx, pos)
}
