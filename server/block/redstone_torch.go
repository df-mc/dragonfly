package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
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
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateStrongRedstone(pos, tx)
	})
}

// UseOnBlock ...
func (t RedstoneTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		return false
	}
	if _, ok := tx.Block(pos).(world.Liquid); ok {
		return false
	}
	if !tx.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, tx) {
		found := false
		for _, i := range []cube.Face{cube.FaceSouth, cube.FaceWest, cube.FaceNorth, cube.FaceEast, cube.FaceDown} {
			if tx.Block(pos.Side(i)).Model().FaceSolid(pos.Side(i), i.Opposite(), tx) {
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

	place(tx, pos, t, user, ctx)
	if placed(ctx) {
		t.RedstoneUpdate(pos, tx)
		updateStrongRedstone(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
		updateDirectionalRedstone(pos, tx, t.Facing.Opposite())
	}
}

// RedstoneUpdate ...
func (t RedstoneTorch) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	if t.inputStrength(pos, tx) > 0 != t.Lit {
		return
	}
	tx.ScheduleBlockUpdate(pos, t, time.Millisecond*100)
}

// ScheduledTick ...
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if t.inputStrength(pos, tx) > 0 != t.Lit {
		return
	}
	t.Lit = !t.Lit
	tx.SetBlock(pos, t, nil)
	updateStrongRedstone(pos, tx)
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

// RedstoneSource ...
func (t RedstoneTorch) RedstoneSource() bool {
	return t.Lit
}

// WeakPower ...
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if !t.Lit {
		return 0
	}
	if face != t.Facing.Opposite() {
		return 15
	}
	return 0
}

// StrongPower ...
func (t RedstoneTorch) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if t.Lit && face == cube.FaceDown {
		return 15
	}
	return 0
}

// inputStrength ...
func (t RedstoneTorch) inputStrength(pos cube.Pos, tx *world.Tx) int {
	return tx.RedstonePower(pos.Side(t.Facing), t.Facing, true)
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
