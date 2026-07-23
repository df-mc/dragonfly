package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlassBottle is an item that can hold various liquids.
type GlassBottle struct{}

// Dispense fills the bottle from the block or liquid in front of a dispenser, falling back to ordinary ejection when
// no bottle-filling target is present.
func (GlassBottle) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	front := pos.Side(face)
	b := tx.Block(front)
	filler, fromBlock := b.(bottleFiller)
	if !fromBlock {
		liquid, ok := tx.Liquid(front)
		if !ok {
			return DispenseDefault
		}
		filler, ok = liquid.(bottleFiller)
		if !ok {
			return DispenseDefault
		}
	}
	result, filled, ok := filler.FillBottle()
	if !ok {
		return DispenseDefault
	}
	if fromBlock && result != b {
		tx.SetBlock(front, result, nil)
	}
	ctx.NewItem = filled
	ctx.SubtractFromCount(1)
	return DispenseSuccess
}

// bottleFiller is implemented by blocks that can fill bottles by clicking on them.
type bottleFiller interface {
	// FillBottle fills a GlassBottle by interacting with a block. Blocks that implement this interface return both the
	// block that should be placed in the world after filling the bottle, and the item that was produced as a result of
	// the filling.
	// If the bool returned is false, nothing will happen when using a GlassBottle on the block.
	FillBottle() (world.Block, Stack, bool)
}

// UseOnBlock ...
func (g GlassBottle) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	bl := tx.Block(pos)
	if b, ok := bl.(bottleFiller); ok {
		var res world.Block
		if res, ctx.NewItem, ok = b.FillBottle(); ok {
			ctx.SubtractFromCount(1)
			if res != bl {
				// Some blocks (think a cauldron) change when using a bottle on it.
				tx.SetBlock(pos, res, nil)
			}
			return true
		}
	}
	return false
}

// EncodeItem ...
func (g GlassBottle) EncodeItem() (name string, meta int16) {
	return "minecraft:glass_bottle", 0
}
