package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Sponge is a block that can be used to remove water around itself when placed, turning into a wet sponge in the
// process.
type Sponge struct {
	solid

	// Wet specifies whether the dry or the wet variant of the block is used.
	Wet bool
}

// BreakInfo ...
func (s Sponge) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, nothingEffective, oneOf(s))
}

// EncodeItem ...
func (s Sponge) EncodeItem() (name string, meta int16) {
	if s.Wet {
		meta = 1
	}
	return "minecraft:sponge", meta
}

// EncodeBlock ...
func (s Sponge) EncodeBlock() (string, map[string]any) {
	if s.Wet {
		return "minecraft:sponge", map[string]any{"sponge_type": "wet"}
	}
	return "minecraft:sponge", map[string]any{"sponge_type": "dry"}
}

// UseOnBlock places the sponge, absorbs nearby water if it's still dry and flags it as wet if any water has been
// absorbed.
func (s Sponge) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick checks for nearby water flow. If water could be found and the sponge is dry, it will absorb the
// water and be flagged as wet.
func (s Sponge) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	// The sponge is dry, so it can absorb nearby water.
	if !s.Wet {
		if s.absorbWater(pos, w) > 0 {
			// Water has been absorbed, so we flag the sponge as wet.
			s.setWet(pos, w)
		}
	}
}

// setWet flags a sponge as wet. It replaces the block at pos by a wet sponge block and displays a block break
// particle at the sponge's position with an offset of 0.5 on each axis.
func (s Sponge) setWet(pos cube.Pos, w *world.World) {
	s.Wet = true
	w.SetBlock(pos, s, nil)
	w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: Water{Depth: 1}})
}

// absorbWater replaces water blocks near the sponge by air out to a taxicab geometry of 7 in all directions.
// The maximum for absorbed blocks is 65.
// The returned int specifies the amount of replaced water blocks.
func (s Sponge) absorbWater(pos cube.Pos, w *world.World) int {
	// distanceToSponge binds a world.Pos to its distance from the sponge's position.
	type distanceToSponge struct {
		block    cube.Pos
		distance int32
	}

	queue := make([]distanceToSponge, 0)
	queue = append(queue, distanceToSponge{pos, 0})

	// A sponge can only absorb up to 65 water blocks.
	replaced := 0
	for replaced < 65 {
		if len(queue) == 0 {
			break
		}

		// Pop the next distanceToSponge entry from the queue.
		next := queue[0]
		queue = queue[1:]

		next.block.Neighbours(func(neighbour cube.Pos) {
			liquid, found := w.Liquid(neighbour)
			if found {
				if _, isWater := liquid.(Water); isWater {
					w.SetLiquid(neighbour, nil)
					replaced++
					if next.distance < 7 {
						queue = append(queue, distanceToSponge{neighbour, next.distance + 1})
					}
				}
			}
		}, w.Range())
	}

	return replaced
}
