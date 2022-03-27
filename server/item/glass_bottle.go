package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlassBottle is an item that can hold various liquids.
type GlassBottle struct{}

// bottleFiller is implemented by blocks that can fill bottles by clicking on them.
type bottleFiller interface {
	// FillBottle fills a GlassBottle by interacting with a block. Blocks that implement this interface return both the
	// block that should be placed in the world after filling the bottle, and the item that was produced as a result of
	// the filling.
	// If the bool returned is false, nothing will happen when using a GlassBottle on the block.
	FillBottle() (world.Block, Stack, bool)
}

// UseOnBlock ...
func (g GlassBottle) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	bl := w.Block(pos)
	if b, ok := bl.(bottleFiller); ok {
		var res world.Block
		if res, ctx.NewItem, ok = b.FillBottle(); ok {
			ctx.CountSub = 1
			if res != bl {
				// Some blocks (think a cauldron) change when using a bottle on it.
				w.SetBlock(pos, res, nil)
			}
		}
	}
	return false
}

// EncodeItem ...
func (g GlassBottle) EncodeItem() (name string, meta int16) {
	return "minecraft:glass_bottle", 0
}
