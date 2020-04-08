package block_internal

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	_ "unsafe" // Imported for compiler directives.
)

//go:linkname world_blocksByName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.blocksByName
//noinspection ALL
var world_blocksByName = map[string]world.Block{}

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
