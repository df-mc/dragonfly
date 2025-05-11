package cube

import (
	"github.com/go-gl/mathgl/mgl64"
)

// BBox represents an Axis Aligned Bounding Box in a 3D space. It is defined as
// two Vec3s, of which one is the minimum and one is the maximum.
type BBox struct {
	min, max mgl64.Vec3
}

// Box creates a new axis aligned bounding box with the minimum and maximum
// coordinates provided. The returned box has minimum and maximum coordinates
// swapped if necessary so that it is well-formed.
func Box(x0, y0, z0, x1, y1, z1 float64) BBox {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if z0 > z1 {
		z0, z1 = z1, z0
	}
	return BBox{min: mgl64.Vec3{x0, y0, z0}, max: mgl64.Vec3{x1, y1, z1}}
}

// Grow grows the bounding box in all directions by x and returns the new
// bounding box.
func (box BBox) Grow(x float64) BBox {
	add := mgl64.Vec3{x, x, x}
	return BBox{min: box.min.Sub(add), max: box.max.Add(add)}
}

// GrowVec3 grows the BBox on all axes as represented by the Vec3 passed. The
// vec values are subtracted from the minimum values of the BBox and added to
// the maximum values of the BBox.
func (box BBox) GrowVec3(vec mgl64.Vec3) BBox {
	return BBox{min: box.min.Sub(vec), max: box.max.Add(vec)}
}

// Min returns the minimum coordinate of the bounding box.
func (box BBox) Min() mgl64.Vec3 {
	return box.min
}

// Max returns the maximum coordinate of the bounding box.
func (box BBox) Max() mgl64.Vec3 {
	return box.max
}

// Width returns the width of the BBox.
func (box BBox) Width() float64 {
	return box.max[0] - box.min[0]
}

// Length returns the length of the BBox.
func (box BBox) Length() float64 {
	return box.max[2] - box.min[2]
}

// Height returns the height of the BBox.
func (box BBox) Height() float64 {
	return box.max[1] - box.min[1]
}

// Extend expands the BBox on all axes as represented by the Vec3 passed.
// Negative coordinates result in an expansion towards the negative axis, and
// vice versa for positive coordinates.
func (box BBox) Extend(vec mgl64.Vec3) BBox {
	if vec[0] < 0 {
		box.min[0] += vec[0]
	} else if vec[0] > 0 {
		box.max[0] += vec[0]
	}
	if vec[1] < 0 {
		box.min[1] += vec[1]
	} else if vec[1] > 0 {
		box.max[1] += vec[1]
	}
	if vec[2] < 0 {
		box.min[2] += vec[2]
	} else if vec[2] > 0 {
		box.max[2] += vec[2]
	}
	return box
}

// ExtendTowards extends the bounding box by x in a given direction.
func (box BBox) ExtendTowards(f Face, x float64) BBox {
	switch f {
	case FaceDown:
		box.min[1] -= x
	case FaceUp:
		box.max[1] += x
	case FaceNorth:
		box.min[2] -= x
	case FaceSouth:
		box.max[2] += x
	case FaceWest:
		box.min[0] -= x
	case FaceEast:
		box.max[0] += x
	}
	return box
}

// Stretch stretches the bounding box by x in a given axis.
func (box BBox) Stretch(a Axis, x float64) BBox {
	switch a {
	case Y:
		box.min[1] -= x
		box.max[1] += x
	case Z:
		box.min[2] -= x
		box.max[2] += x
	case X:
		box.min[0] -= x
		box.max[0] += x
	}
	return box
}

// Translate moves the entire BBox with the Vec3 given. The (minimum and
// maximum) x, y and z coordinates are moved by those in the Vec3 passed.
func (box BBox) Translate(vec mgl64.Vec3) BBox {
	return BBox{min: box.min.Add(vec), max: box.max.Add(vec)}
}

// TranslateTowards moves the entire BBox by x in the direction of a Face f.
func (box BBox) TranslateTowards(f Face, x float64) BBox {
	switch f {
	case FaceDown:
		return box.Translate(mgl64.Vec3{0, -x, 0})
	case FaceUp:
		return box.Translate(mgl64.Vec3{0, x, 0})
	case FaceNorth:
		return box.Translate(mgl64.Vec3{0, 0, -x})
	case FaceSouth:
		return box.Translate(mgl64.Vec3{0, 0, x})
	case FaceWest:
		return box.Translate(mgl64.Vec3{-x, 0, 0})
	case FaceEast:
		return box.Translate(mgl64.Vec3{x, 0, 0})
	}
	return box
}

// IntersectsWith checks if the BBox intersects with another BBox.
func (box BBox) IntersectsWith(other BBox) bool {
	return box.intersectsWith(other, 1e-5)
}

// intersectsWith checks if the BBox intersects with another BBox using a
// specific epsilon.
func (box BBox) intersectsWith(other BBox, epsilon float64) bool {
	if other.max[0]-box.min[0] > epsilon && box.max[0]-other.min[0] > epsilon {
		if other.max[1]-box.min[1] > epsilon && box.max[1]-other.min[1] > epsilon {
			return other.max[2]-box.min[2] > epsilon && box.max[2]-other.min[2] > epsilon
		}
	}
	return false
}

// AnyIntersections checks if any of boxes intersect with search.
func AnyIntersections(boxes []BBox, search BBox) bool {
	for _, box := range boxes {
		if box.intersectsWith(search, 0) {
			return true
		}
	}
	return false
}

// Vec3Within checks if a BBox has vec within it.
func (box BBox) Vec3Within(vec mgl64.Vec3) bool {
	if vec[0] <= box.min[0] || vec[0] >= box.max[0] {
		return false
	}
	if vec[2] <= box.min[2] || vec[2] >= box.max[2] {
		return false
	}
	return vec[1] > box.min[1] && vec[1] < box.max[1]
}

// Vec3WithinYZ checks if a BBox has vec within its Y and Z bounds.
func (box BBox) Vec3WithinYZ(vec mgl64.Vec3) bool {
	if vec[2] < box.min[2] || vec[2] > box.max[2] {
		return false
	}
	return vec[1] >= box.min[1] && vec[1] <= box.max[1]
}

// Vec3WithinXZ checks if a BBox has vec within its X and Z bounds.
func (box BBox) Vec3WithinXZ(vec mgl64.Vec3) bool {
	if vec[0] < box.min[0] || vec[0] > box.max[0] {
		return false
	}
	return vec[2] >= box.min[2] && vec[2] <= box.max[2]
}

// Vec3WithinXY checks if a BBox has vec within its X and Y bounds.
func (box BBox) Vec3WithinXY(vec mgl64.Vec3) bool {
	if vec[0] < box.min[0] || vec[0] > box.max[0] {
		return false
	}
	return vec[1] >= box.min[1] && vec[1] <= box.max[1]
}

// XOffset calculates the offset on the X axis between two bounding boxes,
// returning a delta always smaller than or equal to deltaX if deltaX is bigger
// than 0, or always bigger than or equal to deltaX if it is smaller than 0.
func (box BBox) XOffset(nearby BBox, deltaX float64) float64 {
	if box.max[1] <= nearby.min[1] || box.min[1] >= nearby.max[1] || box.max[2] <= nearby.min[2] || box.min[2] >= nearby.max[2] {
		// Not in the same Y/Z plane.
		return deltaX
	}
	if deltaX > 0 && box.max[0] <= nearby.min[0] {
		deltaX = min(deltaX, nearby.min[0]-box.max[0])
	} else if deltaX < 0 && box.min[0] >= nearby.max[0] {
		deltaX = max(deltaX, nearby.max[0]-box.min[0])
	}
	return deltaX
}

// YOffset calculates the offset on the Y axis between two bounding boxes,
// returning a delta always smaller than or equal to deltaY if deltaY is bigger
// than 0, or always bigger than or equal to deltaY if it is smaller than 0.
func (box BBox) YOffset(nearby BBox, deltaY float64) float64 {
	if box.max[0] <= nearby.min[0] || box.min[0] >= nearby.max[0] || box.max[2] <= nearby.min[2] || box.min[2] >= nearby.max[2] {
		// Not the same X/Z plane.
		return deltaY
	}
	if deltaY > 0 && box.max[1] <= nearby.min[1] {
		deltaY = min(deltaY, nearby.min[1]-box.max[1])
	}
	if deltaY < 0 && box.min[1] >= nearby.max[1] {
		deltaY = max(deltaY, nearby.max[1]-box.min[1])
	}
	return deltaY
}

// ZOffset calculates the offset on the Z axis between two bounding boxes,
// returning a delta always smaller than or equal to deltaZ if deltaZ is bigger
// than 0, or always bigger than or equal to deltaZ if it is smaller than 0.
func (box BBox) ZOffset(nearby BBox, deltaZ float64) float64 {
	if box.max[0] <= nearby.min[0] || box.min[0] >= nearby.max[0] || box.max[1] <= nearby.min[1] || box.min[1] >= nearby.max[1] {
		// Not the same X/Y plane.
		return deltaZ
	}
	if deltaZ > 0 && box.max[2] <= nearby.min[2] {
		deltaZ = min(deltaZ, nearby.min[2]-box.max[2])
	}
	if deltaZ < 0 && box.min[2] >= nearby.max[2] {
		deltaZ = max(deltaZ, nearby.max[2]-box.min[2])
	}
	return deltaZ
}

// Corners returns the positions of all corners of a BBox.
func (box BBox) Corners() []mgl64.Vec3 {
	bbmin, bbmax := box.min, box.max
	return []mgl64.Vec3{
		box.min,
		box.max,
		{bbmin[0], bbmin[1], bbmax[2]},
		{bbmin[0], bbmax[1], bbmin[2]},
		{bbmin[0], bbmax[1], bbmax[2]},
		{bbmax[0], bbmax[1], bbmin[2]},
		{bbmax[0], bbmin[1], bbmax[2]},
		{bbmax[0], bbmin[1], bbmin[2]},
	}
}

// Mul performs a scalar multiplication of the min and max points of a BBox.
func (box BBox) Mul(val float64) BBox {
	return BBox{min: box.min.Mul(val), max: box.max.Mul(val)}
}

// Volume calculates the volume of a BBox.
func (box BBox) Volume() float64 {
	return box.Height() * box.Length() * box.Width()
}
