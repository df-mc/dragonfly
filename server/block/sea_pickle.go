package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
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
func (SeaPickle) canSurvive(pos cube.Pos, w *world.World) bool {
	below := w.Block(pos.Side(cube.FaceDown))
	if !below.Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w) {
		return false
	}
	if emitter, ok := below.(LightDiffuser); ok && emitter.LightDiffusionLevel() != 15 {
		return false
	}
	return true
}

// BoneMeal ...
func (s SeaPickle) BoneMeal(pos cube.Pos, w *world.World) bool {
	if s.Dead {
		return false
	}
	if coral, ok := w.Block(pos.Side(cube.FaceDown)).(CoralBlock); !ok || coral.Dead {
		return false
	}

	if s.AdditionalCount != 3 {
		s.AdditionalCount = 3
		w.SetBlock(pos, s, nil)
	}

	for x := -2; x <= 2; x++ {
		distance := -int(math.Abs(float64(x))) + 2
		for z := -distance; z <= distance; z++ {
			for y := -1; y < 1; y++ {
				if (x == 0 && y == 0 && z == 0) || rand.Intn(6) != 0 {
					continue
				}
				newPos := pos.Add(cube.Pos{x, y, z})

				if _, ok := w.Block(newPos).(Water); !ok {
					continue
				}
				if coral, ok := w.Block(newPos.Side(cube.FaceDown)).(CoralBlock); !ok || coral.Dead {
					continue
				}
				w.SetBlock(newPos, SeaPickle{AdditionalCount: rand.Intn(3) + 1}, nil)
			}
		}
	}

	return true
}

// UseOnBlock ...
func (s SeaPickle) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	if existing, ok := w.Block(pos).(SeaPickle); ok {
		if existing.AdditionalCount >= 3 {
			return false
		}

		existing.AdditionalCount++
		place(w, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(w, pos, face, s)
	if !used {
		return false
	}
	if !s.canSurvive(pos, w) {
		return false
	}

	s.Dead = true
	if liquid, ok := w.Liquid(pos); ok {
		_, ok = liquid.(Water)
		s.Dead = !ok
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s SeaPickle) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !s.canSurvive(pos, w) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
		dropItem(w, item.NewStack(s, s.AdditionalCount+1), pos.Vec3Centre())
		return
	}

	alive := false
	if liquid, ok := w.Liquid(pos); ok {
		_, alive = liquid.(Water)
	}
	if s.Dead == alive {
		s.Dead = !alive
		w.SetBlock(pos, s, nil)
	}
}

// HasLiquidDrops ...
func (SeaPickle) HasLiquidDrops() bool {
	return true
}

// SideClosed ...
func (SeaPickle) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
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
