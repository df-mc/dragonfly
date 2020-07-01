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

// UseOnBlock ...
func (s Sponge) UseOnBlock(pos world.BlockPos, face world.Face, v mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	// the sponge is already wet so it cannot consume any more water.
	if !s.Wet {
		// tuple binds a world.BlockPos to its distance from the sponge's position.
		type tuple struct {
			block    world.BlockPos
			distance int32
		}

		queue := make([]tuple, 0)
		queue = append(queue, tuple{pos.Side(face), 0})

		// a sponge can only absorb up to 65 water blocks.
		replaced := 0
		for replaced < 65 {
			if len(queue) == 0 {
				break
			}

			// pop the next tuple entry from the queue.
			next := queue[0]
			queue = queue[1:]

			next.block.Neighbours(func(neighbour world.BlockPos) {
				block := w.Block(neighbour)
				if _, isWater := block.(Water); isWater {
					w.SetBlock(neighbour, Air{})
					replaced++
					if next.distance < 7 {
						queue = append(queue, tuple{neighbour, next.distance + 1})
					}
				} else if _, isAir := block.(Air); isAir {
					if next.distance < 7 {
						queue = append(queue, tuple{neighbour, next.distance + 1})
					}
				}
			})
		}

		// at least one water block has been consumed, so the sponge is no longer dry.
		if replaced > 0 {
			s.Wet = true
		}
	}

	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}