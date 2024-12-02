package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// NewEgg creates an Egg entity. Egg is as a throwable entity that can be used
// to spawn chicks.
func NewEgg(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := eggConf
	conf.Owner = owner.H()
	return opts.New(EggType, conf)
}

// TODO: Spawn chicken(e) 12.5% of the time.
var eggConf = ProjectileBehaviourConfig{
	Gravity:       0.03,
	Drag:          0.01,
	Particle:      particle.EggSmash{},
	ParticleCount: 6,
}

// EggType is a world.EntityType implementation for Egg.
var EggType eggType

type eggType struct{}

func (t eggType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (eggType) EncodeEntity() string { return "minecraft:egg" }
func (eggType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (eggType) DecodeNBT(_ map[string]any, data *world.EntityData) { data.Data = eggConf.New() }
func (eggType) EncodeNBT(_ *world.EntityData) map[string]any       { return nil }
