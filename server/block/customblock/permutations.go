package customblock

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type Properties struct {
	CollisionBox cube.BBox
	Cube         bool
	Geometry     string
	MapColour    string
	Rotation     cube.Pos
	Scale        mgl64.Vec3
	SelectionBox cube.BBox
	Textures     map[string]Material
	Translation  mgl64.Vec3
}

type Permutation struct {
	Properties
	Condition string
}
