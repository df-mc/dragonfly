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
	sourceWaterDisplacer

	// Facing represents the direction the piston is facing.
	Facing cube.Face
}

// PistonImmovable ...
func (PistonArmCollision) PistonImmovable() bool {
	return true
}

// SideClosed ...
func (PistonArmCollision) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeBlock ...
func (c PistonArmCollision) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston_arm_collision", map[string]any{"facing_direction": int32(c.Facing)}
}

// BreakInfo ...
func (c PistonArmCollision) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, simpleDrops()).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		pistonPos := pos.Side(c.pistonFace())
		if p, ok := w.Block(pistonPos).(Piston); ok {
			w.SetBlock(pistonPos, nil, nil)
			if g, ok := u.(interface {
				GameMode() world.GameMode
			}); !ok || !g.GameMode().CreativeInventory() {
				dropItem(w, item.NewStack(Piston{Sticky: p.Sticky}, 1), pos.Vec3Centre())
			}
		}
	})
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

// allPistonArmCollisions ...
func allPistonArmCollisions() (pistonArmCollisions []world.Block) {
	for _, f := range cube.Faces() {
		pistonArmCollisions = append(pistonArmCollisions, PistonArmCollision{Facing: f})
	}
	return
}
