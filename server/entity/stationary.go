package entity

import (
	"math"
	"time"
)

// StationaryBehaviourConfig holds settings that influence the way
// StationaryBehaviour operates. StationaryBehaviourConfig.New() may be called
// to create a new behaviour with this config.
type StationaryBehaviourConfig struct {
	// ExistenceDuration is the duration that an entity with this behaviour
	// should last. Once this time expires, the entity is closed.
	ExistenceDuration time.Duration
}

// New creates a StationaryBehaviour using the settings provided in conf.
func (conf StationaryBehaviourConfig) New() *StationaryBehaviour {
	if conf.ExistenceDuration == 0 {
		conf.ExistenceDuration = math.MaxInt64
	}
	return &StationaryBehaviour{conf: conf}
}

// StationaryBehaviour implements the behaviour of an entity that is unable to
// move, such as a text entity or an area effect cloud. Applying velocity to
// such entities will not move them.
type StationaryBehaviour struct {
	conf  StationaryBehaviourConfig
	age   time.Duration
	close bool
}

// Tick checks if the entity should be closed and runs whatever additional
// behaviour the entity might require.
func (s *StationaryBehaviour) Tick(e *Ent) *Movement {
	if s.close {
		_ = e.Close()
		return nil
	}

	s.age += time.Second / 20
	if s.age > s.conf.ExistenceDuration {
		s.close = true
	}
	// Stationary entities never mode. Always return nil here.
	return nil
}

// Immobile always returns true.
func (s *StationaryBehaviour) Immobile() bool {
	return true
}
