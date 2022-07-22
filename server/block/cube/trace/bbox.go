package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// BBoxResult is the result of a basic ray trace collision with a bounding box.
type BBoxResult struct {
	bb   cube.BBox
	pos  mgl64.Vec3
	face cube.Face
}

// BBox ...
func (r BBoxResult) BBox() cube.BBox {
	return r.bb
}

// Position ...
func (r BBoxResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face ...
func (r BBoxResult) Face() cube.Face {
	return r.face
}

// BBoxIntercept performs a ray trace and calculates the point on the BBox's edge nearest to the start position that the ray trace
// collided with.
// BBoxIntercept returns a BBoxResult with the colliding vector closest to the start position, if no colliding point was found,
// a zero BBoxResult is returned and ok is false.
func BBoxIntercept(bb cube.BBox, start, end mgl64.Vec3) (result BBoxResult, ok bool) {
	min, max := bb.Min(), bb.Max()
	v1 := vec3OnLineWithX(start, end, min[0])
	v2 := vec3OnLineWithX(start, end, max[0])
	v3 := vec3OnLineWithY(start, end, min[1])
	v4 := vec3OnLineWithY(start, end, max[1])
	v5 := vec3OnLineWithZ(start, end, min[2])
	v6 := vec3OnLineWithZ(start, end, max[2])

	if v1 != nil && !bb.Vec3WithinYZ(*v1) {
		v1 = nil
	}
	if v2 != nil && !bb.Vec3WithinYZ(*v2) {
		v2 = nil
	}
	if v3 != nil && !bb.Vec3WithinXZ(*v3) {
		v3 = nil
	}
	if v4 != nil && !bb.Vec3WithinXZ(*v4) {
		v4 = nil
	}
	if v5 != nil && !bb.Vec3WithinXY(*v5) {
		v5 = nil
	}
	if v6 != nil && !bb.Vec3WithinXY(*v6) {
		v6 = nil
	}

	var (
		vec  *mgl64.Vec3
		dist = math.MaxFloat64
	)

	for _, v := range [...]*mgl64.Vec3{v1, v2, v3, v4, v5, v6} {
		if v == nil {
			continue
		}

		if d := start.Sub(*v).LenSqr(); d < dist {
			vec = v
			dist = d
		}
	}

	if vec == nil {
		return
	}

	var f cube.Face
	switch vec {
	case v1:
		f = cube.FaceWest
	case v2:
		f = cube.FaceEast
	case v3:
		f = cube.FaceDown
	case v4:
		f = cube.FaceUp
	case v5:
		f = cube.FaceNorth
	case v6:
		f = cube.FaceSouth
	}

	return BBoxResult{bb: bb, pos: *vec, face: f}, true
}

// vec3OnLineWithX returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with an X value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithX(a, b mgl64.Vec3, x float64) *mgl64.Vec3 {
	if mgl64.FloatEqual(b[0], a[0]) {
		return nil
	}

	f := (x - a[0]) / (b[0] - a[0])
	if f < 0 || f > 1 {
		return nil
	}

	return &mgl64.Vec3{x, a[1] + (b[1]-a[1])*f, a[2] + (b[2]-a[2])*f}
}

// vec3OnLineWithY returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with a Y value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithY(a, b mgl64.Vec3, y float64) *mgl64.Vec3 {
	if mgl64.FloatEqual(a[1], b[1]) {
		return nil
	}

	f := (y - a[1]) / (b[1] - a[1])
	if f < 0 || f > 1 {
		return nil
	}

	return &mgl64.Vec3{a[0] + (b[0]-a[0])*f, y, a[2] + (b[2]-a[2])*f}
}

// vec3OnLineWithZ returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with a Z value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithZ(a, b mgl64.Vec3, z float64) *mgl64.Vec3 {
	if mgl64.FloatEqual(a[2], b[2]) {
		return nil
	}

	f := (z - a[2]) / (b[2] - a[2])
	if f < 0 || f > 1 {
		return nil
	}

	return &mgl64.Vec3{a[0] + (b[0]-a[0])*f, a[1] + (b[1]-a[1])*f, z}
}
