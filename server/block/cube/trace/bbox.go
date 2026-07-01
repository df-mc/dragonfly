package trace

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
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

	var (
		faces = [6]cube.Face{
			cube.FaceWest,
			cube.FaceEast,
			cube.FaceDown,
			cube.FaceUp,
			cube.FaceNorth,
			cube.FaceSouth,
		}
		vecs    [6]mgl64.Vec3
		results [6]bool
	)

	vecs[0], results[0] = vec3OnLineWithX(start, end, min[0])
	vecs[1], results[1] = vec3OnLineWithX(start, end, max[0])
	vecs[2], results[2] = vec3OnLineWithY(start, end, min[1])
	vecs[3], results[3] = vec3OnLineWithY(start, end, max[1])
	vecs[4], results[4] = vec3OnLineWithZ(start, end, min[2])
	vecs[5], results[5] = vec3OnLineWithZ(start, end, max[2])

	results[0] = results[0] && bb.Vec3WithinYZ(vecs[0])
	results[1] = results[1] && bb.Vec3WithinYZ(vecs[1])
	results[2] = results[2] && bb.Vec3WithinXZ(vecs[2])
	results[3] = results[3] && bb.Vec3WithinXZ(vecs[3])
	results[4] = results[4] && bb.Vec3WithinXY(vecs[4])
	results[5] = results[5] && bb.Vec3WithinXY(vecs[5])

	var (
		vec  mgl64.Vec3
		dist = math.MaxFloat64
		face cube.Face
		has  bool
	)

	for i := range 6 {
		if !results[i] {
			continue
		}
		v := vecs[i]
		if d := start.Sub(v).LenSqr(); d < dist {
			vec = v
			dist = d
			has = true
			face = faces[i]
		}
	}

	return BBoxResult{bb: bb, pos: vec, face: face}, has
}

// BBoxIntersects checks if the line segment from start to end intersects the BBox.
// Unlike BBoxIntercept, it only reports whether an intersection exists and does not
// calculate the closest hit position or face.
func BBoxIntersects(bb cube.BBox, start, end mgl64.Vec3) bool {
	min, max := bb.Min(), bb.Max()
	dir := end.Sub(start)
	tMin, tMax := 0.0, 1.0

	for axis := range 3 {
		if mgl64.FloatEqual(dir[axis], 0) {
			if start[axis] < min[axis] || start[axis] > max[axis] {
				return false
			}
			continue
		}

		inv := 1 / dir[axis]
		t1 := (min[axis] - start[axis]) * inv
		t2 := (max[axis] - start[axis]) * inv
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t1 > tMin {
			tMin = t1
		}
		if t2 < tMax {
			tMax = t2
		}
		if tMin > tMax {
			return false
		}
	}
	return true
}

// vec3OnLineWithX returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with an X value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithX(a, b mgl64.Vec3, x float64) (mgl64.Vec3, bool) {
	if mgl64.FloatEqual(b[0], a[0]) {
		return mgl64.Vec3{}, false
	}

	f := (x - a[0]) / (b[0] - a[0])
	if f < 0 || f > 1 {
		return mgl64.Vec3{}, false
	}

	return mgl64.Vec3{x, a[1] + (b[1]-a[1])*f, a[2] + (b[2]-a[2])*f}, true
}

// vec3OnLineWithY returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with a Y value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithY(a, b mgl64.Vec3, y float64) (mgl64.Vec3, bool) {
	if mgl64.FloatEqual(a[1], b[1]) {
		return mgl64.Vec3{}, false
	}

	f := (y - a[1]) / (b[1] - a[1])
	if f < 0 || f > 1 {
		return mgl64.Vec3{}, false
	}

	return mgl64.Vec3{a[0] + (b[0]-a[0])*f, y, a[2] + (b[2]-a[2])*f}, true
}

// vec3OnLineWithZ returns an mgl64.Vec3 on the line between mgl64.Vec3 a and b with a Z value passed. If no such vec3
// could be found, the bool returned is false.
func vec3OnLineWithZ(a, b mgl64.Vec3, z float64) (mgl64.Vec3, bool) {
	if mgl64.FloatEqual(a[2], b[2]) {
		return mgl64.Vec3{}, false
	}

	f := (z - a[2]) / (b[2] - a[2])
	if f < 0 || f > 1 {
		return mgl64.Vec3{}, false
	}

	return mgl64.Vec3{a[0] + (b[0]-a[0])*f, a[1] + (b[1]-a[1])*f, z}, true
}
