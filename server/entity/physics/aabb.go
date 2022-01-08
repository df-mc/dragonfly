package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// AABB represents an Axis Aligned Bounding Box in a 3D space. It is defined as two Vec3s, of which one is the
// minimum and one is the maximum.
type AABB struct {
	min, max mgl64.Vec3
}

// NewAABB creates a new axis aligned bounding box with the minimum and maximum coordinates provided.
func NewAABB(min, max mgl64.Vec3) AABB {
	return AABB{min: min, max: max}
}

// Grow grows the bounding box in all directions by x and returns the new bounding box.
func (aabb AABB) Grow(x float64) AABB {
	add := mgl64.Vec3{x, x, x}
	return AABB{min: aabb.min.Sub(add), max: aabb.max.Add(add)}
}

// GrowVec3 grows the AABB on all axes as represented by the Vec3 passed. The vec values are subtracted from
// the minimum values of the AABB and added to the maximum values of the AABB.
func (aabb AABB) GrowVec3(vec mgl64.Vec3) AABB {
	return AABB{min: aabb.min.Sub(vec), max: aabb.max.Add(vec)}
}

// Min returns the minimum coordinate of the bounding box.
func (aabb AABB) Min() mgl64.Vec3 {
	return aabb.min
}

// Max returns the maximum coordinate of the bounding box.
func (aabb AABB) Max() mgl64.Vec3 {
	return aabb.max
}

// Width returns the width of the AABB.
func (aabb AABB) Width() float64 {
	return aabb.max[0] - aabb.min[0]
}

// Length returns the length of the AABB.
func (aabb AABB) Length() float64 {
	return aabb.max[2] - aabb.min[2]
}

// Height returns the height of the AABB.
func (aabb AABB) Height() float64 {
	return aabb.max[1] - aabb.min[1]
}

// Extend expands the AABB on all axes as represented by the Vec3 passed. Negative coordinates result in an
// expansion towards the negative axis, and vice versa for positive coordinates.
func (aabb AABB) Extend(vec mgl64.Vec3) AABB {
	if vec[0] < 0 {
		aabb.min[0] += vec[0]
	} else if vec[0] > 0 {
		aabb.max[0] += vec[0]
	}
	if vec[1] < 0 {
		aabb.min[1] += vec[1]
	} else if vec[1] > 0 {
		aabb.max[1] += vec[1]
	}
	if vec[2] < 0 {
		aabb.min[2] += vec[2]
	} else if vec[2] > 0 {
		aabb.max[2] += vec[2]
	}
	return aabb
}

// ExtendTowards extends the bounding box by x in a given direction.
func (aabb AABB) ExtendTowards(f cube.Face, x float64) AABB {
	switch f {
	case cube.FaceDown:
		aabb.max[1] -= x
	case cube.FaceUp:
		aabb.min[1] += x
	case cube.FaceNorth:
		aabb.min[2] -= x
	case cube.FaceSouth:
		aabb.max[2] += x
	case cube.FaceWest:
		aabb.min[0] -= x
	case cube.FaceEast:
		aabb.max[0] += x
	}
	return aabb
}

// Stretch stretches the bounding box by x in a given axis.
func (aabb AABB) Stretch(a cube.Axis, x float64) AABB {
	switch a {
	case cube.Y:
		aabb.min[1] -= x
		aabb.max[1] += x
	case cube.Z:
		aabb.min[2] -= x
		aabb.max[2] += x
	case cube.X:
		aabb.min[0] -= x
		aabb.max[0] += x
	}
	return aabb
}

// Translate moves the entire AABB with the Vec3 given. The (minimum and maximum) x, y and z coordinates are
// moved by those in the Vec3 passed.
func (aabb AABB) Translate(vec mgl64.Vec3) AABB {
	return NewAABB(aabb.min.Add(vec), aabb.max.Add(vec))
}

// IntersectsWith checks if the AABB intersects with another AABB, returning true if this is the case.
func (aabb AABB) IntersectsWith(other AABB) bool {
	if other.max[0]-aabb.min[0] > 1e-5 && aabb.max[0]-other.min[0] > 1e-5 {
		if other.max[1]-aabb.min[1] > 1e-5 && aabb.max[1]-other.min[1] > 1e-5 {
			return other.max[2]-aabb.min[2] > 1e-5 && aabb.max[2]-other.min[2] > 1e-5
		}
	}
	return false
}

// AnyIntersections checks if any of boxes1 have intersections with any of boxes2 and returns true if this
// happens to be the case.
func AnyIntersections(boxes []AABB, search AABB) bool {
	for _, box := range boxes {
		if box.IntersectsWith(search) {
			return true
		}
	}
	return false
}

// Vec3Within checks if the AABB has a Vec3 within it, returning true if it does.
func (aabb AABB) Vec3Within(vec mgl64.Vec3) bool {
	if vec[0] <= aabb.min[0] || vec[0] >= aabb.max[0] {
		return false
	}
	if vec[2] <= aabb.min[2] || vec[2] >= aabb.max[2] {
		return false
	}
	return vec[1] > aabb.min[1] && vec[1] < aabb.max[1]
}

// Vec3WithinYZ checks if the AABB has a Vec3 within its Y and Z bounds, returning true if it does.
func (aabb AABB) Vec3WithinYZ(vec mgl64.Vec3) bool {
	if vec[2] < aabb.min[2] || vec[2] > aabb.max[2] {
		return false
	}
	return vec[1] >= aabb.min[1] && vec[1] <= aabb.max[1]
}

// Vec3WithinXZ checks if the AABB has a Vec3 within its X and Z bounds, returning true if it does.
func (aabb AABB) Vec3WithinXZ(vec mgl64.Vec3) bool {
	if vec[0] < aabb.min[0] || vec[0] > aabb.max[0] {
		return false
	}
	return vec[2] >= aabb.min[2] && vec[2] <= aabb.max[2]
}

// Vec3WithinXY checks if the AABB has a Vec3 within its X and Y bounds, returning true if it does.
func (aabb AABB) Vec3WithinXY(vec mgl64.Vec3) bool {
	if vec[0] < aabb.min[0] || vec[0] > aabb.max[0] {
		return false
	}
	return vec[1] >= aabb.min[1] && vec[1] <= aabb.max[1]
}

// CalculateXOffset calculates the offset on the X axis between two bounding boxes, returning a delta always
// smaller than or equal to deltaX if deltaX is bigger than 0, or always bigger than or equal to deltaX if it
// is smaller than 0.
func (aabb AABB) CalculateXOffset(nearby AABB, deltaX float64) float64 {
	// Bail out if not within the same Y/Z plane.
	if aabb.max[1] <= nearby.min[1] || aabb.min[1] >= nearby.max[1] {
		return deltaX
	} else if aabb.max[2] <= nearby.min[2] || aabb.min[2] >= nearby.max[2] {
		return deltaX
	}
	if deltaX > 0 && aabb.max[0] <= nearby.min[0] {
		difference := nearby.min[0] - aabb.max[0]
		if difference < deltaX {
			deltaX = difference
		}
	}
	if deltaX < 0 && aabb.min[0] >= nearby.max[0] {
		difference := nearby.max[0] - aabb.min[0]

		if difference > deltaX {
			deltaX = difference
		}
	}
	return deltaX
}

// CalculateYOffset calculates the offset on the Y axis between two bounding boxes, returning a delta always
// smaller than or equal to deltaY if deltaY is bigger than 0, or always bigger than or equal to deltaY if it
// is smaller than 0.
func (aabb AABB) CalculateYOffset(nearby AABB, deltaY float64) float64 {
	// Bail out if not within the same X/Z plane.
	if aabb.max[0] <= nearby.min[0] || aabb.min[0] >= nearby.max[0] {
		return deltaY
	} else if aabb.max[2] <= nearby.min[2] || aabb.min[2] >= nearby.max[2] {
		return deltaY
	}
	if deltaY > 0 && aabb.max[1] <= nearby.min[1] {
		difference := nearby.min[1] - aabb.max[1]
		if difference < deltaY {
			deltaY = difference
		}
	}
	if deltaY < 0 && aabb.min[1] >= nearby.max[1] {
		difference := nearby.max[1] - aabb.min[1]

		if difference > deltaY {
			deltaY = difference
		}
	}
	return deltaY
}

// CalculateZOffset calculates the offset on the Z axis between two bounding boxes, returning a delta always
// smaller than or equal to deltaZ if deltaZ is bigger than 0, or always bigger than or equal to deltaZ if it
// is smaller than 0.
func (aabb AABB) CalculateZOffset(nearby AABB, deltaZ float64) float64 {
	// Bail out if not within the same X/Y plane.
	if aabb.max[0] <= nearby.min[0] || aabb.min[0] >= nearby.max[0] {
		return deltaZ
	} else if aabb.max[1] <= nearby.min[1] || aabb.min[1] >= nearby.max[1] {
		return deltaZ
	}
	if deltaZ > 0 && aabb.max[2] <= nearby.min[2] {
		difference := nearby.min[2] - aabb.max[2]
		if difference < deltaZ {
			deltaZ = difference
		}
	}
	if deltaZ < 0 && aabb.min[2] >= nearby.max[2] {
		difference := nearby.max[2] - aabb.min[2]

		if difference > deltaZ {
			deltaZ = difference
		}
	}
	return deltaZ
}
