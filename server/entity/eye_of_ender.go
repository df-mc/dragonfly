package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// NewEyeOfEnder creates a throwable eye of ender signal entity.
func NewEyeOfEnder(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := eyeOfEnderConf
	conf.Owner = owner.H()
	return opts.New(EyeOfEnderType, conf)
}

var eyeOfEnderConf = ProjectileBehaviourConfig{
	Gravity: 0.02,
	Drag:    0.01,
	Damage:  -1,
}

// EyeOfEnderType is a world.EntityType implementation for thrown eyes of
// ender.
var EyeOfEnderType eyeOfEnderType

type eyeOfEnderType struct{}

func (t eyeOfEnderType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (eyeOfEnderType) EncodeEntity() string { return "minecraft:eye_of_ender_signal" }
func (eyeOfEnderType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}
func (eyeOfEnderType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = eyeOfEnderConf.New()
}
func (eyeOfEnderType) EncodeNBT(*world.EntityData) map[string]any { return nil }
