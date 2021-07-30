package entity

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// FallingBlock is the entity form of a block that appears when a gravity-affected block loses its support.
type FallingBlock struct {
	transform
	block world.Block

	c *MovementComputer
}

// NewFallingBlock ...
func NewFallingBlock(block world.Block, pos mgl64.Vec3) *FallingBlock {
	b := &FallingBlock{block: block, c: &MovementComputer{
		Gravity:           0.04,
		DragBeforeGravity: true,
		Drag:              0.02,
	}}
	b.transform = newTransform(b, pos)
	return b
}

// Name ...
func (f *FallingBlock) Name() string {
	return fmt.Sprintf("%T", f.block)
}

// EncodeEntity ...
func (f *FallingBlock) EncodeEntity() string {
	return "minecraft:falling_block"
}

// AABB ...
func (f *FallingBlock) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.49, 0, -0.49}, mgl64.Vec3{0.49, 0.98, 0.49})
}

// Block ...
func (f *FallingBlock) Block() world.Block {
	return f.block
}

// Tick ...
func (f *FallingBlock) Tick(_ int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.pos, f.vel = f.c.TickMovement(f, f.pos, f.vel, 0, 0)
	w := f.World()

	pos := cube.PosFromVec3(f.pos)

	if a, ok := f.block.(Solidifiable); (ok && a.Solidifies(pos, w)) || f.c.OnGround() {
		b := f.World().Block(pos)
		if r, ok := b.(replaceable); ok && r.ReplaceableBy(f.block) {
			f.World().PlaceBlock(pos, f.block)
		} else {
			if i, ok := f.block.(world.Item); ok {
				f.World().AddEntity(NewItem(item.NewStack(i, 1), f.pos))
			}
		}

		_ = f.Close()
	}
}

// DecodeNBT decodes the relevant data from the entity NBT passed and returns a new FallingBlock entity.
func (f *FallingBlock) DecodeNBT(data map[string]interface{}) interface{} {
	b := nbtconv.MapBlock(data, "FallingBlock")
	if b == nil {
		return nil
	}
	n := NewFallingBlock(b, nbtconv.MapVec3(data, "Pos"))
	n.SetVelocity(nbtconv.MapVec3(data, "Motion"))
	return n
}

// EncodeNBT encodes the FallingBlock entity to a map that can be encoded for NBT.
func (f *FallingBlock) EncodeNBT() map[string]interface{} {
	pos, vel := f.Position(), f.Velocity()
	name, properties := f.block.EncodeBlock()
	return map[string]interface{}{
		"UniqueID": -rand.Int63(),
		"Pos":      nbtconv.Vec3ToFloat32Slice(pos),
		"Motion":   nbtconv.Vec3ToFloat32Slice(vel),
		"FallingBlock": map[string]interface{}{
			"name":    name,
			"states":  properties,
			"version": chunk.CurrentBlockVersion,
		},
	}
}

// Solidifiable represents a block that can solidify by specific adjacent blocks. An example is concrete
// powder, which can turn into concrete by touching water.
type Solidifiable interface {
	// Solidifies returns whether the falling block can solidify at the position it is currently in. If so,
	// the block will immediately stop falling.
	Solidifies(pos cube.Pos, w *world.World) bool
}

type replaceable interface {
	ReplaceableBy(b world.Block) bool
}
