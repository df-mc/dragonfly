package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

var (
	_ world.RedstonePowerSource = RedstoneTorch{}
	_ world.RedstonePowerAction = RedstoneTorch{}
	_ world.ScheduledTicker     = RedstoneTorch{}
)

// RedstoneTorch is a non-solid block that emits light and provides a full-strength redstone signal when lit.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block it is attached to.
	Facing cube.Face
	// Lit indicates whether the redstone torch is currently lit and emitting power.
	Lit bool
}

func (RedstoneTorch) HasLiquidDrops() bool {
	return true
}

func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		tx.ClearRedstoneTorchBurnout(pos)
	})
}

func (t RedstoneTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used || face == cube.FaceDown {
		return false
	}
	if _, ok := tx.Block(pos).(world.Liquid); ok {
		return false
	}
	if !tx.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, tx) {
		fallbackFace, ok := findTorchPlacementFace(pos, tx)
		if !ok {
			return false
		}
		face = fallbackFace
	}
	t.Facing = face.Opposite()
	t.Lit = !t.attachmentPowered(pos, tx)

	place(tx, pos, t, user, ctx)
	ok := placed(ctx)
	if ok {
		tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
	}
	return ok
}

// NeighbourUpdateTick breaks unsupported torches and otherwise schedules inverse-state refreshes.
func (t RedstoneTorch) NeighbourUpdateTick(pos, changed cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		tx.ClearRedstoneTorchBurnout(pos)
		breakBlock(t, pos, tx)
		return
	}
	if burnedOut, _ := tx.RedstoneTorchBurnoutStatus(pos); burnedOut {
		if changed != pos && !t.attachmentPowered(pos, tx) {
			tx.ClearRedstoneTorchBurnout(pos)
			tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
		}
		return
	}
	tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
}

// ScheduledTick refreshes the lit state after the torch's one-redstone-tick inversion delay.
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if tx == nil {
		return
	}
	attachmentPowered := t.attachmentPowered(pos, tx)
	if burnedOut, _ := tx.RedstoneTorchBurnoutStatus(pos); burnedOut {
		return
	}
	lit := !attachmentPowered
	if t.Lit != lit {
		if !lit && tx.RecordRedstoneTorchTurnOff(pos) {
			t.Lit = false
			tx.PlaySound(pos.Vec3Centre(), sound.Fizz{})
			tx.SetBlock(pos, t, nil)
			return
		}
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

// RedstonePowerAction schedules the delayed inverse-state refresh after an uncancelled power transition.
func (t RedstoneTorch) RedstonePowerAction(pos cube.Pos, tx *world.Tx, _, _ int) bool {
	return t.RedstonePowerActionUpdate(pos, tx, world.RedstoneUpdate{})
}

// RedstonePowerActionUpdate schedules torch refreshes and keeps burnout from self-recovering through its own loop.
func (t RedstoneTorch) RedstonePowerActionUpdate(pos cube.Pos, tx *world.Tx, update world.RedstoneUpdate) bool {
	if tx == nil {
		return false
	}
	attachmentPowered := t.attachmentPowered(pos, tx)
	if burnedOut, _ := tx.RedstoneTorchBurnoutStatus(pos); burnedOut {
		if attachmentPowered || update.ChangedNeighbour == pos {
			return false
		}
		tx.ClearRedstoneTorchBurnout(pos)
		tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
		return true
	}
	if t.Lit == !attachmentPowered {
		return false
	}
	tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
	return true
}

// attachmentPowered reports whether the block the torch is attached to is powered.
func (t RedstoneTorch) attachmentPowered(pos cube.Pos, tx *world.Tx) bool {
	if tx == nil {
		return false
	}
	attached := pos.Side(t.Facing)
	attachedBlock := tx.Block(attached)
	if source, ok := attachedBlock.(world.RedstonePowerSource); ok {
		for _, face := range cube.Faces() {
			if redstonePower(source.RedstonePower(attached, tx, face)) > 0 {
				return true
			}
		}
	}
	return redstoneConductiveBlock(attached, attachedBlock, tx) && tx.RedstoneConductivePower(attached) > 0
}

func (RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

func (t RedstoneTorch) EncodeBlock() (name string, properties map[string]any) {
	name = "minecraft:unlit_redstone_torch"
	if t.Lit {
		name = "minecraft:redstone_torch"
	}
	var direction string
	switch t.Facing {
	case cube.FaceDown:
		direction = "top"
	case unknownFace:
		direction = "unknown"
	default:
		direction = t.Facing.String()
	}
	return name, map[string]any{"torch_facing_direction": direction}
}

// redstoneTorchFallbackSides lists the faces to check for placing a redstone torch on a non-solid block.
var redstoneTorchFallbackSides = [...]cube.Face{
	cube.FaceSouth,
	cube.FaceWest,
	cube.FaceNorth,
	cube.FaceEast,
	cube.FaceDown,
}

// findTorchPlacementFace finds a valid face for placing a redstone torch on a non-solid block.
func findTorchPlacementFace(pos cube.Pos, tx *world.Tx) (cube.Face, bool) {
	for _, side := range redstoneTorchFallbackSides {
		if tx.Block(pos.Side(side)).Model().FaceSolid(pos.Side(side), side.Opposite(), tx) {
			return side.Opposite(), true
		}
	}
	return 0, false
}

func allRedstoneTorches() (all []world.Block) {
	for _, face := range cube.Faces() {
		if face == cube.FaceUp {
			face = unknownFace
		}
		all = append(all, RedstoneTorch{Facing: face}, RedstoneTorch{Facing: face, Lit: true})
	}
	return
}
