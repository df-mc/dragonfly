package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Web is a block that significantly slows down entities moving through it.
type Web struct {
	transparent
}

// BlastResistance returns the blast resistance of the web.
func (w Web) BlastResistance() float64 {
	return 4
}

// BreakInfo returns the break info for web.
func (w Web) BreakInfo() BreakInfo {
	return newBreakInfo(4, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears || t.ToolType() == item.TypeSword
	}, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears || t.ToolType() == item.TypeSword
	}, w.dropFunc)
}

// dropFunc determines the drops based on the tool used.
func (w Web) dropFunc(t item.Tool, _ []item.Enchantment) []item.Stack {
	if t.ToolType() == item.TypeShears {
		return []item.Stack{item.NewStack(w, 1)}
	}
	// Sword and default: no drops
	return []item.Stack{}
}

// EntityInside slows down entities inside the web.
func (w Web) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	// Use a minimal interface so we don't import entity package and cause cycles.
	if l, ok := e.(interface{ Velocity() mgl64.Vec3; SetVelocity(mgl64.Vec3) }); ok {
		v := l.Velocity()
		// Damp horizontal velocity strongly.
		vx, vz := v.X()*0.15, v.Z()*0.15
		vy := v.Y() * 0.15

		// If the vertical velocity is nearly zero, nudge the entity gently downwards
		// so that landing on the web does not freeze the entity in place.
		if vy > -0.01 && vy < 0.01 {
			vy = -0.03
		}

		l.SetVelocity(mgl64.Vec3{vx, vy, vz})
	}
}

// NeighbourUpdateTick handles water breaking the web.
func (w Web) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(changedNeighbour).(Water); ok {
		tx.SetBlock(pos, Air{}, nil)
	}
}

// Model returns the block model of the web.
func (w Web) Model() world.BlockModel {
	return model.Solid{}
}

// EncodeItem encodes the web as an item.
func (w Web) EncodeItem() (name string, meta int16) {
	return "minecraft:web", 0
}

// EncodeBlock encodes the web as a block.
func (w Web) EncodeBlock() (string, map[string]any) {
	return "minecraft:web", nil
}
