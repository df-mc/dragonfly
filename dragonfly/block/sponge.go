package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Sponge is a block that can be used to remove water around itself when placed, turning into a wet sponge in the
// process.
type Sponge struct {
	// Wet specifies whether the dry or the wet variant of the block is used.
	Wet bool
}

// BreakInfo ...
func (s Sponge) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.6,
		Drops:    simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sponge) EncodeItem() (id int32, meta int16) {
	if s.Wet {
		meta = 1
	}

	return 19, meta
}

// EncodeBlock ...
func (s Sponge) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Wet {
		return "minecraft:sponge", map[string]interface{}{"sponge_type": "wet"}
	}
	return "minecraft:sponge", map[string]interface{}{"sponge_type": "dry"}
}

// UseOnBlock places the sponge, absorbs nearby water if it's still dry and flags it as wet if any water has been
// absorbed.
func (s Sponge) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}

	place(w, pos, s, user, ctx)

	// the sponge is dry, so it can absorb nearby water.
	if !s.Wet {
		if s.absorbWater(pos, w) > 0 {
			// water has been absorbed, so we flag the sponge as wet.
			s.Wet = true
			w.SetBlock(pos, s)
		}
	}

	return placed(ctx)
}

// NeighbourUpdateTick checks for nearby water flow. If water could be found and the sponge is dry, it will absorb the
// water and be flagged as wet.
func (s Sponge) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	// the sponge is dry, so it can absorb nearby water.
	if !s.Wet {
		if s.absorbWater(pos, w) > 0 {
			// water has been absorbed, so we flag the sponge as wet.
			s.Wet = true
			w.SetBlock(pos, s)
		}
	}
}

// absorbWater replaces water blocks near the sponge by air out to a taxicab geometry of 7 in all directions.
// The maximum for absorbed blocks is 65.
// The returned int specifies the amount of replaced water blocks.
func (s Sponge) absorbWater(pos world.BlockPos, w *world.World) int {
	// distanceToSponge binds a world.BlockPos to its distance from the sponge's position.
	type distanceToSponge struct {
		block    world.BlockPos
		distance int32
	}

	queue := make([]distanceToSponge, 0)
	queue = append(queue, distanceToSponge{pos, 0})

	// a sponge can only absorb up to 65 water blocks.
	replaced := 0
	for replaced < 65 {
		if len(queue) == 0 {
			break
		}

		// pop the next distanceToSponge entry from the queue.
		next := queue[0]
		queue = queue[1:]

		// TODO: absorb water only if it's next to the sponge or connected to it.
		next.block.Neighbours(func(neighbour world.BlockPos) {
			liquid, found := w.Liquid(neighbour)
			if found {
				if _, isWater := liquid.(Water); isWater {
					w.SetLiquid(neighbour, nil)
					replaced++
					if next.distance < 7 {
						queue = append(queue, distanceToSponge{neighbour, next.distance + 1})
					}
				}
			} else if _, isAir := w.Block(neighbour).(Air); isAir {
				if next.distance < 7 {
					queue = append(queue, distanceToSponge{neighbour, next.distance + 1})
				}
			}
		})
	}

	return replaced
}
