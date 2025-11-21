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

// tryAdvanceDay attempts to advance the day of the world, by first ensuring that all sleepers are sleeping, and then
// updating the time of day.
func (ticker) tryAdvanceDay(tx *Tx, timeCycle bool) {
	sleepers := tx.Sleepers()

	var thunderAnywhere bool
	for s := range sleepers {
		if !thunderAnywhere {
			thunderAnywhere = tx.ThunderingAt(cube.PosFromVec3(s.Position()))
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
	time := totalTime % TimeFull
	if (time < TimeNight || time >= TimeSunrise) && !thunderAnywhere {
		// The conditions for sleeping aren't being met.
		return
	}

	if timeCycle {
		tx.w.SetTime(totalTime + TimeFull - time)
	}
	tx.w.StopRaining()
}
