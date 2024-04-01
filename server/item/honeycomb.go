package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Honeycomb is an item obtained from bee nests and beehives.
type Honeycomb struct{}

// UseOnBlock handles the logic of using an ink sac on a sign. Glowing ink sacs turn the text of these signs glowing,
// whereas normal ink sacs revert them back to non-glowing text.
func (Honeycomb) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, w *world.World, user User, ctx *UseContext) bool {
	if wa, ok := w.Block(pos).(waxable); ok {
		if res, ok := wa.Wax(pos, user.Position()); ok {
			w.SetBlock(pos, res, nil)
			w.PlaySound(pos.Vec3(), sound.SignWaxed{})
			ctx.SubtractFromCount(1)
			return true
		}
	}
	return false
}

// waxable represents a block that may be waxed.
type waxable interface {
	// Wax uses an ink sac on the block, returning the resulting block and a bool specifying if waxing the block was
	// successful.
	Wax(pos cube.Pos, userPos mgl64.Vec3) (world.Block, bool)
}

// EncodeItem ...
func (Honeycomb) EncodeItem() (name string, meta int16) {
	return "minecraft:honeycomb", 0
}
