package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneTorch is a non-solid blocks that emits little light and also a full-strength redstone signal when lit.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Lit is if the redstone torch is lit and emitting power.
	Lit bool
}

// BreakInfo ...
func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// LightEmissionLevel ...
func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

// UseOnBlock ...
func (t RedstoneTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		return false
	}
	if _, ok := w.Block(pos).(world.Liquid); ok {
		return false
	}
	if !w.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, w) {
		found := false
		for _, i := range []cube.Face{cube.FaceSouth, cube.FaceWest, cube.FaceNorth, cube.FaceEast, cube.FaceDown} {
			if w.Block(pos.Side(i)).Model().FaceSolid(pos.Side(i), i.Opposite(), w) {
				found = true
				face = i.Opposite()
				break
			}
		}
		if !found {
			return false
		}
	}
	t.Facing = face.Opposite()
	t.Lit = true

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !w.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), w) {
		w.SetBlock(pos, nil, nil)
	}
}

// HasLiquidDrops ...
func (t RedstoneTorch) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (t RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

// EncodeBlock ...
func (t RedstoneTorch) EncodeBlock() (name string, properties map[string]any) {
	face := t.Facing.String()
	if t.Facing == cube.FaceDown {
		face = "top"
	}
	if t.Lit {
		return "minecraft:redstone_torch", map[string]any{"torch_facing_direction": face}
	}
	return "minecraft:unlit_redstone_torch", map[string]any{"torch_facing_direction": face}
}

// WeakPower ...
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.World) int {
	if t.Lit && face != cube.FaceUp {
		return 15
	}
	return 0
}

// StrongPower ...
func (t RedstoneTorch) StrongPower(cube.Pos, cube.Face, *world.World) int {
	return 0
}

// allRedstoneTorches ...
func allRedstoneTorches() (all []world.Block) {
	for i := cube.Face(0); i < 6; i++ {
		if i == cube.FaceUp {
			continue
		}
		all = append(all, RedstoneTorch{Facing: i, Lit: true})
		all = append(all, RedstoneTorch{Facing: i})
	}
	return
}
