package world

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"sync"
)

// Settings holds the settings of a World. These are typically saved to a level.dat file. It is safe to pass the same
// Settings to multiple worlds created using New, in which case the Settings are synchronised between the worlds.
type Settings struct {
	sync.Mutex
	ref atomic.Int32

	// Name is the display name of the World.
	Name string
	// Spawn is the spawn position of the World. New players that join the world will be spawned here.
	Spawn cube.Pos
	// Time is the current time of the World. It advances every tick if TimeCycle is set to true.
	Time int64
	// TimeCycle specifies if the time should advance every tick. If set to false, time won't change.
	TimeCycle bool
	// RainTime is the current rain time of the World. It advances every tick if WeatherCycle is set to true.
	RainTime int64
	// Raining is the current rain level of the World.
	Raining bool
	// ThunderTime is the current thunder time of the World. It advances every tick if WeatherCycle is set to true.
	ThunderTime int64
	// Thunder is the current thunder level of the World.
	Thundering bool
	// WeatherCycle specifies if weather should be enabled in this world. If set to false, weather will be disabled.
	WeatherCycle bool
	// CurrentTick is the current tick of the world. This is similar to the Time, except that it has no visible effect
	// to the client. It can also not be changed through commands and will only ever go up.
	CurrentTick int64
	// DefaultGameMode is the GameMode assigned to players that join the World for the first time.
	DefaultGameMode GameMode
	// Difficulty is the difficulty of the World. Behaviour of hunger, regeneration and monsters differs based on the
	// difficulty of the world.
	Difficulty Difficulty
	// TickRange is the radius in chunks around a Viewer that has its blocks and entities ticked when the world is
	// ticked. If set to 0, blocks and entities will never be ticked.
	TickRange int32
}

// defaultSettings returns the default Settings for a new World.
func defaultSettings() *Settings {
	return &Settings{
		Name:            "World",
		DefaultGameMode: GameModeSurvival,
		Difficulty:      DifficultyNormal,
		TimeCycle:       true,
		WeatherCycle:    true,
		TickRange:       6,
	}
}
