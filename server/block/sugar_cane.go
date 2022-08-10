package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// SugarCane is a plant block that generates naturally near water.
type SugarCane struct {
	empty
	transparent

	// Age is the growth state of sugar cane. Values range from 0 to 15.
	Age int
}

// UseOnBlock ensures the placement of the block is OK.
func (c SugarCane) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}
	if !c.canGrowHere(pos, w, true) {
		return false
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c SugarCane) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !c.canGrowHere(pos, w, true) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: c})
		dropItem(w, item.NewStack(c, 1), pos.Vec3Centre())
	}
}

// RandomTick ...
func (c SugarCane) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if c.Age < 15 {
		c.Age++
	} else if c.Age == 15 {
		c.Age = 0
		if c.canGrowHere(pos.Side(cube.FaceDown), w, false) {
			for y := 1; y < 3; y++ {
				if _, ok := w.Block(pos.Add(cube.Pos{0, y})).(Air); ok {
					w.SetBlock(pos.Add(cube.Pos{0, y}), SugarCane{Age: 0}, nil)
					break
				} else if _, ok := w.Block(pos.Add(cube.Pos{0, y})).(SugarCane); !ok {
					break
				}
			}
		}
	}
	w.SetBlock(pos, c, nil)
}

// canGrowHere implements logic to check if sugar cane can live/grow here.
func (c SugarCane) canGrowHere(pos cube.Pos, w *world.World, recursive bool) bool {
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(SugarCane); ok && recursive {
		return c.canGrowHere(pos.Side(cube.FaceDown), w, recursive)
	}

	if supportsVegetation(c, w.Block(pos.Sub(cube.Pos{0, 1}))) {
		for _, face := range cube.HorizontalFaces() {
			if _, ok := w.Block(pos.Side(face).Side(cube.FaceDown)).(Water); ok {
				return true
			}
		}
	}
	return false
}

// BreakInfo ...
func (c SugarCane) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(c))
}

// EncodeItem ...
func (c SugarCane) EncodeItem() (name string, meta int16) {
	return "minecraft:sugar_cane", 0
}

// EncodeBlock ...
func (c SugarCane) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:reeds", map[string]any{"age": int32(c.Age)}
}

// allSugarCane returns all possible states of a sugar cane block.
func allSugarCane() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, SugarCane{Age: i})
	}
	return
}
