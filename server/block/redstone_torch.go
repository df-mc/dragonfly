package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

var (
	_ world.RedstonePowerSource   = RedstoneTorch{}
	_ world.RedstonePowerConsumer = RedstoneTorch{}
	_ world.ScheduledTicker       = RedstoneTorch{}
)

// RedstoneTorch is a torch-like inverter that turns off while the block it is attached to is powered.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block it is attached to.
	Facing cube.Face
	// Lit is true if the torch is currently emitting redstone power.
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

// UseOnBlock places a redstone torch on the clicked block face.
func (t RedstoneTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used || face == cube.FaceDown {
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
	t.Lit = !t.attachmentPowered(pos, tx)

	place(tx, pos, t, user, ctx)
	if placed(ctx) {
		tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
	}
	return placed(ctx)
}

// NeighbourUpdateTick breaks unsupported torches and otherwise schedules inverse-state refreshes.
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
		return
	}
	tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
}

// ScheduledTick refreshes the lit state after the torch's one-redstone-tick inversion delay.
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if tx == nil {
		return
	}
	lit := !t.attachmentPowered(pos, tx)
	if t.Lit != lit {
		t.Lit = lit
		tx.SetBlock(pos, t, nil)
	}
}

// RedstonePower emits power from every side except the attached block while lit.
func (t RedstoneTorch) RedstonePower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if t.Lit && face != t.Facing {
		return 15
	}
	return 0
}

// RedstoneStrongPower strongly powers the block above the torch while lit.
func (t RedstoneTorch) RedstoneStrongPower(pos cube.Pos, tx *world.Tx, face cube.Face) int {
	if face == cube.FaceUp {
		return t.RedstonePower(pos, tx, face)
	}
	return 0
}

// RedstonePowerUpdate schedules the delayed inverse-state refresh.
func (t RedstoneTorch) RedstonePowerUpdate(pos cube.Pos, tx *world.Tx, _ int) (world.Block, bool) {
	if tx == nil || t.Lit == !t.attachmentPowered(pos, tx) {
		return t, false
	}
	tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
	return t, false
}

func (t RedstoneTorch) attachmentPowered(pos cube.Pos, tx *world.Tx) bool {
	if tx == nil {
		return false
	}
	attached := pos.Side(t.Facing)
	if tx.RedstonePower(attached) > 0 {
		return true
	}
	if source, ok := tx.Block(attached).(world.RedstonePowerSource); ok {
		for _, face := range cube.Faces() {
			if redstonePower(source.RedstonePower(attached, tx, face)) > 0 {
				return true
			}
		}
	}
	return false
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
	name = "minecraft:unlit_redstone_torch"
	if t.Lit {
		name = "minecraft:redstone_torch"
	}
	return name, map[string]any{"torch_facing_direction": torchFacingDirection(t.Facing)}
}

func torchFacingDirection(face cube.Face) string {
	switch face {
	case cube.FaceDown:
		return "top"
	case unknownFace:
		return "unknown"
	default:
		return face.String()
	}
}

func allRedstoneTorches() (torches []world.Block) {
	for _, face := range cube.Faces() {
		if face == cube.FaceUp {
			face = unknownFace
		}
		torches = append(torches, RedstoneTorch{Facing: face}, RedstoneTorch{Facing: face, Lit: true})
	}
	return
}
