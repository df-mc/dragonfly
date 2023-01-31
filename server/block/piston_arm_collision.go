package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// PistonArmCollision is a block that is used when a piston is extended and colliding with a block.
type PistonArmCollision struct {
	empty
	transparent

	// Facing represents the direction the piston is facing.
	Facing cube.Face
	// Sticky is true if the piston arm is sticky.
	Sticky bool
}

// BreakInfo ...
func (c PistonArmCollision) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, simpleDrops()).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		pistonPos := pos.Side(c.pistonFace())
		if p, ok := w.Block(pistonPos).(Piston); ok {
			w.SetBlock(pistonPos, nil, nil)
			dropItem(w, item.NewStack(p, 1), pos.Vec3Centre())
		}
	})
}

// EncodeBlock ...
func (c PistonArmCollision) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston_arm_collision", map[string]any{"facing_direction": int32(c.Facing)}
}

// NeighbourUpdateTick ...
func (c PistonArmCollision) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Block(pos.Side(c.pistonFace())).(Piston); !ok {
		w.SetBlock(pos, nil, nil)
	}
}

// pistonFace ...
func (c PistonArmCollision) pistonFace() cube.Face {
	if c.Facing.Axis() != cube.Y {
		return c.Facing
	}
	return c.Facing.Opposite()
}
