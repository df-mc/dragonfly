package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperGolemStatue is the result of a copper golem fully oxidizing and petrifying into a block.
// It can be posed in four different ways.
type CopperGolemStatue struct {
	transparent
	sourceWaterDisplacer
	solid

	// Facing is the direction the copper golem statue is facing.
	Facing cube.Direction
	// Pose is the pose of the copper golem statue.
	Pose CopperGolemPose
	// Oxidation is the level of oxidation of the copper lantern.
	Oxidation OxidationType
	// Waxed bool is whether the copper lantern has been waxed with honeycomb.
	Waxed bool
}

// BreakInfo ...
func (c CopperGolemStatue) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// Activate ...
func (c CopperGolemStatue) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	held, _ := u.HeldItems()
	if !held.Empty() {
		// copper golems can't be activated while holding an item.
		return false
	}
	poses := CopperGolemPoses()
	nextIndex := int(c.Pose.Uint8()) + 1
	if nextIndex >= len(poses) {
		nextIndex = 0
	}
	c.Pose = poses[nextIndex]
	tx.SetBlock(pos, c, nil)
	return true
}

// UseOnBlock ...
func (c CopperGolemStatue) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}

	c.Facing = user.Rotation().Direction().Opposite()

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// Wax waxes the copper lantern to stop it from oxidising further.
func (c CopperGolemStatue) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if c.Waxed {
		return c, false
	}
	c.Waxed = true
	return c, true
}

// Strip ...
func (c CopperGolemStatue) Strip() (world.Block, world.Sound, bool) {
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
func (c CopperGolemStatue) CanOxidate() bool {
	return !c.Waxed
}

// OxidationLevel ...
func (c CopperGolemStatue) OxidationLevel() OxidationType {
	return c.Oxidation
}

// WithOxidationLevel ...
func (c CopperGolemStatue) WithOxidationLevel(o OxidationType) Oxidisable {
	c.Oxidation = o
	return c
}

// RandomTick ...
func (c CopperGolemStatue) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, c)
}

// DecodeNBT ...
func (c CopperGolemStatue) DecodeNBT(data map[string]any) any {
	c.Pose = CopperGolemPose{pose(nbtconv.Int32(data, "Pose"))}
	return c
}

// EncodeNBT ...
func (c CopperGolemStatue) EncodeNBT() map[string]any {
	return map[string]any{
		"Pose": int32(c.Pose.Uint8()),
		"id":   "CopperGolemStatue",
	}
}

// EncodeItem ...
func (c CopperGolemStatue) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_golem_statue", c.Oxidation, c.Waxed), 0
}

// EncodeBlock ...
func (c CopperGolemStatue) EncodeBlock() (string, map[string]any) {
	return copperBlockName("copper_golem_statue", c.Oxidation, c.Waxed), map[string]any{"minecraft:cardinal_direction": c.Facing.String()}
}

// allCopperGolemStatues ...
func allCopperGolemStatues() (golems []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			for _, direction := range cube.Directions() {
				golems = append(golems, CopperGolemStatue{Facing: direction, Oxidation: o, Waxed: waxed})
			}
		}
	}
	f(true)
	f(false)
	return
}
