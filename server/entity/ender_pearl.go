package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// NewEnderPearl creates an EnderPearl entity. EnderPearl is a smooth, greenish-
// blue item used to teleport.
func NewEnderPearl(pos mgl64.Vec3, owner world.Entity) *Ent {
	return Config{Behaviour: enderPearlConf.New(owner)}.New(EnderPearlType{}, pos)
}

var enderPearlConf = ProjectileBehaviourConfig{
	Gravity:  0.03,
	Drag:     0.01,
	Particle: particle.EndermanTeleport{},
	Sound:    sound.Teleport{},
	Hit:      teleport,
}

// teleporter represents a living entity that can teleport.
type teleporter interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
	Living
}

// teleport teleports the owner of an Ent to a trace.Result's position.
func teleport(e *Ent, target trace.Result) {
	if user, ok := e.Owner().(teleporter); ok {
		e.World().PlaySound(user.Position(), sound.Teleport{})
		user.Teleport(target.Position())
		user.Hurt(5, FallDamageSource{})
	}
}

// EnderPearlType is a world.EntityType implementation for EnderPearl.
type EnderPearlType struct{}

func (EnderPearlType) EncodeEntity() string { return "minecraft:ender_pearl" }
func (EnderPearlType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (EnderPearlType) DecodeNBT(m map[string]any) world.Entity {
	ep := NewEnderPearl(nbtconv.Vec3(m, "Pos"), nil)
	ep.vel = nbtconv.Vec3(m, "Motion")
	return ep
}

func (EnderPearlType) EncodeNBT(e world.Entity) map[string]any {
	ep := e.(*Ent)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(ep.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(ep.Velocity()),
	}
}
