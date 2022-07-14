package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// InkSac is an item dropped by a squid upon death used to create black dye, dark prismarine and book and quill. The
// glowing variant, obtained by killing a glow squid, may be used to cause sign text to light up.
type InkSac struct {
	// Glowing specifies if the ink sac is that of a glow squid. If true, it may be used on a sign to light up its text.
	Glowing bool
}

// UseOnBlock handles the logic of using an ink sac on a sign. Glowing ink sacs turn the text of these signs glowing,
// whereas normal ink sacs revert them back to non-glowing text.
func (i InkSac) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if in, ok := w.Block(pos).(inkable); ok {
		if res, ok := in.Ink(i.Glowing); ok {
			w.SetBlock(pos, res, nil)
			ctx.SubtractFromCount(1)
			return true
		}
	}
	return false
}

// inkable represents a block that may be inked, either glowing or reverted from glowing, by using a (glow) ink sac
// on it.
type inkable interface {
	// Ink uses an ink sac on the block, returning the resulting block and a bool specifying if inking the block was
	// successful.
	Ink(glowing bool) (world.Block, bool)
}

// EncodeItem ...
func (i InkSac) EncodeItem() (name string, meta int16) {
	if i.Glowing {
		return "minecraft:glow_ink_sac", 0
	}
	return "minecraft:ink_sac", 0
}
