package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperBars are blocks that serve a similar purpose to glass panes, but made of copper instead of glass.
type CopperBars struct {
	transparent
	thin
	sourceWaterDisplacer

	// Oxidation is the level of oxidation of the copper bars.
	Oxidation OxidationType
	// Waxed bool is whether the copper bars has been waxed with honeycomb.
	Waxed bool
}

// BreakInfo ...
func (c CopperBars) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// SideClosed ...
func (c CopperBars) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// Wax waxes the copper bars to stop it from oxidising further.
func (c CopperBars) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

// Strip ...
func (c CopperBars) Strip() (world.Block, world.Sound, bool) {
	if c.Waxed {
		c.Waxed = false
		return c, sound.WaxRemoved{}, true
	} else if ot, ok := c.Oxidation.Decrease(); ok {
		c.Oxidation = ot
		return c, sound.CopperScraped{}, true
	}
	return c, nil, false
}

// CanOxidate ...
func (c CopperBars) CanOxidate() bool {
	return !c.Waxed
}

// OxidationLevel ...
func (c CopperBars) OxidationLevel() OxidationType {
	return c.Oxidation
}

// WithOxidationLevel ...
func (c CopperBars) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

// RandomTick ...
func (c CopperBars) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// EncodeItem ...
func (c CopperBars) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_bars", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperBars) EncodeBlock() (name string, properties map[string]any) {
	return copperBlockName("copper_bars", c.Oxidation, c.Waxed), nil
}

// allCopperBars ...
func allCopperBars() (bars []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			bars = append(bars, CopperBars{Oxidation: o, Waxed: waxed})
		}
	}
	f(true)
	f(false)
	return
}
