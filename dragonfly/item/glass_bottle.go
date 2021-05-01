package item

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/potion"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlassBottle is an item that can hold various liquids.
type GlassBottle struct{}

// UseOnBlock ...
func (g GlassBottle) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b, ok := w.Block(pos).(world.Liquid); ok && item_internal.IsWater(b) && b.LiquidDepth() == 8 {
		ctx.CountSub = 1
		ctx.NewItem = NewStack(Potion{Type: potion.Water()}, 1)
		return true
	}
	return false
}

// EncodeItem ...
func (g GlassBottle) EncodeItem() (id int32, name string, meta int16) {
	return 374, "minecraft:glass_bottle", 0
}
