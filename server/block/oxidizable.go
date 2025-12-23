package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Oxidisable is a block that can naturally oxidise over time, such as copper.
type Oxidisable interface {
	world.Block
	// CanOxidate returns whether the block can oxidate, i.e. if it's not waxed.
	CanOxidate() bool
	// OxidationLevel returns the currently level of oxidation of the block.
	OxidationLevel() OxidationType
	// WithOxidationLevel returns the oxidizable block with the oxidation level passed.
	WithOxidationLevel(OxidationType) Oxidisable
}

// attemptOxidation attempts to oxidise the block at the position passed. The details for this logic is
// described on the Minecraft Wiki: https://minecraft.wiki/w/Oxidation.
func attemptOxidation(pos cube.Pos, tx *world.Tx, r *rand.Rand, o Oxidisable) {
	level := o.OxidationLevel()
	if level == OxidisedOxidation() || !o.CanOxidate() {
		return
	} else if r.Float64() > 64.0/1125.0 {
		return
	}

	var all, higher int
	for x := -4; x <= 4; x++ {
		for y := -4; y <= 4; y++ {
			for z := -4; z <= 4; z++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}
				nPos := pos.Add(cube.Pos{x, y, z})
				dist := abs(nPos.X()-pos.X()) + abs(nPos.Y()-pos.Y()) + abs(nPos.Z()-pos.Z())
				if dist > 4 {
					continue
				}

				b, ok := tx.Block(nPos).(Oxidisable)
				if !ok || !b.CanOxidate() {
					continue
				} else if b.OxidationLevel().Uint8() < level.Uint8() {
					return
				}
				all++
				if b.OxidationLevel().Uint8() > level.Uint8() {
					higher++
				}
			}
		}
	}

	chance := float64(higher+1) / float64(all+1)
	if level == UnoxidisedOxidation() {
		chance *= chance * 0.75
	} else {
		chance *= chance
	}
	if r.Float64() < chance {
		level, _ = level.Increase()
		tx.SetBlock(pos, o.WithOxidationLevel(level), nil)
	}
}

// copperBlockName returns the name of a copper block with the given oxidation and waxed status.
func copperBlockName(blockName string, oxidation OxidationType, waxed bool) string {
	name := blockName
	if oxidation != UnoxidisedOxidation() {
		name = oxidation.String() + "_" + name
	}
	if waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name
}
