package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/google/uuid"
)

// Sleeper represents an entity that can sleep.
type Sleeper interface {
	Entity

	Name() string
	UUID() uuid.UUID

	Messaget(t chat.Translation, a ...any)
	SendSleepingIndicator(sleeping, max int)

	Sleep(pos cube.Pos)
	Sleeping() (cube.Pos, bool)
	Wake()
}

// Time constants for sleep usage.
const (
	TimeSleep         = 12010
	TimeWake          = 23991
	TimeSleepWithRain = 12542
	TimeWakeWithRain  = 23459
	TimeFull          = 24000
)

// tryAdvanceDay attempts to advance the day of the world, by first ensuring that all sleepers are sleeping, and then
// updating the time of day.
func (ticker) tryAdvanceDay(tx *Tx, timeCycle bool) {
	sleepers := tx.Sleepers()
	time := tx.w.Time() % TimeFull

	for s := range sleepers {
		pos := cube.PosFromVec3(s.Position())

		if !tx.ThunderingAt(pos) {
			if time <= TimeSleep || time >= TimeWake {
				return
			}

			if !tx.RainingAt(pos) && (time <= TimeSleepWithRain || time >= TimeWakeWithRain) {
				return
			}
		}

		if _, ok := s.Sleeping(); !ok {
			// We can't advance the time - not everyone is sleeping.
			return
		}
	}

	for s := range sleepers {
		s.Wake()
	}

	totalTime := tx.w.Time()
	if timeCycle {
		tx.w.SetTime(totalTime + TimeFull - time)
	}
	tx.w.StopRaining()
}
