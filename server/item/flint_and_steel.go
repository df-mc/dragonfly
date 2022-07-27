package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// FlintAndSteel is an item used to light blocks on fire.
type FlintAndSteel struct{}

// MaxCount ...
func (f FlintAndSteel) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (f FlintAndSteel) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 65,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// ignitable represents a block that can be lit by a fire emitter, such as flint and steel.
type ignitable interface {
	// Ignite is called when the block is lit by flint and steel.
	Ignite(pos cube.Pos, w *world.World) bool
}

// UseOnBlock ...
func (f FlintAndSteel) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	ctx.DamageItem(1)
	if l, ok := w.Block(pos).(ignitable); ok && l.Ignite(pos, w) {
		return true
	} else if s := pos.Side(face); w.Block(s.Side(cube.FaceDown)) != air() && w.Block(s) == air() {
		w.PlaySound(s.Vec3Centre(), sound.Ignite{})
		w.SetBlock(s, fire(), nil)
		w.ScheduleBlockUpdate(s, time.Duration(30+rand.Intn(10))*time.Second/20)
		return true
	}
	return false
}

// EncodeItem ...
func (f FlintAndSteel) EncodeItem() (name string, meta int16) {
	return "minecraft:flint_and_steel", 0
}

// air returns an air block.
func air() world.Block {
	a, ok := world.BlockByName("minecraft:air", nil)
	if !ok {
		panic("could not find air block")
	}
	return a
}

// fire returns a fire block.
func fire() world.Block {
	f, ok := world.BlockByName("minecraft:fire", map[string]any{"age": int32(0)})
	if !ok {
		panic("could not find fire block")
	}
	return f
}
