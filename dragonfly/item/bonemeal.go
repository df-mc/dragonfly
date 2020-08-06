package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Bonemeal is an item used to force growth in plants & crops.
type Bonemeal struct{}

// UseOnBlock ...
func (b Bonemeal) UseOnBlock(pos world.BlockPos, _ world.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	ok := item_internal.Bonemeal(pos, w)
	if ok {
		ctx.CountSub = 1
		w.AddParticle(pos.Vec3(), particle.Bonemeal{})
	}
	return ok
}

// EncodeItem ...
func (b Bonemeal) EncodeItem() (id int32, meta int16) {
	return 351, 15
}
