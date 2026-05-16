package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// NewRaft creates a static raft entity. If chest is true, a chest raft is created.
func NewRaft(opts world.EntitySpawnOpts, chest bool) *world.EntityHandle {
	if chest {
		return opts.New(ChestRaftType, raftBehaviourConfig{Chest: true})
	}
	return opts.New(RaftType, raftBehaviourConfig{})
}

var raftConf StationaryBehaviourConfig

type raftBehaviourConfig struct {
	Chest bool
}

func (c raftBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = &raftBehaviour{StationaryBehaviour: raftConf.New(), chest: c.Chest}
}

type raftBehaviour struct {
	*StationaryBehaviour
	chest bool
}

// Variant returns the wood variant for bamboo boats/rafts.
func (r *raftBehaviour) Variant() int32 {
	return 7
}

// MarkVariant is used by the client for extra boat model variation.
func (r *raftBehaviour) MarkVariant() int32 {
	if r.chest {
		return 1
	}
	return 0
}

// RaftType is a world.EntityType implementation for a bamboo raft.
var RaftType raftType

// ChestRaftType is a world.EntityType implementation for a bamboo chest raft.
var ChestRaftType chestRaftType

type raftType struct{}
type chestRaftType struct{}

func (t raftType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (t chestRaftType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (raftType) EncodeEntity() string      { return "minecraft:boat" }
func (chestRaftType) EncodeEntity() string { return "minecraft:chest_boat" }

func (raftType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.6875, 0, -0.6875, 0.6875, 0.5625, 0.6875)
}

func (chestRaftType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.6875, 0, -0.6875, 0.6875, 0.5625, 0.6875)
}

func (raftType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = &raftBehaviour{StationaryBehaviour: raftConf.New()}
}
func (chestRaftType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = &raftBehaviour{StationaryBehaviour: raftConf.New(), chest: true}
}

func (raftType) EncodeNBT(_ *world.EntityData) map[string]any      { return nil }
func (chestRaftType) EncodeNBT(_ *world.EntityData) map[string]any { return nil }
