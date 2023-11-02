package cmd

import (
	"github.com/go-gl/mathgl/mgl64"
)

// Target represents the target of a command. A []Target may be used as command parameter
// types to allow passing targets to the command.
type Target interface {
	// Position returns the position of the Target as an mgl64.Vec3.
	Position() mgl64.Vec3
}

// NamedTarget is a Target that has a name.
type NamedTarget interface {
	Target
	// Name returns a name of the Target. Note that this name needs not to be and is not unique for a Target.
	Name() string
}

// TargetFunc is a function used to find Targets eligible for a command executed by a given Source. Multiple
// functions may be added by using AddTargetFunc.
type TargetFunc func(src Source) (entities []Target, players []NamedTarget)

// AddTargetFunc adds a TargetFunc to the list of functions used to find targets that may be targeted by a
// Source.
func AddTargetFunc(f TargetFunc) {
	targetFunctions = append(targetFunctions, f)
}

// targetFunctions holds a list of all TargetFunc registered using AddTargetFunc.
var targetFunctions []TargetFunc

// targets returns all Targets selectable by the Source passed.
func targets(src Source) (entities []Target, players []NamedTarget) {
	for _, f := range targetFunctions {
		e, p := f(src)
		entities = append(entities, e...)
		players = append(players, p...)
	}
	return
}
