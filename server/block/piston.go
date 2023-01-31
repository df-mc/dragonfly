package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Piston is a block capable of pushing blocks, players, and mobs when given a redstone pulse.
type Piston struct {
	solid

	// Facing represents the direction the piston is facing.
	Facing cube.Face
	// Sticky is true if the piston is sticky, false if not.
	Sticky bool

	// AttachedBlocks ...
	// TODO: Make this a []cube.Pos and convert to []int32 on encode.
	AttachedBlocks []int32
	// BreakBlocks ...
	// TODO: Make this a []cube.Pos and convert to []int32 on encode.
	BreakBlocks []int32

	// Progress is how far the block has been moved. It can either be 0.0, 0.5, or 1.0.
	Progress float64
	// LastProgress ...
	LastProgress float64

	// State ...
	State int
	// NewState ...
	NewState int
}

// BreakInfo ...
func (p Piston) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, oneOf(p))
}

// EncodeItem ...
func (p Piston) EncodeItem() (name string, meta int16) {
	if p.Sticky {
		return "minecraft:sticky_piston", 0
	}
	return "minecraft:piston", 0
}

// EncodeBlock ...
func (p Piston) EncodeBlock() (string, map[string]any) {
	if p.Sticky {
		return "minecraft:sticky_piston", map[string]any{"facing_direction": int32(p.Facing)}
	}
	return "minecraft:piston", map[string]any{"facing_direction": int32(p.Facing)}
}

// UseOnBlock ...
func (p Piston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, p)
	if !used {
		return false
	}
	p.Facing = calculateAnySidedFace(user, pos, false)

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeNBT ...
func (p Piston) EncodeNBT() map[string]any {
	return map[string]any{
		"AttachedBlocks": p.AttachedBlocks,
		"BreakBlocks":    p.BreakBlocks,

		"LastProgress": float32(p.LastProgress),
		"Progress":     float32(p.Progress),

		"NewState": uint8(p.NewState),
		"State":    uint8(p.State),

		"Sticky": boolByte(p.Sticky),

		"id": "PistonArm",
	}
}

// DecodeNBT ...
func (p Piston) DecodeNBT(m map[string]any) any {
	if attached := nbtconv.Slice(m, "AttachedBlocks"); attached != nil {
		p.AttachedBlocks = make([]int32, len(attached))
		for i, v := range attached {
			p.AttachedBlocks[i] = v.(int32)
		}
	}
	if breakBlocks := nbtconv.Slice(m, "BreakBlocks"); breakBlocks != nil {
		p.BreakBlocks = make([]int32, len(breakBlocks))
		for i, v := range breakBlocks {
			p.BreakBlocks[i] = v.(int32)
		}
	}
	p.LastProgress = float64(nbtconv.Float32(m, "LastProgress"))
	p.Progress = float64(nbtconv.Float32(m, "Progress"))
	p.NewState = int(nbtconv.Uint8(m, "NewState"))
	p.State = int(nbtconv.Uint8(m, "State"))
	p.Sticky = nbtconv.Bool(m, "Sticky")
	return p
}

// RedstoneUpdate ...
func (Piston) RedstoneUpdate(pos cube.Pos, w *world.World) {
	w.ScheduleBlockUpdate(pos, time.Millisecond*50)
}

// ScheduledTick ...
func (p Piston) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if receivedRedstonePower(pos, w, p.armFace()) {
		if !p.push(pos, w) {
			return
		}
	} else if !p.pull(pos, w) {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Millisecond*50)
}

// armFace ...
func (p Piston) armFace() cube.Face {
	if p.Facing.Axis() == cube.Y {
		return p.Facing
	}
	return p.Facing.Opposite()
}

// push ...
func (p Piston) push(pos cube.Pos, w *world.World) bool {
	//TODO: Implement.
	return false
}

// pull ...
func (p Piston) pull(pos cube.Pos, w *world.World) bool {
	//TODO: Implement.
	return false
}

// allPistons ...
func allPistons() (pistons []world.Block) {
	for _, f := range cube.Faces() {
		for _, s := range []bool{false, true} {
			pistons = append(pistons, Piston{Facing: f, Sticky: s})
		}
	}
	return
}
