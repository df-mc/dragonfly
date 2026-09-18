package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// StrawBed is a bed crafted from hay bales that a player can sleep in without resetting their spawn point.
// It breaks once its sleeper wakes up, and destroys itself instead of exploding outside the Overworld.
type StrawBed struct {
	transparent
	sourceWaterDisplacer

	// Facing is the direction that the straw bed is facing.
	Facing cube.Direction
	// Head is true if the straw bed is the head side.
	Head bool
	// Occupied is true while an entity is sleeping in the straw bed.
	Occupied bool
	// Sleeper is the user that is using the straw bed. It is only set for the Head part of the bed.
	Sleeper *world.EntityHandle
}

// MaxCount returns 16.
func (StrawBed) MaxCount() int {
	return 16
}

// Model ...
func (StrawBed) Model() world.BlockModel {
	return model.Bed{}
}

// SideClosed ...
func (StrawBed) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FlammabilityInfo ...
func (StrawBed) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 20, true)
}

// BreakInfo ...
func (s StrawBed) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, hoeEffective, oneOf(s)).withBlastResistance(0.2).
		withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
			headSide, _, ok := s.head(pos, tx)
			if !ok {
				return
			}
			h := headSide.Sleeper
			if h == nil {
				return
			}
			ent, ok := h.Entity(tx)
			if !ok {
				return
			}
			if sleeper, ok := ent.(world.Sleeper); ok {
				sleeper.Wake()
			}
		})
}

// UseOnBlock ...
func (s StrawBed) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	if pos, _, used = firstReplaceable(tx, pos, face, s); !used {
		return
	}
	if !supportedFromBelow(pos, tx) {
		return
	}

	s.Facing = user.Rotation().Direction()

	side, sidePos := s, pos.Side(s.Facing.Face())
	side.Head = true

	if !replaceableWith(tx, sidePos, side) {
		return
	}
	if !supportedFromBelow(sidePos, tx) {
		return
	}

	ctx.IgnoreBBox = true
	place(tx, sidePos, side, user, ctx)
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// Activate ...
func (s StrawBed) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	sleeper, ok := u.(world.Sleeper)
	if !ok {
		return false
	}

	w := tx.World()
	if w.Dimension() != world.Overworld {
		// A straw bed does not explode outside the Overworld: it breaks as if its sleeper had woken up.
		s.StopSleeping(pos, tx)
		return true
	}

	_, sidePos, ok := s.side(pos, tx)
	if !ok {
		return false
	}

	userPos := sleeper.Position()
	if sidePos.Vec3Middle().Sub(userPos).Len() > 2 && pos.Vec3Middle().Sub(userPos).Len() > 2 {
		sleeper.Messaget(chat.MessageBedTooFar)
		return true
	}

	headSide, headPos, ok := s.head(pos, tx)
	if !ok {
		return false
	}
	if _, ok = tx.Liquid(headPos); ok {
		return false
	}

	time := w.Time() % world.TimeFull
	if !tx.Thundering() {
		if !tx.Raining() && (time <= world.TimeSleep || time >= world.TimeWake) {
			sleeper.Messaget(chat.MessageNoSleep)
			return true
		}
		if time <= world.TimeSleepWithRain || time >= world.TimeWakeWithRain {
			sleeper.Messaget(chat.MessageNoSleep)
			return true
		}
	}
	if headSide.Sleeper != nil {
		sleeper.Messaget(chat.MessageBedIsOccupied)
		return true
	}

	sleeper.Sleep(headPos)
	return true
}

// SleepingEntity ...
func (s StrawBed) SleepingEntity() *world.EntityHandle {
	return s.Sleeper
}

// StartSleeping ...
func (s StrawBed) StartSleeping(pos cube.Pos, tx *world.Tx, e *world.EntityHandle) {
	if side, sidePos, ok := s.side(pos, tx); ok {
		side.Occupied = true
		tx.SetBlock(sidePos, side, nil)
	}
	s.Sleeper, s.Occupied = e, true
	tx.SetBlock(pos, s, nil)
}

// StopSleeping breaks the straw bed: it is consumed by being slept in. Both halves are cleared directly
// rather than broken, so that the bed's own leave sound plays instead of the block breaking sound.
func (s StrawBed) StopSleeping(pos cube.Pos, tx *world.Tx) {
	if _, sidePos, ok := s.side(pos, tx); ok {
		tx.SetBlock(sidePos, nil, nil)
	}
	tx.SetBlock(pos, nil, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Custom{Name: "block.straw_bed.break_leave", Volume: 1, Pitch: 1})
}

// EntityLand ...
func (StrawBed) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance *= 0.5
	}
	if v, ok := e.(velocityEntity); ok {
		vel := v.Velocity()
		vel[1] = vel[1] * -2 / 3
		v.SetVelocity(vel)
	}
}

// NeighbourUpdateTick ...
func (s StrawBed) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, _, ok := s.side(pos, tx); !ok {
		breakBlockNoDrops(s, pos, tx)
	}
}

// EncodeItem ...
func (StrawBed) EncodeItem() (name string, meta int16) {
	return "minecraft:straw_bed", 0
}

// EncodeBlock ...
func (s StrawBed) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:straw_bed", map[string]any{
		"minecraft:cardinal_direction": s.Facing.String(),
		"occupied_bit":                 boolByte(s.Occupied),
		"head_piece_bit":               boolByte(s.Head),
	}
}

// head returns the head side of the straw bed. If neither side is a head side, the third return value is false.
func (s StrawBed) head(pos cube.Pos, tx *world.Tx) (StrawBed, cube.Pos, bool) {
	headSide, headPos, ok := s.side(pos, tx)
	if !ok {
		return StrawBed{}, cube.Pos{}, false
	}
	if s.Head {
		return s, pos, true
	}
	return headSide, headPos, true
}

// side returns the other side of the straw bed. If the other side is not a straw bed, the third return
// value is false.
func (s StrawBed) side(pos cube.Pos, tx *world.Tx) (StrawBed, cube.Pos, bool) {
	sidePos := bedSide(pos, s.Facing, s.Head)
	o, ok := tx.Block(sidePos).(StrawBed)
	return o, sidePos, ok
}

// allStrawBeds returns all possible straw beds.
func allStrawBeds() (beds []world.Block) {
	for _, d := range cube.Directions() {
		for _, head := range []bool{false, true} {
			for _, occupied := range []bool{false, true} {
				beds = append(beds, StrawBed{Facing: d, Head: head, Occupied: occupied})
			}
		}
	}
	return
}
