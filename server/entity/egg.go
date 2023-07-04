package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// NewEgg creates an Egg entity. Egg is as a throwable entity that can be used
// to spawn chicks.
func NewEgg(pos mgl64.Vec3, owner world.Entity) *Ent {
	return Config{Behaviour: eggConf.New(owner)}.New(EggType{}, pos)
}

// TODO: Spawn chicken(e) 12.5% of the time.
var eggConf = ProjectileBehaviourConfig{
	Gravity:       0.03,
	Drag:          0.01,
	Particle:      particle.EggSmash{},
	ParticleCount: 6,
}

// EggType is a world.EntityType implementation for Egg.
type EggType struct{}

func (EggType) EncodeEntity() string { return "minecraft:egg" }
func (EggType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (EggType) DecodeNBT(m map[string]any) world.Entity {
	egg := NewEgg(nbtconv.Vec3(m, "Pos"), nil)
	egg.vel = nbtconv.Vec3(m, "Motion")
	return egg
}

func (EggType) EncodeNBT(e world.Entity) map[string]any {
	egg := e.(*Ent)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(egg.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(egg.Velocity()),
	}
}
