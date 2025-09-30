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

// redstoneTorchBurnoutData stores burnout tracking data indexed by block position.
var (
	redstoneTorchBurnoutData   = make(map[cube.Pos]*burnoutData)
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

// getBurnoutData retrieves or creates burnout tracking data for the given position.
func getBurnoutData(pos cube.Pos) *burnoutData {
	redstoneTorchBurnoutDataMu.RLock()
	data, exists := redstoneTorchBurnoutData[pos]
	redstoneTorchBurnoutDataMu.RUnlock()

	if exists {
		return data
	}

	redstoneTorchBurnoutDataMu.Lock()
	defer redstoneTorchBurnoutDataMu.Unlock()

	if d, ok := redstoneTorchBurnoutData[pos]; ok {
		return d
	}

	data = &burnoutData{
		gameTicks: make([]int64, 0, BurnoutThreshold+1),
		burnedOut: false,
	}
	redstoneTorchBurnoutData[pos] = data
	return data
}

// removeBurnoutData cleans up burnout tracking data for the given position.
func removeBurnoutData(pos cube.Pos) {
	redstoneTorchBurnoutDataMu.Lock()
	defer redstoneTorchBurnoutDataMu.Unlock()
	delete(redstoneTorchBurnoutData, pos)
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
		removeBurnoutData(pos)
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
		getBurnoutData(pos)
		t.RedstoneUpdate(pos, tx)
		updateStrongRedstone(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		removeBurnoutData(pos)
		breakBlock(t, pos, tx)
		updateDirectionalRedstone(pos, tx, t.Facing.Opposite())
	}
}

// RedstoneUpdate ...
func (t RedstoneTorch) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	data := getBurnoutData(pos)
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

// ScheduledTick ...
func (t RedstoneTorch) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	data := getBurnoutData(pos)
	currentTime := int64(tx.World().Time())

	data.mu.RLock()
	isBurnedOut := data.burnedOut
	data.mu.RUnlock()

	// If burned out, ignore scheduled ticks
	if isBurnedOut {
		return
	}

	// Normal state change logic
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

// burnOut puts the redstone torch into burnout state, turning it off.
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
