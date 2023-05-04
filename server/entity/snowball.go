package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// NewSnowball creates a snowball entity at a position with an owner entity.
func NewSnowball(pos mgl64.Vec3, owner world.Entity) *Ent {
	return Config{Behaviour: snowballConf.New(owner)}.New(SnowballType{}, pos)
}

var snowballConf = ProjectileBehaviourConfig{
	Gravity:       0.03,
	Drag:          0.01,
	Particle:      particle.SnowballPoof{},
	ParticleCount: 6,
}

// SnowballType is a world.EntityType implementation for snowballs.
type SnowballType struct{}

func (SnowballType) EncodeEntity() string { return "minecraft:snowball" }
func (SnowballType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (SnowballType) DecodeNBT(m map[string]any) world.Entity {
	s := NewSnowball(nbtconv.Vec3(m, "Pos"), nil)
	s.vel = nbtconv.Vec3(m, "Motion")
	return s
}

func (SnowballType) EncodeNBT(e world.Entity) map[string]any {
	s := e.(*Ent)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(s.Velocity()),
	}
}
