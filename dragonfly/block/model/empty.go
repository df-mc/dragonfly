package model

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Empty is a model that is completely empty. It has no collision boxes or solid faces.
type Empty struct{}

// AABB ...
func (Empty) AABB(cube.Pos, *world.World) []physics.AABB {
	return nil
}

// FaceSolid ...
func (Empty) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
