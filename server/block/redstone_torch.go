package block

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// burnoutKey uniquely identifies a torch position within a specific world.
type burnoutKey struct {
	worldName string
	pos       cube.Pos
}

// redstoneTorchBurnoutData stores burnout tracking data indexed by world and block position.
var (
	redstoneTorchBurnoutData   = make(map[burnoutKey]*burnoutData)
	redstoneTorchBurnoutDataMu sync.RWMutex
)

// burnoutData holds the burnout state and state change history for a redstone torch.
type burnoutData struct {
	gameTicks   []int64
	burnedOut   bool
	burnoutTime int64
	mu          sync.RWMutex
}

const (
	// BurnoutThreshold is the maximum number of state changes allowed before burnout occurs.
	BurnoutThreshold = 8
	// BurnoutTimeout is the time window (in game ticks) during which state changes are counted.
	// 60 game ticks equals 3 seconds of game time.
	BurnoutTimeout = 60
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

// getBurnoutData retrieves or creates burnout tracking data for the given position and world.
func getBurnoutData(pos cube.Pos, w *world.World) *burnoutData {
	key := burnoutKey{
		worldName: w.Name(),
		pos:       pos,
	}

	redstoneTorchBurnoutDataMu.RLock()
	data, exists := redstoneTorchBurnoutData[key]
	redstoneTorchBurnoutDataMu.RUnlock()

	if exists {
		return data
	}

	redstoneTorchBurnoutDataMu.Lock()
	defer redstoneTorchBurnoutDataMu.Unlock()

	if d, ok := redstoneTorchBurnoutData[key]; ok {
		return d
	}

	data = &burnoutData{
		gameTicks: make([]int64, 0, BurnoutThreshold+1),
		burnedOut: false,
	}
	redstoneTorchBurnoutData[key] = data
	return data
}

// removeBurnoutData cleans up burnout tracking data for the given position and world.
func removeBurnoutData(pos cube.Pos, w *world.World) {
	key := burnoutKey{
		worldName: w.Name(),
		pos:       pos,
	}

	redstoneTorchBurnoutDataMu.Lock()
	defer redstoneTorchBurnoutDataMu.Unlock()
	delete(redstoneTorchBurnoutData, key)
}

// countStateChange records a new state change with an expiration time.
func (data *burnoutData) countStateChange(currentTime int64, timeout int64) {
	data.mu.Lock()
	defer data.mu.Unlock()

	expirationTime := currentTime + timeout
	data.gameTicks = append(data.gameTicks, expirationTime)
}

// counter returns the number of active non-expired state changes.
func (data *burnoutData) counter(currentTime int64) int {
	data.mu.RLock()
	defer data.mu.RUnlock()

	count := 0
	for _, expirationTime := range data.gameTicks {
		if expirationTime >= currentTime {
			count++
		}
	}
	return count
}

// removeExpired removes expired state change entries.
func (data *burnoutData) removeExpired(currentTime int64) {
	data.mu.Lock()
	defer data.mu.Unlock()

	newGameTicks := make([]int64, 0, len(data.gameTicks))
	for _, expirationTime := range data.gameTicks {
		if expirationTime >= currentTime {
			newGameTicks = append(newGameTicks, expirationTime)
		}
	}
	data.gameTicks = newGameTicks
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
		removeBurnoutData(pos, tx.World())
		updateStrongRedstone(pos, tx)
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
		getBurnoutData(pos, tx.World())
		t.RedstoneUpdate(pos, tx)
		updateStrongRedstone(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick is called when a neighboring block is updated.
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		removeBurnoutData(pos, tx.World())
		breakBlock(t, pos, tx)
		updateDirectionalRedstone(pos, tx, t.Facing.Opposite())
	}
}

// RedstoneUpdate is called when the redstone power state changes nearby.
// This method handles burnout recovery and schedules state changes.
func (t RedstoneTorch) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	data := getBurnoutData(pos, tx.World())
	currentTime := int64(tx.World().Time())

	data.mu.RLock()
	isBurnedOut := data.burnedOut
	data.mu.RUnlock()

	if isBurnedOut {
		counter := data.counter(currentTime)
		if counter < BurnoutThreshold {
			data.mu.Lock()
			data.burnedOut = false
			data.gameTicks = make([]int64, 0, BurnoutThreshold+1)
			data.mu.Unlock()

			shouldBeLit := t.inputStrength(pos, tx) == 0
			t.Lit = shouldBeLit
			tx.SetBlock(pos, t, nil)
			updateStrongRedstone(pos, tx)
		}
		return
	}

	if t.inputStrength(pos, tx) > 0 != t.Lit {
		return
	}
	tx.ScheduleBlockUpdate(pos, t, time.Millisecond*100)
}

// ScheduledTick is called when a scheduled block update occurs.
// This method handles state changes and checks for burnout conditions.
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	data := getBurnoutData(pos, tx.World())
	currentTime := int64(tx.World().Time())

	data.mu.RLock()
	isBurnedOut := data.burnedOut
	data.mu.RUnlock()

	if isBurnedOut {
		return
	}

	if t.inputStrength(pos, tx) > 0 != t.Lit {
		return
	}

	data.countStateChange(currentTime, BurnoutTimeout)

	counter := data.counter(currentTime)
	if counter > BurnoutThreshold {
		t.burnOut(pos, tx, data, currentTime)
		return
	}

	data.removeExpired(currentTime)

	t.Lit = !t.Lit
	tx.SetBlock(pos, t, nil)
	updateStrongRedstone(pos, tx)
}

// burnOut puts the redstone torch into burnout state, turning it off and playing effects.
func (t RedstoneTorch) burnOut(pos cube.Pos, tx *world.Tx, data *burnoutData, currentTime int64) {
	data.mu.Lock()
	data.burnedOut = true
	data.burnoutTime = currentTime
	data.mu.Unlock()

	t.Lit = false
	tx.PlaySound(pos.Vec3Centre(), sound.Fizz{})
	tx.SetBlock(pos, t, nil)
	updateStrongRedstone(pos, tx)
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

// RedstoneSource returns whether the redstone torch is currently providing redstone power.
func (t RedstoneTorch) RedstoneSource() bool {
	return t.Lit
}

// WeakPower returns the weak redstone power level provided to adjacent blocks.
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if !t.Lit {
		return 0
	}
	if face != t.Facing.Opposite() {
		return 15
	}
	return 0
}

// StrongPower returns the strong redstone power level provided to blocks above the torch.
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
