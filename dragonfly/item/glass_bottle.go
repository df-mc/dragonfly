package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/potion"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlassBottle is an item that can hold various liquids.
type GlassBottle struct{}

// UseOnBlock ...
func (g GlassBottle) UseOnBlock(pos world.BlockPos, _ world.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if item_internal.IsWaterSource(w.Block(pos)) {
		ctx.CountSub = 1
		ctx.NewItem = NewStack(Potion{Type: potion.Water()}, 1)
		return true
	}
	return false
}

// EncodeItem ...
func (g GlassBottle) EncodeItem() (id int32, meta int16) {
	return 374, 0
}
