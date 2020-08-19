package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
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

// UseOnBlock ...
func (f FlintAndSteel) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	ctx.DamageItem(1)
	w.PlaySound(pos.Vec3(), sound.Ignite{})
	if w.Block(pos.Side(face)) == item_internal.Air {
		w.PlaceBlock(pos.Side(face), item_internal.Fire)
		return true
	}
	return false
}

// EncodeItem ...
func (f FlintAndSteel) EncodeItem() (id int32, meta int16) {
	return 259, 0
}
