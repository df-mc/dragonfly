package physics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// AABB represents an Axis Aligned Bounding Box in a 3D space. It is defined as two Vec3s, of which one is the
// minimum and one is the maximum.
type AABB struct {
	min, max mgl32.Vec3
}

// NewAABB creates a new axis aligned bounding box with the minimum and maximum coordinates provided.
func NewAABB(min, max mgl32.Vec3) AABB {
	return AABB{min: min, max: max}
}

// Grow grows the bounding box in all directions by x and returns the new bounding box.
func (aabb AABB) Grow(x float32) AABB {
	add := mgl32.Vec3{x, x, x}
	return AABB{min: aabb.min.Sub(add), max: aabb.max.Add(add)}
}

// Min returns the minimum coordinate of the bounding box.
func (aabb AABB) Min() mgl32.Vec3 {
	return aabb.min
}

// Max returns the maximum coordinate of the bounding box.
func (aabb AABB) Max() mgl32.Vec3 {
	return aabb.max
}

// Extend expands the AABB on all axes as represented by the Vec3 passed. Negative coordinates result in an
// expansion towards the negative axis, and vice versa for positive coordinates.
func (aabb AABB) Extend(vec mgl32.Vec3) AABB {
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

// Translate moves the entire AABB with the Vec3 given. The (minimum and maximum) x, y and z coordinates are
// moved by those in the Vec3 passed.
func (aabb AABB) Translate(vec mgl32.Vec3) AABB {
	return NewAABB(aabb.min.Add(vec), aabb.max.Add(vec))
}

// IntersectsWith checks if the AABB intersects with another AABB, returning true if this is the case.
func (aabb AABB) IntersectsWith(other AABB) bool {
	if other.max[0]-aabb.min[0] > 0 && aabb.max[0]-other.min[0] > 0 {
		if other.max[1]-aabb.min[1] > 0 && aabb.max[1]-other.min[1] > 0 {
			return other.max[2]-aabb.min[2] > 0 && aabb.max[2]-other.min[2] > 0
		}
	}
	return false
}

// Vec3Within checks if the AABB has a Vec3 within it, returning true if it does.
func (aabb AABB) Vec3Within(vec mgl32.Vec3) bool {
	if vec[0] <= aabb.min[0] || vec[0] >= aabb.max[0] {
		return false
	}
	if vec[2] <= aabb.min[2] || vec[2] >= aabb.max[2] {
		return false
	}
	return vec[1] > aabb.min[1] && vec[1] < aabb.max[1]
}

// CalculateXOffset calculates the offset on the X axis between two bounding boxes, returning a delta always
// smaller than or equal to deltaX if deltaX is bigger than 0, or always bigger than or equal to deltaX if it
// is smaller than 0.
func (aabb AABB) CalculateXOffset(nearby AABB, deltaX float32) float32 {
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
func (aabb AABB) CalculateYOffset(nearby AABB, deltaY float32) float32 {
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
func (aabb AABB) CalculateZOffset(nearby AABB, deltaZ float32) float32 {
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

// AABBer represents an entity or a block that has one or multiple specific Axis Aligned Bounding Boxes. These
// boxes are used to calculate collision.
type AABBer interface {
	// AABB returns all the axis aligned bounding boxes of the block, or a single box if the AABBer is an
	// entity.
	AABB() []AABB
}
