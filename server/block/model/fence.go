package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Fence is a model used by fences of any type. It can attach to blocks with solid faces and other fences of the same
// type and has a model height just slightly over 1.
type Fence struct {
	// Wood specifies if the Fence is made from wood. This field is used to check if two fences are able to attach to
	// each other.
	Wood bool
}

const (
	fenceHeight = 1.5
	fenceInset  = 0.375
)

// BBox returns multiple physics.BBox depending on how many connections it has with the surrounding blocks.
func (f Fence) BBox(pos cube.Pos, s world.BlockSource) []cube.BBox {
	boxes := make([]cube.BBox, 0, 2)

	connectWest, connectEast := f.checkConnection(pos, cube.FaceWest, s), f.checkConnection(pos, cube.FaceEast, s)
	connectNorth, connectSouth := f.checkConnection(pos, cube.FaceNorth, s), f.checkConnection(pos, cube.FaceSouth, s)

	// Check if we have any connections on the Z axis
	if connectWest || connectEast {
		sideBox := cube.Box(0, 0, 0, 1, fenceHeight, 1).Stretch(cube.Z, -fenceInset)
		if connectWest {
			boxes = append(boxes, sideBox.ExtendTowards(cube.FaceEast, -fenceInset))
		}
		if connectEast {
			boxes = append(boxes, sideBox.ExtendTowards(cube.FaceWest, -fenceInset))
		}
	}

	// Check if we have any connections on the X axis
	if connectNorth || connectSouth {
		sideBox := cube.Box(0, 0, 0, 1, fenceHeight, 1).Stretch(cube.X, -fenceInset)
		if connectNorth {
			boxes = append(boxes, sideBox.ExtendTowards(cube.FaceSouth, -fenceInset))
		}
		if connectSouth {
			boxes = append(boxes, sideBox.ExtendTowards(cube.FaceNorth, -fenceInset))
		}
	}

	// If no connections, create a center post box
	if len(boxes) == 0 {
		boxes = append(boxes, cube.Box(fenceInset, 0, fenceInset, 1-fenceInset, fenceHeight, 1-fenceInset))
	}

	return boxes
}

// FaceSolid returns true if the face is cube.FaceDown or cube.FaceUp.
func (f Fence) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceDown || face == cube.FaceUp
}

// checkConnection checks if the block at the given position and face has a connection to the current fence block.
func (f Fence) checkConnection(pos cube.Pos, face cube.Face, src world.BlockSource) bool {
	sidePos := pos.Side(face)
	sideBlock := src.Block(sidePos)
	if fence, ok := sideBlock.Model().(Fence); ok && fence.Wood == f.Wood {
		return true
	}
	if sideBlock.Model().FaceSolid(sidePos, face.Opposite(), src) {
		return true
	}
	_, ok := sideBlock.Model().(FenceGate)
	return ok
}
