package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
)

// NewFallingBlock creates a new FallingBlock entity.
func NewFallingBlock(opts world.EntitySpawnOpts, block world.Block) *world.EntityHandle {
	conf := fallingBlockConf
	conf.Block = block
	return opts.New(FallingBlockType, conf)
}

var fallingBlockConf = FallingBlockBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
}

// FallingBlockType is a world.EntityType implementation for FallingBlock.
var FallingBlockType fallingBlockType

type fallingBlockType struct{}

func (t fallingBlockType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}
func (fallingBlockType) EncodeEntity() string   { return "minecraft:falling_block" }
func (fallingBlockType) NetworkOffset() float64 { return 0.49 }
func (fallingBlockType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

func (fallingBlockType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := fallingBlockConf
	conf.Block = nbtconv.Block(m, "FallingBlock")
	conf.DistanceFallen = nbtconv.Float64(m, "FallDistance")
	data.Data = conf.New()
}

func (fallingBlockType) EncodeNBT(data *world.EntityData) map[string]any {
	b := data.Data.(*FallingBlockBehaviour)
	return map[string]any{"FallDistance": b.passive.fallDistance, "FallingBlock": nbtconv.WriteBlock(b.block)}
}
