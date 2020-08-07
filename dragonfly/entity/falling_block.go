package entity

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/entity/state"
	"github.com/df-mc/dragonfly/dragonfly/internal/entity_internal"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync/atomic"
)

// FallingBlock is the entity form of a block that appears when a gravity-affected block loses its support.
type FallingBlock struct {
	block         world.Block
	velocity, pos atomic.Value

	*movementComputer
}

// NewFallingBlock ...
func NewFallingBlock(block world.Block, pos mgl64.Vec3) *FallingBlock {
	f := &FallingBlock{block: block, movementComputer: &movementComputer{gravity: 0.04, dragBeforeGravity: true}}
	f.pos.Store(pos)
	f.velocity.Store(mgl64.Vec3{})
	return f
}

// Block ...
func (f *FallingBlock) Block() world.Block {
	return f.block
}

// Tick ...
func (f *FallingBlock) Tick(_ int64) {
	f.pos.Store(f.tickMovement(f))

	pos := world.BlockPosFromVec3(f.Position())
	if f.OnGround() || entity_internal.CanSolidify(f.block, pos, f.World()) {
		if item_internal.Replaceable(f.World(), pos, f.block) {
			f.World().PlaceBlock(pos, f.block)
		} else {
			if i, ok := f.block.(world.Item); ok {
				itemEntity := NewItem(item.NewStack(i, 1), f.Position())
				itemEntity.SetVelocity(mgl64.Vec3{})
				f.World().AddEntity(itemEntity)
			}
		}

		f.Close()
	}
}

// Close ...
func (f *FallingBlock) Close() error {
	if f.World() != nil {
		f.World().RemoveEntity(f)
	}
	return nil
}

// AABB ...
func (f *FallingBlock) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{0.98, 0.98, 0.98})
}

// Position ...
func (f *FallingBlock) Position() mgl64.Vec3 {
	return f.pos.Load().(mgl64.Vec3)
}

// World ...
func (f *FallingBlock) World() *world.World {
	w, _ := world.OfEntity(f)
	return w
}

// Yaw ...
func (f *FallingBlock) Yaw() float64 {
	return 0
}

// Pitch ...
func (f *FallingBlock) Pitch() float64 {
	return 0
}

// State ...
func (f *FallingBlock) State() []state.State {
	return nil
}

// Velocity ...
func (f *FallingBlock) Velocity() mgl64.Vec3 {
	return f.velocity.Load().(mgl64.Vec3)
}

// SetVelocity ...
func (f *FallingBlock) SetVelocity(v mgl64.Vec3) {
	f.velocity.Store(v)
}
