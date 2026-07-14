package block

import (
	"math/rand/v2"

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

// RedstonePowerUpdate lights the lamp as soon as it is powered. Turning off
// is delayed by two redstone ticks, keeping the lamp lit through short pulses.
func (r RedstoneLamp) RedstonePowerUpdate(pos cube.Pos, tx *world.Tx, power int) (world.Block, bool) {
	if power > 0 {
		if r.Lit {
			return r, false
		}
		r.Lit = true
		return r, true
	}
	if r.Lit {
		tx.ScheduleBlockUpdate(pos, r, redstoneTicks(2))
	}
	return r, false
}

// ScheduledTick turns the lamp off if it is still unpowered.
func (r RedstoneLamp) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !r.Lit || tx.RedstonePower(pos) > 0 {
		return
	}
	r.Lit = false
	tx.SetBlock(pos, r, nil)
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
