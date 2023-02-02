package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// RedstoneTorch is a non-solid blocks that emits little light and also a full-strength redstone signal when lit.
// TODO: Redstone torches should burn out when used too recently and excessively.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Lit is if the redstone torch is lit and emitting power.
	Lit bool
}

// HasLiquidDrops ...
func (RedstoneTorch) HasLiquidDrops() bool {
	return true
}

// LightEmissionLevel ...
func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

// BreakInfo ...
func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateDirectionalRedstone(pos, w, t.Facing.Opposite())
	})
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
	if placed(ctx) {
		t.RedstoneUpdate(pos, w)
		updateDirectionalRedstone(pos, w, t.Facing.Opposite())
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !w.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), w) {
		w.SetBlock(pos, nil, nil)
		dropItem(w, item.NewStack(t, 1), pos.Vec3Centre())
		updateDirectionalRedstone(pos, w, t.Facing.Opposite())
	}
}

// RedstoneUpdate ...
func (t RedstoneTorch) RedstoneUpdate(pos cube.Pos, w *world.World) {
	if t.inputStrength(pos, w) > 0 != t.Lit {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Millisecond*100)
}

// ScheduledTick ...
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if t.inputStrength(pos, w) > 0 != t.Lit {
		return
	}
	t.Lit = !t.Lit
	w.SetBlock(pos, t, nil)
	updateDirectionalRedstone(pos, w, t.Facing.Opposite())
}

// EncodeItem ...
func (RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

// EncodeBlock ...
func (t RedstoneTorch) EncodeBlock() (name string, properties map[string]any) {
	face := "unknown"
	if t.Facing != unknownFace {
		face = t.Facing.String()
		if t.Facing == cube.FaceDown {
			face = "top"
		}
	}
	if t.Lit {
		return "minecraft:redstone_torch", map[string]any{"torch_facing_direction": face}
	}
	return "minecraft:unlit_redstone_torch", map[string]any{"torch_facing_direction": face}
}

// Source ...
func (t RedstoneTorch) Source() bool {
	return t.Lit
}

// WeakPower ...
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if !t.Lit {
		return 0
	}
	if face == cube.FaceDown {
		if t.Facing.Opposite() != cube.FaceDown {
			return 15
		}
		return 0
	}
	if face != t.Facing.Opposite() {
		return 15
	}
	return 0
}

// StrongPower ...
func (t RedstoneTorch) StrongPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if t.Lit && face == cube.FaceDown {
		return 15
	}
	return 0
}

// inputStrength ...
func (t RedstoneTorch) inputStrength(pos cube.Pos, w *world.World) int {
	return w.RedstonePower(pos.Side(t.Facing), t.Facing, true)
}

// allRedstoneTorches ...
func allRedstoneTorches() (all []world.Block) {
	for _, f := range append(cube.Faces(), unknownFace) {
		if f == cube.FaceUp {
			continue
		}
		all = append(all, RedstoneTorch{Facing: f, Lit: true})
		all = append(all, RedstoneTorch{Facing: f})
	}
	return
}
