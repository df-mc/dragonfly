package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
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
	for _, b := range []cube.Pos{pos, pos.Side(face)} {
		if l, ok := w.Block(b).(ignitable); ok && l.Ignite(b, w) {
			return true
		}
	}
	return false
}

// EncodeItem ...
func (f FlintAndSteel) EncodeItem() (name string, meta int16) {
	return "minecraft:flint_and_steel", 0
}
