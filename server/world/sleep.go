package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/google/uuid"
)

// Sleeper represents an entity that can sleep.
type Sleeper interface {
	Entity

	UUID() uuid.UUID

	Message(a ...any)
	Messaget(key string, a ...string)
	SendSleepingIndicator(sleeping, max int)

	Sleep(pos cube.Pos)
	Sleeping() (cube.Pos, bool)
	Wake()
}

// tryAdvanceDay attempts to advance the day of the world, by first ensuring that all sleepers are sleeping, and then
// updating the time of day.
func (t ticker) tryAdvanceDay() {
	sleepers := t.w.Sleepers()
	if len(sleepers) == 0 {
		// No sleepers in the world.
		return
	}

	thunderAnywhere, advanceTime := false, true
	for _, s := range sleepers {
		if !thunderAnywhere {
			thunderAnywhere = t.w.ThunderingAt(cube.PosFromVec3(s.Position()))
		}
		if _, ok := s.Sleeping(); !ok {
			advanceTime = false
			break
		}
	}
	if !advanceTime {
		// We can't advance the time - not everyone is sleeping.
		return
	}

	for _, s := range sleepers {
		s.Wake()
	}

	totalTime := t.w.Time()
	time := totalTime % TimeFull
	if (time < TimeNight || time >= TimeSunrise) && !thunderAnywhere {
		// The conditions for sleeping aren't being met.
		return
	}

	t.w.SetTime(totalTime + TimeFull - time)
	t.w.StopRaining()
}
