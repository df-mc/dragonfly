package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// NewSnowball creates a snowball entity at a position with an owner entity.
func NewSnowball(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := snowballConf
	conf.Owner = owner
	return opts.New(SnowballType{}, conf)
}

var snowballConf = ProjectileBehaviourConfig{
	Gravity:       0.03,
	Drag:          0.01,
	Particle:      particle.SnowballPoof{},
	ParticleCount: 6,
}

// SnowballType is a world.EntityType implementation for snowballs.
type SnowballType struct{}

func (t SnowballType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (SnowballType) EncodeEntity() string { return "minecraft:snowball" }
func (SnowballType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (SnowballType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = snowballConf.New()
}
func (SnowballType) EncodeNBT(*world.EntityData) map[string]any { return nil }
