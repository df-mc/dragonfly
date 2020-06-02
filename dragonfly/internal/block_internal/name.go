package block_internal

//lint:file-ignore ST1022 Exported variables in this package have compiler directives. These variables are not otherwise exposed to users.
//lint:file-ignore ST1020 Exported functions in this package have compiler directives. These functions are not otherwise exposed to users.

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/particle"
	_ "unsafe" // Imported for compiler directives.
)

//go:linkname world_blocksByName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.blocksByName
//noinspection ALL
var world_blocksByName = map[string]world.Block{}

//go:linkname world_breakParticle git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.breakParticle
//noinspection ALL
var world_breakParticle func(b world.Block) world.Particle

//go:linkname World_registeredStates git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.registeredStates
//noinspection ALL
var World_registeredStates []world.Block

//go:linkname World_runtimeID git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.runtimeID
//noinspection ALL
func World_runtimeID(w *world.World, pos world.BlockPos) uint32

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
