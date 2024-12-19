package cmd

import (
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"slices"
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

// targets returns all Targets selectable by the Source passed.
func targets(tx *world.Tx) (entities []Target, players []NamedTarget) {
	ent := sliceutil.Convert[Target](slices.Collect(tx.Entities()))
	pl := sliceutil.Convert[NamedTarget](slices.Collect(tx.Players()))
	return ent, pl
}
