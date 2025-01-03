package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// SugarCane is a plant block that generates naturally near water.
type SugarCane struct {
	empty
	transparent

	// Age is the growth state of sugar cane. Values range from 0 to 15.
	Age int
}

// UseOnBlock ensures the placement of the block is OK.
func (c SugarCane) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}
	if !c.canGrowHere(pos, tx, true) {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (c SugarCane) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !c.canGrowHere(pos, tx, true) {
		breakBlock(c, pos, tx)
	}
}

// RandomTick ...
func (c SugarCane) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !c.canGrowHere(pos, tx, true) {
		breakBlock(c, pos, tx)
		return
	}
	if c.Age < 15 {
		c.Age++
	} else if c.Age == 15 {
		c.Age = 0
		if c.canGrowHere(pos.Side(cube.FaceDown), tx, false) {
			for y := 1; y < 3; y++ {
				if _, ok := tx.Block(pos.Add(cube.Pos{0, y})).(Air); ok {
					tx.SetBlock(pos.Add(cube.Pos{0, y}), SugarCane{}, nil)
					break
				} else if _, ok := tx.Block(pos.Add(cube.Pos{0, y})).(SugarCane); !ok {
					break
				}
			}
		}
	}
	tx.SetBlock(pos, c, nil)
}

// BoneMeal ...
func (c SugarCane) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	for _, ok := tx.Block(pos.Side(cube.FaceDown)).(SugarCane); ok; _, ok = tx.Block(pos.Side(cube.FaceDown)).(SugarCane) {
		pos = pos.Side(cube.FaceDown)
	}
	if c.canGrowHere(pos.Side(cube.FaceDown), tx, false) {
		for y := 1; y < 3; y++ {
			if _, ok := tx.Block(pos.Add(cube.Pos{0, y})).(Air); ok {
				tx.SetBlock(pos.Add(cube.Pos{0, y}), SugarCane{}, nil)
			}
		}
		return true
	}
	return false
}

// canGrowHere implements logic to check if sugar cane can live/grow here.
func (c SugarCane) canGrowHere(pos cube.Pos, tx *world.Tx, recursive bool) bool {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(SugarCane); ok && recursive {
		return c.canGrowHere(pos.Side(cube.FaceDown), tx, recursive)
	}

	if supportsVegetation(c, tx.Block(pos.Sub(cube.Pos{0, 1}))) {
		for _, face := range cube.HorizontalFaces() {
			if liquid, ok := tx.Liquid(pos.Side(face).Side(cube.FaceDown)); ok {
				if _, ok := liquid.(Water); ok {
					return true
				}
			}
		}
	}
	return false
}

// HasLiquidDrops ...
func (c SugarCane) HasLiquidDrops() bool {
	return true
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
