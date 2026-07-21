package cube

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

type number interface {
	~float32 | ~float64
}

type vec3[T number] interface {
	~[3]T
}

type boundingBox[T number, V vec3[T]] struct {
	min, max V
}

// BBox represents an Axis Aligned Bounding Box in a 3D space. It is defined as
// two Vec3s, of which one is the minimum and one is the maximum.
type BBox = boundingBox[float64, mgl64.Vec3]

// BBox32 represents a float32 Axis Aligned Bounding Box in a 3D space. It is
// defined as two Vec3s, of which one is the minimum and one is the maximum.
type BBox32 = boundingBox[float32, mgl32.Vec3]

// Box creates a new axis aligned bounding box with the minimum and maximum
// coordinates provided. The returned box has minimum and maximum coordinates
// swapped if necessary so that it is well-formed.
func Box(x0, y0, z0, x1, y1, z1 float64) BBox {
	return newBox[float64, mgl64.Vec3](x0, y0, z0, x1, y1, z1)
}

// Box32 creates a new float32 axis aligned bounding box with the minimum and
// maximum coordinates provided. The returned box has minimum and maximum
// coordinates swapped if necessary so that it is well-formed.
func Box32(x0, y0, z0, x1, y1, z1 float32) BBox32 {
	return newBox[float32, mgl32.Vec3](x0, y0, z0, x1, y1, z1)
}

func newBox[T number, V vec3[T]](x0, y0, z0, x1, y1, z1 T) boundingBox[T, V] {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if z0 > z1 {
		z0, z1 = z1, z0
	}
	return boundingBox[T, V]{min: V{x0, y0, z0}, max: V{x1, y1, z1}}
}

// Grow grows the bounding box in all directions by x and returns the new
// bounding box.
func (box boundingBox[T, V]) Grow(x T) boundingBox[T, V] {
	return box.GrowVec3(V{x, x, x})
}

// GrowVec3 grows the BBox on all axes as represented by the Vec3 passed. The
// vec values are subtracted from the minimum values of the BBox and added to
// the maximum values of the BBox.
func (box boundingBox[T, V]) GrowVec3(vec V) boundingBox[T, V] {
	for i := range 3 {
		box.min[i] -= vec[i]
		box.max[i] += vec[i]
	}
	return box
}

// Min returns the minimum coordinate of the bounding box.
func (box boundingBox[T, V]) Min() V {
	return box.min
}

// Max returns the maximum coordinate of the bounding box.
func (box boundingBox[T, V]) Max() V {
	return box.max
}

// Width returns the width of the BBox.
func (box boundingBox[T, V]) Width() T {
	return box.max[0] - box.min[0]
}

// Length returns the length of the BBox.
func (box boundingBox[T, V]) Length() T {
	return box.max[2] - box.min[2]
}

// Height returns the height of the BBox.
func (box boundingBox[T, V]) Height() T {
	return box.max[1] - box.min[1]
}

// Extend expands the BBox on all axes as represented by the Vec3 passed.
// Negative coordinates result in an expansion towards the negative axis, and
// vice versa for positive coordinates.
func (box boundingBox[T, V]) Extend(vec V) boundingBox[T, V] {
	for i := range 3 {
		if vec[i] < 0 {
			box.min[i] += vec[i]
		} else if vec[i] > 0 {
			box.max[i] += vec[i]
		}
	}
	return box
}

// ExtendTowards extends the bounding box by x in a given direction.
func (box boundingBox[T, V]) ExtendTowards(f Face, x T) boundingBox[T, V] {
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
func (box boundingBox[T, V]) Stretch(a Axis, x T) boundingBox[T, V] {
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
func (box boundingBox[T, V]) Translate(vec V) boundingBox[T, V] {
	for i := range 3 {
		box.min[i] += vec[i]
		box.max[i] += vec[i]
	}
	return box
}

// TranslateTowards moves the entire BBox by x in the direction of a Face f.
func (box boundingBox[T, V]) TranslateTowards(f Face, x T) boundingBox[T, V] {
	var vec V
	switch f {
	case FaceDown:
		vec[1] = -x
	case FaceUp:
		vec[1] = x
	case FaceNorth:
		vec[2] = -x
	case FaceSouth:
		vec[2] = x
	case FaceWest:
		vec[0] = -x
	case FaceEast:
		vec[0] = x
	default:
		return box
	}
	return box.Translate(vec)
}

// IntersectsWith checks if the BBox intersects with another BBox.
func (box boundingBox[T, V]) IntersectsWith(other boundingBox[T, V]) bool {
	return box.intersectsWith(other, 1e-5)
}

// intersectsWith checks if the BBox intersects with another BBox using a
// specific epsilon.
func (box boundingBox[T, V]) intersectsWith(other boundingBox[T, V], epsilon T) bool {
	if other.max[0]-box.min[0] > epsilon && box.max[0]-other.min[0] > epsilon {
		if other.max[1]-box.min[1] > epsilon && box.max[1]-other.min[1] > epsilon {
			return other.max[2]-box.min[2] > epsilon && box.max[2]-other.min[2] > epsilon
		}
	}
	return false
}

// AnyIntersections checks if any of boxes intersect with search.
func AnyIntersections(boxes []BBox, search BBox) bool {
	return anyIntersections(boxes, search)
}

// AnyIntersections32 checks if any of the float32 boxes intersect with search.
func AnyIntersections32(boxes []BBox32, search BBox32) bool {
	return anyIntersections(boxes, search)
}

func anyIntersections[T number, V vec3[T]](boxes []boundingBox[T, V], search boundingBox[T, V]) bool {
	for _, box := range boxes {
		if box.intersectsWith(search, 0) {
			return true
		}
	}
	return false
}

// Vec3Within checks if a BBox has vec within it.
func (box boundingBox[T, V]) Vec3Within(vec V) bool {
	if vec[0] <= box.min[0] || vec[0] >= box.max[0] {
		return false
	}
	if vec[2] <= box.min[2] || vec[2] >= box.max[2] {
		return false
	}
	return vec[1] > box.min[1] && vec[1] < box.max[1]
}

// Vec3WithinYZ checks if a BBox has vec within its Y and Z bounds.
func (box boundingBox[T, V]) Vec3WithinYZ(vec V) bool {
	if vec[2] < box.min[2] || vec[2] > box.max[2] {
		return false
	}
	return vec[1] >= box.min[1] && vec[1] <= box.max[1]
}

// Vec3WithinXZ checks if a BBox has vec within its X and Z bounds.
func (box boundingBox[T, V]) Vec3WithinXZ(vec V) bool {
	if vec[0] < box.min[0] || vec[0] > box.max[0] {
		return false
	}
	return vec[2] >= box.min[2] && vec[2] <= box.max[2]
}

// Vec3WithinXY checks if a BBox has vec within its X and Y bounds.
func (box boundingBox[T, V]) Vec3WithinXY(vec V) bool {
	if vec[0] < box.min[0] || vec[0] > box.max[0] {
		return false
	}
	return vec[1] >= box.min[1] && vec[1] <= box.max[1]
}

// XOffset calculates the offset on the X axis between two bounding boxes,
// returning a delta always smaller than or equal to deltaX if deltaX is bigger
// than 0, or always bigger than or equal to deltaX if it is smaller than 0.
func (box boundingBox[T, V]) XOffset(nearby boundingBox[T, V], deltaX T) T {
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
func (box boundingBox[T, V]) YOffset(nearby boundingBox[T, V], deltaY T) T {
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
func (box boundingBox[T, V]) ZOffset(nearby boundingBox[T, V], deltaZ T) T {
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
func (box boundingBox[T, V]) Corners() []V {
	bbmin, bbmax := box.min, box.max
	return []V{
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
func (box boundingBox[T, V]) Mul(val T) boundingBox[T, V] {
	for i := range 3 {
		box.min[i] *= val
		box.max[i] *= val
	}
	return box
}

// Volume calculates the volume of a BBox.
func (box boundingBox[T, V]) Volume() T {
	return box.Height() * box.Length() * box.Width()
}
