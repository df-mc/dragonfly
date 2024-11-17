package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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

func (c CopperGrate) CanOxidate() bool {
	return !c.Waxed
}

func (c CopperGrate) OxidationLevel() OxidationType {
	return c.Oxidation
}

func (c CopperGrate) WithOxidationLevel(o OxidationType) Oxidizable {
	c.Oxidation = o
	return c
}

func (c CopperGrate) Activate(pos cube.Pos, _ cube.Face, w *world.World, user item.User, _ *item.UseContext) bool {
	var ok bool
	c.Oxidation, c.Waxed, ok = activateOxidizable(pos, w, user, c.Oxidation, c.Waxed)
	if ok {
		w.SetBlock(pos, c, nil)
		return true
	}
	return false
}

func (c CopperGrate) SneakingActivate(pos cube.Pos, face cube.Face, w *world.World, user item.User, ctx *item.UseContext) bool {
	// Sneaking should still trigger axe functionality.
	return c.Activate(pos, face, w, user, ctx)
}

func (c CopperGrate) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	attemptOxidation(pos, w, r, c)
}

// EncodeItem ...
func (c CopperGrate) EncodeItem() (name string, meta int16) {
	name = "copper_grate"
	if c.Oxidation != NormalOxidation() {
		name = c.Oxidation.String() + "_" + name
	}
	if c.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, 0
}

// EncodeBlock ...
func (c CopperGrate) EncodeBlock() (string, map[string]any) {
	name := "copper_grate"
	if c.Oxidation != NormalOxidation() {
		name = c.Oxidation.String() + "_" + name
	}
	if c.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, nil
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
