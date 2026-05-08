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

// CopperLantern is a light emitting block.
type CopperLantern struct {
	transparent
	sourceWaterDisplacer

	// Hanging determines if a lantern is hanging off a block.
	Hanging bool
	// Oxidation is the level of oxidation of the copper lantern.
	Oxidation OxidationType
	// Waxed bool is whether the copper lantern has been waxed with honeycomb.
	Waxed bool
}

// Model ...
func (c CopperLantern) Model() world.BlockModel {
	return model.Lantern{Hanging: c.Hanging}
}

// NeighbourUpdateTick ...
func (c CopperLantern) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if c.Hanging {
		up := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(up).(CopperChain); !ok && !tx.Block(up).Model().FaceSolid(up, cube.FaceDown, tx) {
			breakBlock(c, pos, tx)
		}
	} else {
		down := pos.Side(cube.FaceDown)
		if !tx.Block(down).Model().FaceSolid(down, cube.FaceUp, tx) {
			breakBlock(c, pos, tx)
		}
	}
}

// LightEmissionLevel ...
func (CopperLantern) LightEmissionLevel() uint8 {
	return 15
}

// UseOnBlock ...
func (c CopperLantern) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		upPos := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(upPos).(CopperChain); !ok && !tx.Block(upPos).Model().FaceSolid(upPos, cube.FaceDown, tx) {
			face = cube.FaceUp
		}
	}
	if face != cube.FaceDown {
		downPos := pos.Side(cube.FaceDown)
		if !tx.Block(downPos).Model().FaceSolid(downPos, cube.FaceUp, tx) {
			return false
		}
	}
	c.Hanging = face == cube.FaceDown

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// SideClosed ...
func (CopperLantern) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (c CopperLantern) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

// Wax waxes the copper lantern to stop it from oxidising further.
func (c CopperLantern) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

// Strip ...
func (c CopperLantern) Strip() (world.Block, world.Sound, bool) {
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
func (c CopperLantern) CanOxidate() bool {
	return !c.Waxed
}

// OxidationLevel ...
func (c CopperLantern) OxidationLevel() OxidationType {
	return c.Oxidation
}

// WithOxidationLevel ...
func (c CopperLantern) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

// RandomTick ...
func (c CopperLantern) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// EncodeItem ...
func (c CopperLantern) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_lantern", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperLantern) EncodeBlock() (name string, properties map[string]any) {
	return copperBlockName("copper_lantern", c.Oxidation, c.Waxed), map[string]any{"hanging": c.Hanging}
}

// allCopperLanterns ...
func allCopperLanterns() (lanterns []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			lanterns = append(lanterns, CopperLantern{Hanging: false, Oxidation: o, Waxed: waxed})
			lanterns = append(lanterns, CopperLantern{Hanging: true, Oxidation: o, Waxed: waxed})
		}
	}
	f(true)
	f(false)
	return
}
