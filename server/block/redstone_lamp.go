package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// RedstoneLamp is a lamp that lights while powered.
type RedstoneLamp struct {
	solid

	// Lit is true when the lamp is powered and emitting light.
	Lit bool
}

// LightEmissionLevel ...
func (r RedstoneLamp) LightEmissionLevel() uint8 {
	if r.Lit {
		return 15
	}
	return 0
}

// RedstonePowerUpdate updates the lamp's lit state to match its redstone input.
func (r RedstoneLamp) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	lit := power > 0
	if r.Lit == lit {
		return r, false
	}
	r.Lit = lit
	return r, true
}

// BreakInfo ...
func (r RedstoneLamp) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(RedstoneLamp{}))
}

// EncodeItem ...
func (RedstoneLamp) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_lamp", 0
}

// EncodeBlock ...
func (r RedstoneLamp) EncodeBlock() (string, map[string]any) {
	if r.Lit {
		return "minecraft:lit_redstone_lamp", nil
	}
	return "minecraft:redstone_lamp", nil
}

func allRedstoneLamps() []world.Block {
	return []world.Block{RedstoneLamp{}, RedstoneLamp{Lit: true}}
}
