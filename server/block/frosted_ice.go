package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// FrostedIce ...
type FrostedIce struct {
	solid

	// Age ...
	Age int
}

// Instrument ...
func (FrostedIce) Instrument() sound.Instrument {
	return sound.Chimes()
}

// BreakInfo ...
func (p FrostedIce) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(p))
}

// Friction ...
func (p FrostedIce) Friction() float64 {
	return 0.98
}

// Tick ...
func (fi FrostedIce) Tick(_ int64, pos cube.Pos, w *world.World) {
	// Tick block once every one and a half seconds.
	if rand.Intn(19) != 0 {
		return
	}

	// One of two conditions must be met in order to increase the age of frosted ice.
	// 1) A random number generator with possible values [0, 1, 2] chooses 0.
	// 2) The frosted ice block has fewer than 4 frosted ice blocks immediately surrounding it
	//    AND the light level at the block is greater than 11 minus its age.
	if rand.Intn(2) != 0 && (frostedIce(pos, w) >= 4 || w.Light(pos) <= uint8((11-fi.Age))) {
		return
	}

	fi.Age++

	if fi.Age > 3 {
		w.SetBlock(pos, Water{
			empty{},
			replaceable{},
			true,
			8,
			false,
		}, nil)

		return
	}

	w.SetBlock(pos, fi, nil)
}

func frostedIce(pos cube.Pos, w *world.World) int {
	count := 0

	for offx := -1; offx < 2; offx++ {
		for offz := -1; offz < 2; offz++ {
			if offx == 0 && offz == 0 {
				continue
			}

			offPos := pos.Add(cube.PosFromVec3(mgl64.Vec3{float64(offx), 0, float64(offz)}))

			if _, isFrostedIce := w.Block(offPos).(FrostedIce); isFrostedIce {
				count++
			}
		}
	}

	return count
}

// EncodeItem ...
func (FrostedIce) EncodeItem() (name string, meta int16) {
	return "minecraft:frosted_ice", 0
}

// EncodeBlock ...
func (fi FrostedIce) EncodeBlock() (string, map[string]any) {
	return "minecraft:frosted_ice", map[string]any{"age": int32(fi.Age)}
}

// DecodeNBT ...
func (fi FrostedIce) DecodeNBT(data map[string]any) any {
	fi.Age, _ = data["age"].(int)
	return fi
}

// EncodeNBT ...
func (fi FrostedIce) EncodeNBT() map[string]any {
	return map[string]any{"age": int32(fi.Age)}
}

// allFrostedIce returns all possible states of a frosted ice block.
func allFrostedIce() (b []world.Block) {
	for i := 0; i < 4; i++ {
		b = append(b, FrostedIce{Age: i})
	}
	return
}
