package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Copper is a solid block commonly found in deserts and beaches underneath sand.
type Copper struct {
	solid
	bassDrum

	// Type is the type of copper of the block.
	Type CopperType
	// Oxidation is the level of oxidation of the copper block.
	Oxidation OxidationType
	// Waxed bool is whether the copper block has been waxed with honeycomb.
	Waxed bool
}

// BreakInfo ...
func (c Copper) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

// Wax waxes the copper block to stop it from oxidising further.
func (c Copper) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

func (c Copper) CanOxidate() bool {
	return !c.Waxed
}

func (c Copper) OxidationLevel() OxidationType {
	return c.Oxidation
}

func (c Copper) WithOxidationLevel(o OxidationType) Oxidizable {
	c.Oxidation = o
	return c
}

func (c Copper) Activate(pos cube.Pos, _ cube.Face, w *world.World, user item.User, _ *item.UseContext) bool {
	var ok bool
	c.Oxidation, c.Waxed, ok = activateOxidizable(pos, w, user, c.Oxidation, c.Waxed)
	if ok {
		w.SetBlock(pos, c, nil)
		return true
	}
	return false
}

func (c Copper) SneakingActivate(pos cube.Pos, face cube.Face, w *world.World, user item.User, ctx *item.UseContext) bool {
	// Sneaking should still trigger axe functionality.
	return c.Activate(pos, face, w, user, ctx)
}

func (c Copper) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	attemptOxidation(pos, w, r, c)
}

// EncodeItem ...
func (c Copper) EncodeItem() (name string, meta int16) {
	if c.Type == NormalCopper() && c.Oxidation == NormalOxidation() && !c.Waxed {
		return "minecraft:copper_block", 0
	}
	name = "copper"
	if c.Type != NormalCopper() {
		name = c.Type.String() + "_" + name
	}
	if c.Oxidation != NormalOxidation() {
		name = c.Oxidation.String() + "_" + name
	}
	if c.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, 0
}

// EncodeBlock ...
func (c Copper) EncodeBlock() (string, map[string]any) {
	if c.Type == NormalCopper() && c.Oxidation == NormalOxidation() && !c.Waxed {
		return "minecraft:copper_block", nil
	}
	name := "copper"
	if c.Type != NormalCopper() {
		name = c.Type.String() + "_" + name
	}
	if c.Oxidation != NormalOxidation() {
		name = c.Oxidation.String() + "_" + name
	}
	if c.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, nil
}

// allCopper returns a list of all copper block variants.
func allCopper() (c []world.Block) {
	f := func(waxed bool) {
		for _, t := range CopperTypes() {
			for _, o := range OxidationTypes() {
				c = append(c, Copper{Type: t, Oxidation: o, Waxed: waxed})
			}
		}
	}
	f(true)
	f(false)
	return
}
