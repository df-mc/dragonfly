package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperChain is a metallic decoration block.
type CopperChain struct {
	transparent
	sourceWaterDisplacer

	// Axis is the axis which the chain faces.
	Axis cube.Axis
	// Oxidation is the level of oxidation of the copper chain.
	Oxidation OxidationType
	// Waxed bool is whether the copper chain has been waxed with honeycomb.
	Waxed bool
}

// SideClosed ...
func (CopperChain) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// UseOnBlock ...
func (c CopperChain) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	c.Axis = face.Axis()

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c CopperChain) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// Wax waxes the copper chain to stop it from oxidising further.
func (c CopperChain) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

// Strip ...
func (c CopperChain) Strip() (world.Block, world.Sound, bool) {
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
func (c CopperChain) CanOxidate() bool {
	return !c.Waxed
}

// OxidationLevel ...
func (c CopperChain) OxidationLevel() OxidationType {
	return c.Oxidation
}

// WithOxidationLevel ...
func (c CopperChain) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

// RandomTick ...
func (c CopperChain) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// EncodeItem ...
func (c CopperChain) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_chain", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperChain) EncodeBlock() (name string, properties map[string]any) {
	return copperBlockName("copper_chain", c.Oxidation, c.Waxed), map[string]any{"pillar_axis": c.Axis.String()}
}

// Model ...
func (c CopperChain) Model() world.BlockModel {
	return model.Chain{Axis: c.Axis}
}

// allCopperChains ...
func allCopperChains() (chains []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			for _, axis := range cube.Axes() {
				chains = append(chains, CopperChain{Axis: axis, Oxidation: o, Waxed: waxed})
			}
		}
	}
	f(true)
	f(false)
	return
}
