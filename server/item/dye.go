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
func (d Dye) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if dy, ok := tx.Block(pos).(dyeable); ok {
		if res, ok := dy.Dye(pos, user.Position(), d.Colour); ok {
			tx.SetBlock(pos, res, nil)
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
	Dye(pos cube.Pos, userPos mgl64.Vec3, c Colour) (world.Block, bool)
}

func (d Dye) EncodeItem() (name string, meta int16) {
	return "minecraft:" + d.Colour.String() + "_dye", 0
}
