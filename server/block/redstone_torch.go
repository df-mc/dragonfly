package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
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

// HasLiquidDrops returns whether the redstone torch drops its item when flowing liquid breaks it.
func (RedstoneTorch) HasLiquidDrops() bool {
	return true
}

// LightEmissionLevel returns the light level emitted by the redstone torch (7 when lit, 0 when unlit).
func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

// BreakInfo returns information about breaking the redstone torch.
func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		tx.Redstone().ClearTorchBurnout(pos)
		updateTorchRedstone(pos, tx)
	})
}

// UseOnBlock handles the placement of a redstone torch on a block surface.
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
		fallbackFace, ok := findTorchPlacementFace(pos, tx)
		if !ok {
			return false
		}
		face = fallbackFace
	}
	t.Facing = face.Opposite()
	t.Lit = true

	place(tx, pos, t, user, ctx)
	if placed(ctx) {
		// Initialise the freshly placed torch state before propagating its output.
		t.RedstoneUpdate(pos, tx)
		updateTorchRedstone(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick is called when a neighbouring block is updated.
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		tx.Redstone().ClearTorchBurnout(pos)
		breakBlock(t, pos, tx)
		return
	}
	if t.recoverFromBurnout(pos, tx) {
		return
	}
	updateRedstone(pos, tx)
}

// RedstoneUpdate is called when the redstone power state changes nearby. This method ignores burned-out torches and
// schedules state changes for active torches.
func (t RedstoneTorch) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	currentTick := tx.CurrentTick()
	if burnedOut, _ := tx.Redstone().TorchBurnoutStatus(pos, currentTick); burnedOut {
		if t.updateSourceTouchesInput(pos, tx) {
			t.recoverFromBurnout(pos, tx)
		}
		return
	}

	shouldBeLit := t.inputStrength(pos, tx) == 0
	if shouldBeLit == t.Lit {
		return
	}
	tx.Redstone().MarkTorchSelfTriggeredIfActive(pos)
	tx.ScheduleBlockUpdate(pos, t, time.Millisecond*100)
}

// recoverFromBurnout relights a burned-out torch after a real neighbouring block update once its rapid-toggle history
// has expired. Redstone propagation alone may visit calculation-only positions, so it must not recover burned-out
// torches that did not receive an actual neighbour update.
func (RedstoneTorch) recoverFromBurnout(pos cube.Pos, tx *world.Tx) bool {
	torch, ok := redstoneTorchAt(pos, tx)
	if !ok {
		return false
	}

	currentTick := tx.CurrentTick()
	burnedOut, recoverable := tx.Redstone().TorchBurnoutStatus(pos, currentTick)
	if !burnedOut {
		return false
	}
	if !recoverable {
		return true
	}
	tx.Redstone().ClearTorchBurnout(pos)

	torch.Lit = torch.inputStrength(pos, tx) == 0
	tx.SetBlock(pos, torch, nil)
	updateTorchRedstone(pos, tx)
	return true
}

// updateSourceTouchesInput reports whether the current redstone update came from the block the torch is attached to,
// or from a block directly beside it. Dust on top of the attached block can legitimately recover a burnout loop, while
// disconnected dust visited by the broad wire walk should not.
func (t RedstoneTorch) updateSourceTouchesInput(pos cube.Pos, tx *world.Tx) bool {
	source, ok := tx.Redstone().UpdateSource()
	if !ok {
		return false
	}
	inputPos := pos.Side(t.Facing)
	if source == inputPos {
		return true
	}
	for _, face := range cube.Faces() {
		if source == inputPos.Side(face) {
			return true
		}
	}
	return false
}

// ScheduledTick is called when a scheduled block update occurs.
// This method handles state changes and checks for burnout conditions.
func (RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	torch, ok := redstoneTorchAt(pos, tx)
	if !ok {
		return
	}

	currentTick := tx.CurrentTick()
	if burnedOut, _ := tx.Redstone().TorchBurnoutStatus(pos, currentTick); burnedOut {
		return
	}

	shouldBeLit := torch.inputStrength(pos, tx) == 0
	if shouldBeLit == torch.Lit {
		tx.Redstone().PruneTorchBurnout(pos, currentTick)
		return
	}

	if tx.Redstone().RecordTorchToggle(pos, currentTick) {
		torch.burnOut(pos, tx)
		return
	}

	torch.Lit = !torch.Lit
	tx.SetBlock(pos, torch, nil)
	updateTorchRedstone(pos, tx)
}

// burnOut puts the redstone torch into burnout state, turning it off and playing effects.
func (RedstoneTorch) burnOut(pos cube.Pos, tx *world.Tx) {
	torch, ok := redstoneTorchAt(pos, tx)
	if !ok {
		return
	}

	tx.Redstone().BurnOutTorch(pos)
	torch.Lit = false
	tx.PlaySound(pos.Vec3Centre(), sound.Fizz{})
	tx.SetBlock(pos, torch, nil)
	updateTorchRedstone(pos, tx)
}

// redstoneTorchAt returns the current torch at pos. Scheduled redstone updates carry an old block value, so mutation
// paths must reload the live world block before writing torch state back.
func redstoneTorchAt(pos cube.Pos, tx *world.Tx) (RedstoneTorch, bool) {
	t, ok := tx.Block(pos).(RedstoneTorch)
	if !ok {
		tx.Redstone().ClearTorchBurnout(pos)
	}
	return t, ok
}

// updateTorchRedstone updates receivers around the torch and behind the block it strongly powers above.
func updateTorchRedstone(pos cube.Pos, tx *world.Tx) {
	tx.Redstone().WithActiveTorchUpdate(pos, func() {
		updateDirectionalRedstone(pos, tx, cube.FaceUp)
	})
}

// EncodeItem encodes the redstone torch as an item.
func (RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

// EncodeBlock encodes the redstone torch as a block for network transmission.
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
	return true
}

// WeakPower returns the weak redstone power level provided to adjacent blocks.
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if !t.Lit {
		return 0
	}
	if face.Opposite() == t.Facing {
		return 0
	}
	return 15
}

// StrongPower returns the strong redstone power level provided to the block above the torch.
func (t RedstoneTorch) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if t.Lit && face == cube.FaceDown {
		return 15
	}
	return 0
}

// inputStrength returns the redstone power level received by the block the torch is attached to.
func (t RedstoneTorch) inputStrength(pos cube.Pos, tx *world.Tx) int {
	return tx.RedstonePower(pos.Side(t.Facing), t.Facing, true)
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
// It returns the face the torch should be placed on and whether it was found.
func findTorchPlacementFace(pos cube.Pos, tx *world.Tx) (cube.Face, bool) {
	for _, side := range redstoneTorchFallbackSides {
		if tx.Block(pos.Side(side)).Model().FaceSolid(pos.Side(side), side.Opposite(), tx) {
			return side.Opposite(), true
		}
	}
	return 0, false
}

// allRedstoneTorches returns all possible redstone torch block states.
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
