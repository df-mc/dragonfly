package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Suger cane is a plant block that generates naturally near water.
type SugarCane struct {
	empty
	transparent

	// Age is the growth state of suger cane. Values range from 0 to 15.
	Age int
}

// UseOnBlock handles making sure a neighbouring blocks contains water.
func (c SugarCane) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}
	if !c.CanGrowHere(pos, w) {
		return false
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c SugarCane) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !c.CanGrowHere(pos, w) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: c})
	}
}

// RandomTick ...
func (c SugarCane) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if c.Age < 15 {
		c.Age++
	} else if c.Age == 15 {
		c.Age = 0
		if c.CanGrowHere(pos.Side(cube.FaceDown), w) {
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
	w.SetBlock(pos, SugarCane{Age: c.Age}, nil)
}

// logic to check if sugar_cane can live/grow here
func (c SugarCane) CanGrowHere(pos cube.Pos, w *world.World) bool {
	// placed on soil.
	if supportsVegetation(c, w.Block(pos.Sub(cube.Pos{0, 1}))) {
		for _, face := range cube.HorizontalFaces() {
			if _, ok := w.Block(pos.Side(face).Side(cube.FaceDown)).(Water); ok {
				return true
			}
		}
		return false // no water
	}

	// placed on one SugarCane
	_, one := w.Block(pos.Side(cube.FaceDown)).(SugarCane)
	if one && supportsVegetation(c, w.Block(pos.Sub(cube.Pos{0, 2}))) {
		return true
	}

	// placed on two SugarCane
	_, two := w.Block(pos.Side(cube.FaceDown)).(SugarCane)
	if one && two && supportsVegetation(c, w.Block(pos.Sub(cube.Pos{0, 3}))) {
		return true
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

// allSugarCane returns all possible states of a suger cane block.
func allSugarCane() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, SugarCane{Age: i})
	}
	return
}
