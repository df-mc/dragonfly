package effect

import (
	"reflect"
)

// Register registers an Effect with a specific ID to translate from and to on disk and network. An Effect
// instance may be created by creating a struct instance in this package like
// effect.Regeneration{}.
func Register(id int, e Type) {
	effects[id] = e
	effectIds[reflect.TypeOf(e)] = id
}

// init registers all implemented effects.
func init() {
	Register(1, Speed{})
	Register(2, Slowness{})
	Register(3, Haste{})
	Register(4, MiningFatigue{})
	Register(5, Strength{})
	Register(6, InstantHealth{})
	Register(7, InstantDamage{})
	Register(8, JumpBoost{})
	Register(9, Nausea{})
	Register(10, Regeneration{})
	Register(11, Resistance{})
	Register(12, FireResistance{})
	Register(13, WaterBreathing{})
	Register(14, Invisibility{})
	Register(15, Blindness{})
	Register(16, NightVision{})
	Register(17, Hunger{})
	Register(18, Weakness{})
	Register(19, Poison{})
	Register(20, Wither{})
	Register(21, HealthBoost{})
	Register(22, Absorption{})
	Register(23, Saturation{})
	Register(24, Levitation{})
	Register(25, FatalPoison{})
	Register(26, ConduitPower{})
	Register(27, SlowFalling{})
	// TODO: (28) Bad omen. (Requires villages ...)
	// TODO: (29) Hero of the village. (Requires villages ...)
	Register(30, Darkness{})
}

var (
	effects   = map[int]Type{}
	effectIds = map[reflect.Type]int{}
)

// ByID attempts to return an effect by the ID it was registered with. If found, the effect found
// is returned and the bool true.
func ByID(id int) (Type, bool) {
	effect, ok := effects[id]
	return effect, ok
}

// ID attempts to return the ID an effect was registered with. If found, the id is returned and
// the bool true.
func ID(e Type) (int, bool) {
	id, ok := effectIds[reflect.TypeOf(e)]
	return id, ok
}
