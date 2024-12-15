package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"time"
)

// StationaryBehaviourConfig holds settings that influence the way
// StationaryBehaviour operates. StationaryBehaviourConfig.New() may be called
// to create a new behaviour with this config.
type StationaryBehaviourConfig struct {
	// ExistenceDuration is the duration that an entity with this behaviour
	// should last. Once this time expires, the entity is closed. If
	// ExistenceDuration is 0, the entity will never expire automatically.
	ExistenceDuration time.Duration
	// SpawnSounds is a slice of sounds to be played upon the spawning of the
	// entity.
	SpawnSounds []world.Sound
	// Tick is a function called every world tick. It may be used to implement
	// additional behaviour for stationary entities.
	Tick func(e *Ent, tx *world.Tx)
}

func (conf StationaryBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
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
	close bool
}

// Tick checks if the entity should be closed and runs whatever additional
// behaviour the entity might require.
func (s *StationaryBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if s.close {
		_ = e.Close()
		return nil
	}

	if e.Age() == 0 {
		for _, ss := range s.conf.SpawnSounds {
			tx.PlaySound(e.Position(), ss)
		}
	}
	if s.conf.Tick != nil {
		s.conf.Tick(e, tx)
	}

	if e.Age() > s.conf.ExistenceDuration {
		s.close = true
	}
	// Stationary entities never move. Always return nil here.
	return nil
}

// Immobile always returns true.
func (s *StationaryBehaviour) Immobile() bool {
	return true
}
