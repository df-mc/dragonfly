package entity

import (
	"math/rand/v2"
	"time"

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

// endCrystalExploder may be implemented by a damage source to control whether
// it makes End crystals explode.
type endCrystalExploder interface {
	ExplodesEndCrystal() bool
}

func (b endCrystalBehaviour) Apply(data *world.EntityData) {
	data.Data = b
}

func (endCrystalBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if tx.World().Dimension() == world.End {
		pos := cube.PosFromVec3(e.Position())
		if _, ok := tx.Block(pos).(block.Fire); !ok {
			flame := block.Fire{}
			tx.SetBlock(pos, flame, nil)
			tx.ScheduleBlockUpdate(pos, flame, time.Duration(30+rand.IntN(10))*time.Second/20)
		}
	}
	return nil
}

func (endCrystalBehaviour) Explode(e *Ent, _ mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	if impact > 0 {
		_ = e.Close()
	}
}

func (endCrystalBehaviour) Hurt(e *Ent, damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	if _, ok := src.(VoidDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	if _, ok := src.(ExplosionDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	if exploder, ok := src.(endCrystalExploder); ok && !exploder.ExplodesEndCrystal() {
		return damage, false
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
