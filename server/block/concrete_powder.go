package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// ConcretePowder is a gravity affected block that comes in 16 different colours. When interacting with water,
// it becomes concrete.
type ConcretePowder struct {
	gravityAffected
	solid
	snare

	// Colour is the colour of the concrete powder.
	Colour item.Colour
}

// Solidifies ...
func (c ConcretePowder) Solidifies(pos cube.Pos, w *world.World) bool {
	_, water := w.Block(pos).(Water)
	return water
}

// NeighbourUpdateTick ...
func (c ConcretePowder) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	for i := cube.Face(0); i < 6; i++ {
		if _, ok := w.Block(pos.Side(i)).(Water); ok {
			w.SetBlock(pos, Concrete{Colour: c.Colour}, nil)
			return
		}
	}
	c.fall(c, pos, w)
}

// BreakInfo ...
func (c ConcretePowder) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(c))
}

// EncodeItem ...
func (c ConcretePowder) EncodeItem() (name string, meta int16) {
	return "minecraft:concrete_powder", int16(c.Colour.Uint8())
}

// EncodeBlock ...
func (c ConcretePowder) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:concrete_powder", map[string]any{"color": c.Colour.String()}
}

// allConcretePowder returns concrete powder with all possible colours.
func allConcretePowder() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range item.Colours() {
		b = append(b, ConcretePowder{Colour: c})
	}
	return b
}
