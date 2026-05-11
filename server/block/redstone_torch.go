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

func (t RedstoneTorch) facing() cube.Face {
	if t.Facing == unknownFace {
		return cube.FaceDown
	}
	return t.Facing
}

func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		tx.Redstone().Torch(pos).ClearBurnout()
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
	facing := t.facing()
	if !tx.Block(pos.Side(facing)).Model().FaceSolid(pos.Side(facing), facing.Opposite(), tx) {
		tx.Redstone().Torch(pos).ClearBurnout()
		breakBlock(t, pos, tx)
		return
	}
	torch := tx.Redstone().Torch(pos)
	if burnedOut, recoverable := torch.BurnoutStatus(); burnedOut {
		if t.recoverBurnout(pos, changed, tx, recoverable, false, t.attachmentPowered(pos, tx)) {
			torch.ClearBurnout()
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
	var ok bool
	if t, ok = redstoneTorchAt(pos, tx); !ok {
		return
	}
	redstone := tx.Redstone()
	torch := redstone.Torch(pos)
	selfTriggered := torch.ConsumeSelfTriggered()
	if burnedOut, _ := torch.BurnoutStatus(); burnedOut {
		return
	}
	attachmentPowered := t.attachmentPowered(pos, tx)
	lit := !attachmentPowered
	if t.Lit != lit {
		if !lit && selfTriggered && torch.RecordTurnOff() {
			t.Lit = false
			tx.PlaySound(pos.Vec3Centre(), sound.Fizz{})
			tx.SetBlock(pos, t, &world.SetOpts{DisableRedstoneUpdates: true})
			redstone.ScheduleUpdate(pos)
			return
		}
		t.Lit = lit
		tx.SetBlock(pos, t, &world.SetOpts{DisableRedstoneUpdates: true})
		redstone.ScheduleUpdate(pos)
	}
}

// redstoneTorchAt returns the live redstone torch at pos. Scheduled tick callers may carry stale block state, so
// mutation paths must reload the world block before writing torch state back.
func redstoneTorchAt(pos cube.Pos, tx *world.Tx) (RedstoneTorch, bool) {
	t, ok := tx.Block(pos).(RedstoneTorch)
	if !ok {
		tx.Redstone().Torch(pos).ClearBurnout()
	}
	return t, ok
}

// RedstonePower emits power from every side except the attached block while lit.
func (t RedstoneTorch) RedstonePower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if t.Lit && face != t.facing() {
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
	torch := tx.Redstone().Torch(pos)
	if burnedOut, recoverable := torch.BurnoutStatus(); burnedOut {
		attachmentPowered := update.NewPower > 0 && t.attachmentPowered(pos, tx)
		if !update.HasChangedNeighbour || update.Cause == world.RedstoneUpdateCauseScheduledTick || !t.recoverBurnout(pos, update.ChangedNeighbour, tx, recoverable, update.ChangedRedstoneRelevant, attachmentPowered) {
			return false
		}
		torch.ClearBurnout()
		tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
		return true
	}
	attachmentPowered := t.attachmentPowered(pos, tx)
	if t.Lit == !attachmentPowered {
		return false
	}
	if attachmentPowered && redstoneTorchSelfTriggered(pos, update) {
		torch.MarkSelfTriggered()
	}
	tx.ScheduleBlockUpdate(pos, t, redstoneTicks(1))
	return true
}

// attachmentPowered reports whether the block the torch is attached to is powered.
func (t RedstoneTorch) attachmentPowered(pos cube.Pos, tx *world.Tx) bool {
	if tx == nil {
		return false
	}
	attached := pos.Side(t.facing())
	attachedBlock := tx.Block(attached)
	if source, ok := attachedBlock.(world.RedstonePowerSource); ok {
		for _, face := range cube.Faces() {
			if world.ClampRedstonePower(source.RedstonePower(attached, tx, face)) > 0 {
				return true
			}
		}
	}
	return redstoneConductiveBlock(attached, attachedBlock, tx) && tx.RedstoneConductivePower(attached) > 0
}

// recoverBurnout reports whether an update should relight a burned-out torch.
func (t RedstoneTorch) recoverBurnout(pos, changed cube.Pos, tx *world.Tx, recoverable, changedRedstoneRelevant, attachmentPowered bool) bool {
	if changed == pos || attachmentPowered {
		return false
	}
	touchesRecoveryArea := t.changeTouchesRecoveryArea(pos, changed, tx)
	if changedRedstoneRelevant {
		return touchesRecoveryArea
	}
	if changed == pos.Side(t.facing()) {
		return true
	}
	return recoverable && touchesRecoveryArea
}

// changeTouchesRecoveryArea reports whether changed is close enough to the torch, its input, or an output wire directly
// beside it to count as a real neighbour update. A dust line extending straight away from the torch is not local.
func (t RedstoneTorch) changeTouchesRecoveryArea(pos, changed cube.Pos, tx *world.Tx) bool {
	inputPos := pos.Side(t.facing())
	if changed == inputPos {
		return true
	}
	for _, face := range cube.Faces() {
		if changed == pos.Side(face) || changed == inputPos.Side(face) {
			return true
		}
	}
	for _, face := range cube.HorizontalFaces() {
		neighbour := pos.Side(face)
		if _, ok := tx.Block(neighbour).(RedstoneWire); !ok {
			continue
		}
		for _, wireFace := range cube.HorizontalFaces() {
			if wireFace == face || wireFace == face.Opposite() {
				continue
			}
			if changed == neighbour.Side(wireFace) {
				return true
			}
		}
	}
	return false
}

// redstoneTorchSelfTriggered reports whether this update came from the torch's own scheduled output propagation.
func redstoneTorchSelfTriggered(pos cube.Pos, update world.RedstoneUpdate) bool {
	if update.Cause != world.RedstoneUpdateCauseScheduledTick {
		return false
	}
	if update.HasSource {
		return update.Source == pos
	}
	return update.HasChangedNeighbour && update.ChangedNeighbour == pos
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
