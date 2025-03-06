package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand/v2"
)

// SeaPickle is a small stationary underwater block that emits light, and is typically found in colonies of up to
// four sea pickles.
type SeaPickle struct {
	empty
	transparent
	sourceWaterDisplacer

	// AdditionalCount is the amount of additional sea pickles clustered together.
	AdditionalCount int
	// Dead is whether the sea pickles are not alive. Sea pickles are only considered alive when inside of water. While
	// alive, sea pickles emit light & can be grown with bone meal.
	Dead bool
}

// canSurvive ...
func (SeaPickle) canSurvive(pos cube.Pos, tx *world.Tx) bool {
	below := tx.Block(pos.Side(cube.FaceDown))
	if !below.Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, tx) {
		return false
	}
	if liquid, ok := tx.Liquid(pos); ok {
		if _, ok = liquid.(Water); !ok || liquid.LiquidDepth() != 8 {
			return false
		}
	}
	if emitter, ok := below.(LightDiffuser); ok && emitter.LightDiffusionLevel() != 15 {
		return false
	}
	return true
}

// BoneMeal ...
func (s SeaPickle) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if s.Dead {
		return false
	}
	if coral, ok := tx.Block(pos.Side(cube.FaceDown)).(CoralBlock); !ok || coral.Dead {
		return false
	}

	if s.AdditionalCount != 3 {
		s.AdditionalCount = 3
		tx.SetBlock(pos, s, nil)
	}

	for x := -2; x <= 2; x++ {
		distance := -int(math.Abs(float64(x))) + 2
		for z := -distance; z <= distance; z++ {
			for y := -1; y < 1; y++ {
				if (x == 0 && y == 0 && z == 0) || rand.IntN(6) != 0 {
					continue
				}
				newPos := pos.Add(cube.Pos{x, y, z})

				if _, ok := tx.Block(newPos).(Water); !ok {
					continue
				}
				if coral, ok := tx.Block(newPos.Side(cube.FaceDown)).(CoralBlock); !ok || coral.Dead {
					continue
				}
				tx.SetBlock(newPos, SeaPickle{AdditionalCount: rand.IntN(3) + 1}, nil)
			}
		}
	}

	return true
}

// UseOnBlock ...
func (s SeaPickle) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if existing, ok := tx.Block(pos).(SeaPickle); ok {
		if existing.AdditionalCount >= 3 {
			return false
		}

		existing.AdditionalCount++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	if !s.canSurvive(pos, tx) {
		return false
	}

	s.Dead = true
	if liquid, ok := tx.Liquid(pos); ok {
		_, ok = liquid.(Water)
		s.Dead = !ok
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s SeaPickle) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !s.canSurvive(pos, tx) {
		breakBlock(s, pos, tx)
		return
	}

	alive := false
	if liquid, ok := tx.Liquid(pos); ok {
		_, alive = liquid.(Water)
	}
	if s.Dead == alive {
		s.Dead = !alive
		tx.SetBlock(pos, s, nil)
	}
}

// HasLiquidDrops ...
func (SeaPickle) HasLiquidDrops() bool {
	return true
}

// SideClosed ...
func (SeaPickle) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// LightEmissionLevel ...
func (s SeaPickle) LightEmissionLevel() uint8 {
	if s.Dead {
		return 0
	}
	return uint8(6 + s.AdditionalCount*3)
}

// BreakInfo ...
func (s SeaPickle) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(s, s.AdditionalCount+1)))
}

// FlammabilityInfo ...
func (SeaPickle) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

// SmeltInfo ...
func (SeaPickle) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(item.Dye{Colour: item.ColourLime()}, 1), 0.1)
}

// CompostChance ...
func (SeaPickle) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (SeaPickle) EncodeItem() (name string, meta int16) {
	return "minecraft:sea_pickle", 0
}

// EncodeBlock ...
func (s SeaPickle) EncodeBlock() (string, map[string]any) {
	return "minecraft:sea_pickle", map[string]any{"cluster_count": int32(s.AdditionalCount), "dead_bit": s.Dead}
}

// allSeaPickles ...
func allSeaPickles() (b []world.Block) {
	for i := 0; i <= 3; i++ {
		b = append(b, SeaPickle{AdditionalCount: i})
		b = append(b, SeaPickle{AdditionalCount: i, Dead: true})
	}
	return
}
