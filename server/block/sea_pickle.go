package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

// SeaPickle is a small stationary underwater block that emits light, and is typically found in colonies of up to
// four sea pickles.
type SeaPickle struct {
	empty
	transparent

	// ClusterCount is the amount of additional sea pickles clustered together.
	ClusterCount int
	// Alive is whether the sea pickle is alive.
	Alive bool
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
	if !s.Alive {
		return false
	}
	if coral, ok := w.Block(pos.Side(cube.FaceDown)).(CoralBlock); !ok || coral.Dead {
		return false
	}

	if s.ClusterCount != 3 {
		s.ClusterCount = 3
		w.PlaceBlock(pos, s)
	}

	for x := -2; x <= 2; x++ {
		distance := -int(math.Abs(float64(x))) + 2
		for z := 0 - distance; z <= 0+distance; z++ {
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
				w.PlaceBlock(newPos, SeaPickle{ClusterCount: rand.Intn(3) + 1, Alive: true})
			}
		}
	}

	return true
}

// UseOnBlock ...
func (s SeaPickle) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	if existing, ok := w.Block(pos).(SeaPickle); ok {
		if existing.ClusterCount >= 3 {
			return false
		}

		existing.ClusterCount++
		w.PlaceBlock(pos, existing)
		ctx.CountSub = 1
		return true
	}

	pos, _, used := firstReplaceable(w, pos, face, s)
	if !used {
		return false
	}
	if !s.canSurvive(pos, w) {
		return false
	}

	if liquid, ok := w.Liquid(pos); ok {
		_, s.Alive = liquid.(Water)
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (s SeaPickle) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !s.canSurvive(pos, w) {
		w.BreakBlock(pos)
		return
	}

	alive := false
	if liquid, ok := w.Liquid(pos); ok {
		_, alive = liquid.(Water)
	}
	if s.Alive != alive {
		s.Alive = alive
		w.PlaceBlock(pos, s)
	}
}

// HasLiquidDrops ...
func (SeaPickle) HasLiquidDrops() bool {
	return true
}

// CanDisplace ...
func (SeaPickle) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (SeaPickle) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// LightEmissionLevel ...
func (s SeaPickle) LightEmissionLevel() uint8 {
	if s.Alive {
		return uint8(6 + s.ClusterCount*3)
	}
	return 0
}

// BreakInfo ...
func (s SeaPickle) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(s, s.ClusterCount+1)))
}

// EncodeItem ...
func (SeaPickle) EncodeItem() (name string, meta int16) {
	return "minecraft:sea_pickle", 0
}

// EncodeBlock ...
func (s SeaPickle) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:sea_pickle", map[string]interface{}{"cluster_count": int32(s.ClusterCount), "dead_bit": !s.Alive}
}

// allSeaPickles ...
func allSeaPickles() (b []world.Block) {
	for i := 0; i <= 3; i++ {
		b = append(b, SeaPickle{ClusterCount: i})
		b = append(b, SeaPickle{ClusterCount: i, Alive: true})
	}
	return
}
