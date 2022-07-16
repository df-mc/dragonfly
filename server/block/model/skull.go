package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Skull is a model used by skull blocks.
type Skull struct {
	// Direction is the direction the skull is facing.
	Direction cube.Face
	// Hanging specifies if the Skull is hanging on a wall.
	Hanging bool
}

// BBox ...
func (s Skull) BBox(cube.Pos, *world.World) []cube.BBox {
	box := cube.Box(0.25, 0, 0.25, 0.75, 0.5, 0.75)
	if !s.Hanging {
		return []cube.BBox{box}
	}
	return []cube.BBox{box.TranslateTowards(s.Direction.Opposite(), 0.25).TranslateTowards(cube.FaceUp, 0.25)}
}

// FaceSolid ...
func (Skull) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}
