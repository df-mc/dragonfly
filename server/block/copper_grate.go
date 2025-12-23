package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperGrate is a solid block commonly found in deserts and beaches underneath sand.
type CopperGrate struct {
	sourceWaterDisplacer
	solid
	transparent
	bassDrum

	// Oxidation is the level of oxidation of the copper grate.
	Oxidation OxidationType
	// Waxed bool is whether the copper grate has been waxed with honeycomb.
	Waxed bool
}

// BreakInfo ...
func (c CopperGrate) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// Wax waxes the copper grate to stop it from oxidising further.
func (c CopperGrate) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

func (c CopperGrate) Strip() (world.Block, world.Sound, bool) {
	if c.Waxed {
		c.Waxed = false
		return c, sound.WaxRemoved{}, true
	} else if ot, ok := c.Oxidation.Decrease(); ok {
		c.Oxidation = ot
		return c, sound.CopperScraped{}, true
	}
	return c, nil, false
}

func (c CopperGrate) CanOxidate() bool {
	return !c.Waxed
}

func (c CopperGrate) OxidationLevel() OxidationType {
	return c.Oxidation
}

func (c CopperGrate) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

func (c CopperGrate) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// EncodeItem ...
func (c CopperGrate) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_grate", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperGrate) EncodeBlock() (string, map[string]any) {
	return copperBlockName("copper_grate", c.Oxidation, c.Waxed), nil
}

// allCopperGrates returns a list of all copper grate variants.
func allCopperGrates() (c []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			c = append(c, CopperGrate{Oxidation: o, Waxed: waxed})
		}
	}
	f(true)
	f(false)
	return
}
