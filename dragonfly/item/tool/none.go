package tool

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"

// None is a tool type typically used in functions for items that do not function as tools.
type None struct{}

// ToolType ...
func (n None) ToolType() Type {
	return TypeNone
}

// HarvestLevel ...
func (n None) HarvestLevel() int {
	return 0
}

// BaseMiningEfficiency ...
func (n None) BaseMiningEfficiency(world.Block) float64 {
	return 1
}
