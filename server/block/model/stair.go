package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Stair is a model for stair-like blocks. These have different solid sides depending on the direction the
// stairs are facing, the corner they form with the stairs around them and whether it is upside down or not.
type Stair struct {
	// Facing specifies the direction that the full side of the Stair faces.
	Facing cube.Direction
	// UpsideDown turns the Stair upside-down, meaning the full side of the Stair is turned to the top side of the
	// block.
	UpsideDown bool
	// Corner specifies if the Stair forms a corner with the stairs around it.
	Corner bool
	// Inner specifies if the corner formed turns inwards, filling up the block rather than a quarter of it. Inner is
	// only used if Corner is true.
	Inner bool
	// Left specifies if the corner formed is on the left side of the Stair. Left is only used if Corner is true.
	Left bool
}

// BBox returns a slice of physics.BBox depending on if the Stair is upside down, which direction it is facing and the
// corner it forms with the stairs around it.
func (s Stair) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	b := []cube.BBox{cube.Box(0, 0, 0, 1, 0.5, 1)}
	if s.UpsideDown {
		b[0] = cube.Box(0, 0.5, 0, 1, 1, 1)
	}
	step := cube.Box(0.5, 0.5, 0.5, 0.5, 1, 0.5)
	if !s.Corner || s.Inner {
		b = append(b, step.ExtendTowards(s.Facing.Face(), 0.5).Stretch(s.Facing.RotateRight().Face().Axis(), 0.5))
	}
	if s.Corner {
		// An outer corner only fills up the quarter of the block in front of the stairs, while an inner corner fills up
		// the quarter behind them on top of the step above.
		face := s.Facing.Face()
		if s.Inner {
			face = s.Facing.Opposite().Face()
		}
		b = append(b, step.ExtendTowards(face, 0.5).ExtendTowards(s.cornerFace(), 0.5))
	}
	if s.UpsideDown {
		for i := range b[1:] {
			b[i+1] = b[i+1].Translate(mgl64.Vec3{0, -0.5})
		}
	}
	return b
}

// FaceSolid returns true for all faces of the Stair that are completely filled.
func (s Stair) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	// Stairs always have a closed side at the top or bottom based on their orientation.
	if (face == cube.FaceUp && s.UpsideDown) || (face == cube.FaceDown && !s.UpsideDown) {
		return true
	}
	if !s.Corner {
		// Not a corner, so only the side behind the stairs is closed.
		return face == s.Facing.Face()
	}
	if !s.Inner {
		// Small corner blocks, they do not block water flowing out horizontally.
		return false
	}
	return face == s.cornerFace() || face == s.Facing.Face()
}

// cornerFace returns the face of the side that the corner of the Stair is on.
func (s Stair) cornerFace() cube.Face {
	if s.Left {
		return s.Facing.RotateLeft().Face()
	}
	return s.Facing.RotateRight().Face()
}
