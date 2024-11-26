package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// NewEnderPearl creates an EnderPearl entity. EnderPearl is a smooth, greenish-
// blue item used to teleport.
func NewEnderPearl(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := enderPearlConf
	conf.Owner = owner.H()
	return opts.New(EnderPearlType, conf)
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
func teleport(e *Ent, tx *world.Tx, target trace.Result) {
	owner, _ := e.Behaviour().(*ProjectileBehaviour).Owner().Entity(tx)
	if user, ok := owner.(teleporter); ok {
		tx.PlaySound(user.Position(), sound.Teleport{})
		user.Teleport(target.Position())
		user.Hurt(5, FallDamageSource{})
	}
}

// EnderPearlType is a world.EntityType implementation for EnderPearl.
var EnderPearlType enderPearlType

type enderPearlType struct{}

func (t enderPearlType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (enderPearlType) EncodeEntity() string { return "minecraft:ender_pearl" }
func (enderPearlType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}
func (enderPearlType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = enderPearlConf.New()
}
func (enderPearlType) EncodeNBT(*world.EntityData) map[string]any { return nil }
