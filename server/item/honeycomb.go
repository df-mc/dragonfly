package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Honeycomb is an item obtained from bee nests and beehives.
type Honeycomb struct{}

// Dispense waxes the block in front of a dispenser.
func (Honeycomb) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	if !waxBlock(pos.Side(face), pos.Vec3Centre(), tx) {
		return DispenseFailure
	}
	ctx.SubtractFromCount(1)
	return DispenseSuccess
}

// UseOnBlock waxes the block at pos if it supports waxing.
func (Honeycomb) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if !waxBlock(pos, user.Position(), tx) {
		return false
	}
	ctx.SubtractFromCount(1)
	return true
}

func waxBlock(pos cube.Pos, source mgl64.Vec3, tx *world.Tx) bool {
	wa, ok := tx.Block(pos).(waxable)
	if !ok {
		return false
	}
	res, ok := wa.Wax(pos, source)
	if !ok {
		return false
	}
	tx.SetBlock(pos, res, nil)
	tx.PlaySound(pos.Vec3(), sound.SignWaxed{})
	return true
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
