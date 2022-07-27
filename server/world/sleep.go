package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// sleeper represents an entity that can sleep.
type sleeper interface {
	Sleeping() (cube.Pos, bool)
	Wake()
}

// tryAdvanceDay attempts to advance the day of the world, by first ensuring that all sleepers are sleeping, and then
// updating the time of day.
func (t ticker) tryAdvanceDay() {
	ent := t.w.Entities()
	sleepers := make([]sleeper, 0, len(ent))
	for _, e := range ent {
		if s, ok := e.(sleeper); ok {
			sleepers = append(sleepers, s)
		}
	}
	if len(sleepers) == 0 {
		// No sleepers in the world.
		return
	}

	advanceTime := true
	for _, s := range sleepers {
		if _, ok := s.Sleeping(); !ok {
			advanceTime = false
			break
		}
	}
	if !advanceTime {
		// We can't advance the time - not everyone is sleeping.
		return
	}

	totalTime := t.w.Time()
	timeOfDay := totalTime % TimeFull
	if timeOfDay >= TimeNight && timeOfDay < TimeSunrise {
		t.w.SetTime(totalTime + TimeFull - timeOfDay)
		t.w.StopRaining()
		for _, s := range sleepers {
			s.Wake()
		}
	}
}
