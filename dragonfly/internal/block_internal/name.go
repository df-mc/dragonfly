package block_internal

//lint:file-ignore ST1022 Exported variables in this package have compiler directives. These variables are not otherwise exposed to users.
//lint:file-ignore ST1020 Exported functions in this package have compiler directives. These functions are not otherwise exposed to users.

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/particle"
	_ "unsafe" // Imported for compiler directives.
)

//go:linkname world_blocksByName github.com/df-mc/dragonfly/dragonfly/world.blocksByName
//noinspection ALL
var world_blocksByName = map[string]world.Block{}

//go:linkname world_breakParticle github.com/df-mc/dragonfly/dragonfly/world.breakParticle
//noinspection ALL
var world_breakParticle func(b world.Block) world.Particle

//go:linkname World_registeredStates github.com/df-mc/dragonfly/dragonfly/world.registeredStates
//noinspection ALL
var World_registeredStates []world.Block

//go:linkname World_runtimeID github.com/df-mc/dragonfly/dragonfly/world.runtimeID
//noinspection ALL
func World_runtimeID(w *world.World, pos cube.Pos) uint32

func init() {
	world_breakParticle = func(b world.Block) world.Particle {
		return particle.BlockBreak{Block: b}
	}
}

// BlockByTypeName attempts to return a block by its type name.
func BlockByTypeName(name string) (world.Block, bool) {
	v, ok := world_blocksByName[name]
	return v, ok
}

// BlockNames returns a list of all block names.
func BlockNames() []string {
	m := make([]string, 0, len(world_blocksByName))
	for k := range world_blocksByName {
		m = append(m, k)
	}
	return m
}
