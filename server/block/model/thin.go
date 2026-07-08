package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Thin is a model for thin, partial blocks such as a glass pane or an iron bar. It changes its bounding box depending
// on solid faces next to it.
type Thin struct{}

const (
	thinHeight = 1
	thinInset  = 7.0 / 16.0
)

// BBox returns a slice of physics.BBox that depends on the blocks surrounding the Thin block. Thin blocks can connect
// to any other Thin block, wall or solid faces of other blocks.
func (t Thin) BBox(pos cube.Pos, s world.BlockSource) []cube.BBox {
	boxes := make([]cube.BBox, 0, 2)

	// Check if we have any connections on the Z axis
	connectWest, connectEast := t.checkConnection(pos, cube.FaceWest, s), t.checkConnection(pos, cube.FaceEast, s)
	if connectWest || connectEast {
		box := cube.Box(0, 0, 0, 1, thinHeight, 1).Stretch(cube.Z, -thinInset)
		if !connectWest {
			box = box.ExtendTowards(cube.FaceWest, -thinInset)
		} else if !connectEast {
			box = box.ExtendTowards(cube.FaceEast, -thinInset)
		}
		boxes = append(boxes, box)
	}

	// Check if we have any connections on the X axis
	connectNorth, connectSouth := t.checkConnection(pos, cube.FaceNorth, s), t.checkConnection(pos, cube.FaceSouth, s)
	if connectNorth || connectSouth {
		box := cube.Box(0, 0, 0, 1, thinHeight, 1).Stretch(cube.X, -thinInset)
		if !connectNorth {
			box = box.ExtendTowards(cube.FaceNorth, -thinInset)
		} else if !connectSouth {
			box = box.ExtendTowards(cube.FaceSouth, -thinInset)
		}
		boxes = append(boxes, box)
	}

	// If no connections, create a center post box
	if len(boxes) == 0 {
		boxes = append(boxes, cube.Box(0, 0, 0, 1, thinHeight, 1).Stretch(cube.X, -thinInset).Stretch(cube.Z, -thinInset))
	}

	return boxes
}

// FaceSolid returns true if the face passed is cube.FaceDown.
func (t Thin) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceDown
}

// checkConnection checks if the block at the given position and face has a connection to the current thin block.
func (t Thin) checkConnection(pos cube.Pos, face cube.Face, s world.BlockSource) bool {
	sidePos := pos.Side(face)
	sideBlock := s.Block(sidePos)
	_, isThin := sideBlock.Model().(Thin)
	_, isWall := sideBlock.Model().(Wall)
	return isThin || isWall || sideBlock.Model().FaceSolid(sidePos, face.Opposite(), s)
}
