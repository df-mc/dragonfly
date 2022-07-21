package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Dye is an item that comes in 16 colours which allows you to colour blocks like concrete and sheep.
type Dye struct {
	// Colour is the colour of the dye.
	Colour Colour
}

// UseOnBlock implements the colouring behaviour of signs.
func (d Dye) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if dy, ok := w.Block(pos).(dyeable); ok {
		if res, ok := dy.Dye(d.Colour); ok {
			w.SetBlock(pos, res, nil)
			ctx.SubtractFromCount(1)
			return true
		}
	}
	return false
}

// dyeable represents a block that may be dyed by clicking it with a dye item.
type dyeable interface {
	// Dye uses a dye with the Colour passed on the block. The resulting block is returned. A bool is returned to
	// indicate if dyeing the block was successful.
	Dye(c Colour) (world.Block, bool)
}

// AllDyes returns all 16 dye items
func AllDyes() []world.Item {
	b := make([]world.Item, 0, 16)
	for _, c := range Colours() {
		b = append(b, Dye{Colour: c})
	}
	return b
}

// EncodeItem ...
func (d Dye) EncodeItem() (name string, meta int16) {
	if d.Colour.String() == "silver" {
		return "minecraft:light_gray_dye", 0
	}
	return "minecraft:" + d.Colour.String() + "_dye", 0
}
