package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// RedstoneLamp lights while powered.
type RedstoneLamp struct {
	solid

	// Lit reports whether the lamp is on.
	Lit bool
}

// LightEmissionLevel returns the light emitted by the lamp.
func (r RedstoneLamp) LightEmissionLevel() uint8 {
	if r.Lit {
		return 15
	}
	return 0
}

// RedstonePowerUpdate updates the lamp when its power changes.
func (r RedstoneLamp) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	if power == 0 || r.Lit {
		return r, false
	}
	r.Lit = true
	return r, true
}

// RedstonePowerActionUpdate changes the off timer after the power update is accepted.
// It also schedules initially lit lamps with no power, even without a power edge.
func (r RedstoneLamp) RedstonePowerActionUpdate(pos cube.Pos, tx *world.Tx, update world.RedstoneUpdate) {
	if update.NewPower > 0 {
		tx.CancelBlockUpdate(pos, r)
	} else if r.Lit {
		tx.ScheduleBlockUpdate(pos, r, redstoneLampOffDelay)
	}
}

const redstoneLampOffDelay = 4 * time.Second / 20

// ScheduledTick turns off an unpowered lamp.
func (r RedstoneLamp) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !r.Lit || tx.RedstonePower(pos) > 0 {
		return
	}
	r.Lit = false
	tx.SetBlock(pos, r, nil)
}

// BreakInfo returns the lamp's break information.
func (r RedstoneLamp) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(RedstoneLamp{}))
}

// EncodeItem encodes the lamp as an item.
func (RedstoneLamp) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_lamp", 0
}

// EncodeBlock encodes the lamp as a block.
func (r RedstoneLamp) EncodeBlock() (string, map[string]any) {
	if r.Lit {
		return "minecraft:lit_redstone_lamp", nil
	}
	return "minecraft:redstone_lamp", nil
}

func allRedstoneLamps() []world.Block {
	return []world.Block{RedstoneLamp{}, RedstoneLamp{Lit: true}}
}
