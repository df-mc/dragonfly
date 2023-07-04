package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// NewFallingBlock creates a new FallingBlock entity.
func NewFallingBlock(block world.Block, pos mgl64.Vec3) *Ent {
	return Config{Behaviour: fallingBlockConf.New(block)}.New(FallingBlockType{}, pos)
}

var fallingBlockConf = FallingBlockBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
}

// FallingBlockType is a world.EntityType implementation for FallingBlock.
type FallingBlockType struct{}

func (FallingBlockType) EncodeEntity() string   { return "minecraft:falling_block" }
func (FallingBlockType) NetworkOffset() float64 { return 0.49 }
func (FallingBlockType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

func (FallingBlockType) DecodeNBT(m map[string]any) world.Entity {
	b := nbtconv.Block(m, "FallingBlock")
	if b == nil {
		return nil
	}
	n := NewFallingBlock(b, nbtconv.Vec3(m, "Pos"))
	n.SetVelocity(nbtconv.Vec3(m, "Motion"))
	n.Behaviour().(*FallingBlockBehaviour).passive.fallDistance = nbtconv.Float64(m, "FallDistance")
	return n
}

func (FallingBlockType) EncodeNBT(e world.Entity) map[string]any {
	f := e.(*Ent)
	b := f.Behaviour().(*FallingBlockBehaviour)
	return map[string]any{
		"UniqueID":     -rand.Int63(),
		"FallDistance": b.passive.fallDistance,
		"Pos":          nbtconv.Vec3ToFloat32Slice(f.Position()),
		"Motion":       nbtconv.Vec3ToFloat32Slice(f.Velocity()),
		"FallingBlock": nbtconv.WriteBlock(b.block),
	}
}
