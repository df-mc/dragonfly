package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NewEndCrystal creates a new End crystal entity.
func NewEndCrystal(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(EndCrystalType, endCrystalBehaviour{})
}

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

// EndCrystalType is a world.EntityType implementation for End crystals.
var EndCrystalType endCrystalType

type endCrystalType struct{}

func (endCrystalType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return Open(tx, handle, data)
}

func (endCrystalType) EncodeEntity() string {
	return "minecraft:ender_crystal"
}

func (endCrystalType) BBox(world.Entity) cube.BBox {
	return cube.Box(-1, 0, -1, 1, 2, 1)
}

func (endCrystalType) DecodeNBT(m map[string]any, data *world.EntityData) {
	b := endCrystalBehaviour{showBase: nbtconv.Bool(m, "ShowBottom")}
	x, xOK := m["BlockTargetX"].(int32)
	y, yOK := m["BlockTargetY"].(int32)
	z, zOK := m["BlockTargetZ"].(int32)
	if xOK && yOK && zOK {
		b.beamTarget = cube.Pos{int(x), int(y), int(z)}
		b.hasBeamTarget = true
	}
	b.Apply(data)
}

func (endCrystalType) EncodeNBT(data *world.EntityData) map[string]any {
	b := data.Data.(endCrystalBehaviour)
	m := map[string]any{"ShowBottom": boolByte(b.showBase)}
	if b.hasBeamTarget {
		m["BlockTargetX"] = int32(b.beamTarget[0])
		m["BlockTargetY"] = int32(b.beamTarget[1])
		m["BlockTargetZ"] = int32(b.beamTarget[2])
	}
	return m
}
