package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Scaffolding is a block used for building and climbing, that extends horizontally up to a limited distance from
// a supporting block before it collapses.
type Scaffolding struct {
	transparent
	gravityAffected
	sourceWaterDisplacer

	// Stability is the distance from a supporting block, from 0 to 7. At 7 the scaffolding is unsupported.
	Stability int
	// StabilityCheck is set after the first scheduled tick. An unsupported scaffolding falls if it is not set yet
	// and breaks if it is.
	StabilityCheck bool
}

// NeighbourUpdateTick ...
func (s Scaffolding) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if s.Stability != scaffoldingStability(pos, tx) {
		tx.ScheduleBlockUpdate(pos, s, time.Millisecond*50)
	}
}

// ScheduledTick ...
func (s Scaffolding) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	old := s
	s.Stability = scaffoldingStability(pos, tx)
	if s.Stability == 7 {
		if s.StabilityCheck {
			breakBlock(s, pos, tx)
			return
		}
		if replaceableWith(tx, pos.Side(cube.FaceDown), s) {
			s.fall(s, pos, tx)
			return
		}
	}
	if s.StabilityCheck = true; s != old {
		tx.SetBlock(pos, s, nil)
	}
}

// Landed ...
func (s Scaffolding) Landed(tx *world.Tx, pos cube.Pos) {
	tx.ScheduleBlockUpdate(pos, s, time.Millisecond*100)
}

// UseOnBlock ...
func (s Scaffolding) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, ok := scaffoldingPlacementPos(pos, face, tx, s)
	if !ok {
		return false
	}
	if _, ok := tx.Block(pos).(Lava); ok {
		return false
	}
	s.Stability = scaffoldingStability(pos, tx)
	if s.Stability == 7 && !scaffoldingCanFloat(pos, tx) {
		return false
	}
	// Scaffolding may be placed inside the placing entity, so that a tower can be built while standing on it.
	ctx.IgnoreBBox = true

	place(tx, pos, s, user, ctx)
	if placed(ctx) {
		tx.ScheduleBlockUpdate(pos, s, time.Millisecond*100)
		return true
	}
	return false
}

// scaffoldingPlacementPos returns the position a scaffolding is placed at when clicking pos on face. Clicking the
// bottom of a scaffolding, or the top of one with more scaffolding above it, places it on top of that column.
func scaffoldingPlacementPos(pos cube.Pos, face cube.Face, tx *world.Tx, s Scaffolding) (cube.Pos, bool) {
	if _, ok := tx.Block(pos).(Scaffolding); ok && face == cube.FaceDown {
		pos = scaffoldingColumnTop(pos, tx).Side(cube.FaceUp)
		return pos, replaceableWith(tx, pos, s)
	}
	if resolved, _, used := firstReplaceable(tx, pos, face, s); used {
		return resolved, true
	}
	if target := pos.Side(face); face == cube.FaceUp {
		if _, ok := tx.Block(target).(Scaffolding); ok {
			pos = scaffoldingColumnTop(target, tx).Side(cube.FaceUp)
			return pos, replaceableWith(tx, pos, s)
		}
	}
	return cube.Pos{}, false
}

// scaffoldingColumnTop returns the topmost scaffolding of the column pos is part of.
func scaffoldingColumnTop(pos cube.Pos, tx *world.Tx) cube.Pos {
	for {
		above := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(above).(Scaffolding); !ok {
			return pos
		}
		pos = above
	}
}

// EntityInside ...
func (Scaffolding) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if f, ok := e.(fallDistanceEntity); ok {
		f.ResetFallDistance()
	}
}

// EntityLand ...
func (Scaffolding) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance = 0
	}
}

// SideClosed ...
func (Scaffolding) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (s Scaffolding) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

// FlammabilityInfo ...
func (Scaffolding) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 20, true)
}

// FuelInfo ...
func (Scaffolding) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 5 / 2)
}

// Model ...
func (s Scaffolding) Model() world.BlockModel {
	return model.Scaffolding{Unsupported: s.Stability == 7}
}

// EncodeItem ...
func (Scaffolding) EncodeItem() (name string, meta int16) {
	return "minecraft:scaffolding", 0
}

// EncodeBlock ...
func (s Scaffolding) EncodeBlock() (string, map[string]any) {
	return "minecraft:scaffolding", map[string]any{
		"stability":       int32(s.Stability),
		"stability_check": s.StabilityCheck,
	}
}

// scaffoldingStability returns the Stability a scaffolding block at pos has. A scaffolding on top of another one
// takes its stability, otherwise it is 0 on a sturdy block or one more than the most stable horizontal neighbour.
func scaffoldingStability(pos cube.Pos, tx *world.Tx) int {
	below := pos.Side(cube.FaceDown)
	if s, ok := tx.Block(below).(Scaffolding); ok {
		return s.Stability
	}
	if pos[1] <= tx.Range()[0] || tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		return 0
	}
	dist := 7
	for _, face := range []cube.Face{cube.FaceNorth, cube.FaceEast, cube.FaceSouth, cube.FaceWest} {
		if s, ok := tx.Block(pos.Side(face)).(Scaffolding); ok {
			dist = min(dist, min(s.Stability, 6)+1)
		}
	}
	return dist
}

// scaffoldingCanFloat reports if an unsupported scaffolding may still be placed at pos, which is the case when the
// block below is air or water so that it falls.
func scaffoldingCanFloat(pos cube.Pos, tx *world.Tx) bool {
	switch tx.Block(pos.Side(cube.FaceDown)).(type) {
	case Air, Water:
		return true
	}
	return false
}

// allScaffolding ...
func allScaffolding() (b []world.Block) {
	for stability := 0; stability <= 7; stability++ {
		b = append(b, Scaffolding{Stability: stability})
		b = append(b, Scaffolding{Stability: stability, StabilityCheck: true})
	}
	return
}
