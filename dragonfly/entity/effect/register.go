package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"reflect"
)

// Register registers an entity.Effect with a specific ID to translate from and to on disk and network. An
// entity.Effect instance may be created by creating a struct instance in this package like
// effect.Regeneration{}.
func Register(id int, e entity.Effect) {
	effects[id] = e
	effectIds[reflect.TypeOf(e)] = id
}

// init registers all implemented effects.
func init() {
	Register(1, Speed{})
	Register(2, Slowness{})
	Register(3, Haste{})
	Register(4, MiningFatigue{})
	Register(6, InstantHealth{})
	Register(10, Regeneration{})
}

var (
	effects   = map[int]entity.Effect{}
	effectIds = map[reflect.Type]int{}
)

// effectByID attempts to return an effect by the ID it was registered with. If found, the effect found
// is returned and the bool true.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func effectByID(id int) (entity.Effect, bool) {
	effect, ok := effects[id]
	return effect, ok
}

// idByEffect attempts to return the ID an effect was registered with. If found, the id is returned and
// the bool true.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func idByEffect(e entity.Effect) (int, bool) {
	id, ok := effectIds[reflect.TypeOf(e)]
	return id, ok
}
