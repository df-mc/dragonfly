package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewEndCrystal creates a new End crystal entity.
func NewEndCrystal(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(EndCrystalType, endCrystalConf{})
}

type endCrystalConf struct{}

func (endCrystalConf) Apply(data *world.EntityData) {
	data.Data = endCrystalBehaviour{}
}

type endCrystalBehaviour struct{}

func (endCrystalBehaviour) Tick(*Ent, *world.Tx) *Movement {
	return nil
}

func (endCrystalBehaviour) Explode(e *Ent, _ mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	if impact > 0 {
		explodeEndCrystal(e)
	}
}

// EndCrystal is a stationary explosive entity spawned by using an End crystal item.
type EndCrystal struct {
	*Ent
}

// Health always returns 1.
func (*EndCrystal) Health() float64 {
	return 1
}

// MaxHealth always returns 1.
func (*EndCrystal) MaxHealth() float64 {
	return 1
}

// SetMaxHealth does nothing for End crystals.
func (*EndCrystal) SetMaxHealth(float64) {}

// Dead always returns false for live End crystal handles.
func (*EndCrystal) Dead() bool {
	return false
}

// Hurt destroys the End crystal and creates its explosion.
func (e *EndCrystal) Hurt(damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	if damage == 0 {
		return 0, false
	}
	if _, ok := src.(VoidDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	e.explode()
	return damage, true
}

// Heal does nothing for End crystals.
func (*EndCrystal) Heal(float64, world.HealingSource) {}

// KnockBack does nothing for End crystals.
func (*EndCrystal) KnockBack(mgl64.Vec3, float64, float64) {}

// AddEffect does nothing for End crystals.
func (*EndCrystal) AddEffect(effect.Effect) {}

// RemoveEffect does nothing for End crystals.
func (*EndCrystal) RemoveEffect(effect.Type) {}

// Effects always returns nil for End crystals.
func (*EndCrystal) Effects() []effect.Effect {
	return nil
}

// Speed always returns 0 for End crystals.
func (*EndCrystal) Speed() float64 {
	return 0
}

// SetSpeed does nothing for End crystals.
func (*EndCrystal) SetSpeed(float64) {}

// Immobile always returns true for End crystals.
func (*EndCrystal) Immobile() bool {
	return true
}

// ShowBase returns whether the End crystal should show its bottom base.
func (*EndCrystal) ShowBase() bool {
	return false
}

// BeamTarget returns the End crystal's beam target, if any.
func (*EndCrystal) BeamTarget() (cube.Pos, bool) {
	return cube.Pos{}, false
}

func (e *EndCrystal) explode() {
	explodeEndCrystal(e.Ent)
}

func explodeEndCrystal(e *Ent) {
	if _, ok := e.H().Entity(e.tx); !ok {
		return
	}
	pos := e.Position()
	_ = e.Close()
	block.ExplosionConfig{Size: 6}.Explode(e.tx, pos)
}

// EndCrystalType is a world.EntityType implementation for End crystals.
var EndCrystalType endCrystalType

type endCrystalType struct{}

func (t endCrystalType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &EndCrystal{Ent: Open(tx, handle, data)}
}

func (endCrystalType) EncodeEntity() string {
	return "minecraft:ender_crystal"
}

func (endCrystalType) BBox(world.Entity) cube.BBox {
	return cube.Box(-1, 0, -1, 1, 2, 1)
}

func (endCrystalType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	endCrystalConf{}.Apply(data)
}

func (endCrystalType) EncodeNBT(*world.EntityData) map[string]any {
	return nil
}
