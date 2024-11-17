package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
)

// Oxidizable is a block that can naturally oxidise over time, such as copper.
type Oxidizable interface {
	world.Block
	// CanOxidate returns whether the block can oxidate, i.e. if it's not waxed.
	CanOxidate() bool
	// OxidationLevel returns the currently level of oxidation of the block.
	OxidationLevel() OxidationType
	// WithOxidationLevel returns the oxidizable block with the oxidation level passed.
	WithOxidationLevel(OxidationType) Oxidizable
}

// activateOxidizable performs the logic for activating an oxidizable block, returning the updated oxidation
// level and wax state of the block, as well as whether the block was successfully activated. This function
// will not handle the setting of the block if it has been modified.
func activateOxidizable(pos cube.Pos, w *world.World, user item.User, o OxidationType, waxed bool) (OxidationType, bool, bool) {
	mainHand, _ := user.HeldItems()
	// TODO: Immediately return false if holding shield in offhand (https://bugs.mojang.com/browse/MC-270047).
	if _, ok := mainHand.Item().(item.Axe); !ok {
		return o, waxed, false
	} else if waxed {
		w.PlaySound(pos.Vec3Centre(), sound.WaxRemoved{})
		return o, false, true
	}

	if ox, ok := o.Decrease(); ok {
		w.PlaySound(pos.Vec3Centre(), sound.CopperScraped{})
		return ox, false, true
	}
	return o, false, true
}

// attemptOxidation attempts to oxidise the block at the position passed. The details for this logic is
// described on the Minecraft Wiki: https://minecraft.wiki/w/Oxidation.
func attemptOxidation(pos cube.Pos, w *world.World, r *rand.Rand, o Oxidizable) {
	level := o.OxidationLevel()
	if level == OxidizedOxidation() || !o.CanOxidate() {
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

				b, ok := w.Block(nPos).(Oxidizable)
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
	if level == NormalOxidation() {
		chance *= chance * 0.75
	} else {
		chance *= chance
	}
	if r.Float64() < chance {
		level, _ = level.Increase()
		w.SetBlock(pos, o.WithOxidationLevel(level), nil)
	}
}
