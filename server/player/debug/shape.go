package debug

import (
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
	"sync/atomic"
)

var nextShapeID atomic.Int32

// Shape represents a shape that can be drawn to a player from any point in the world.
type Shape interface {
	// ShapeID returns the unique identifier of the shape. This is used to either update or remove the shape
	// after it has been sent to the player.
	ShapeID() int
}

// shape is a base type for all shapes that implements the Shape interface. It contains a unique identifier
// that is lazily initialized when the ShapeID method is called for the first time.
type shape struct {
	id *int
}

// ShapeID ...
func (s *shape) ShapeID() int {
	if s.id == nil {
		id := int(nextShapeID.Add(1))
		s.id = &id
	}
	return *s.id
}

// Arrow represents an arrow shape that can be drawn at any point in the world. It has a head which can also
// be positioned anywhere, and the length, radius and number of segments can be changed.
type Arrow struct {
	shape

	// Colour is the colour that will be used for the line and head. If empty, it will default to white.
	Colour color.RGBA
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// EndPosition is the end position of the arrow in the world. The arrow will be drawn from Position to
	// EndPosition, with the head being drawn at EndPosition.
	EndPosition mgl64.Vec3
	// HeadLength is the length of the head to be drawn at the end of the arrow. If zero, it will default
	// to 1.0.
	HeadLength float64
	// HeadRadius is the radius of the head to be drawn at the end of the arrow. If zero, it will default
	// to 0.5.
	HeadRadius float64
	// HeadSegments is the number of segments that the head of the arrow will be drawn with. The more
	// segments, the smoother the head will look. If zero, it will default to 4.
	HeadSegments int
}

// Box represents a hollow box that can be drawn at any point in the world, with a bounds that can be set.
type Box struct {
	shape

	// Colour is the colour that will be used for the outline. If empty, it will default to white.
	Colour color.RGBA
	// Bounds is the size of the box in the world, acting as an offset from the Position. If empty,
	// it will default to a 1x1x1 box.
	Bounds mgl64.Vec3
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// Scale is the rate to scale the shape from its origin point. If zero, it will default to 1.0.
	Scale float64
}

// Circle represents a hollow circle that can be drawn at any point in the world, with the scale being used
// to control the radius.
type Circle struct {
	shape

	// Colour is the colour that will be used for the outline. If empty, it will default to white.
	Colour color.RGBA
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// Scale is the radius of the circle. If zero, it will default to 1.0.
	Scale float64
	// Segments is the number of segments that the circle will be drawn with. The more segments, the smoother
	// the circle will look. If empty, it will default to 20.
	Segments int
}

// Line represents a line that can be drawn at any point in the world, with a start and end position.
type Line struct {
	shape

	// Colour is the colour that will be used for the line. If empty, it will default to white.
	Colour color.RGBA
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// EndPosition is the end position of the line in the world. The line will be drawn from Position to
	// EndPosition.
	EndPosition mgl64.Vec3
}

// Sphere represents a hollow sphere that can be drawn at any point in the world, with one line in each axis.
// The scale is used to control the radius of the sphere.
type Sphere struct {
	shape

	// Colour is the colour that will be used for the outline. If empty, it will default to white.
	Colour color.RGBA
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// Scale is the radius of the sphere. If zero, it will default to 1.0.
	Scale float64
	// Segments is the number of segments that the circle will be drawn with. The more segments, the smoother
	// the circle will look. If empty, it will default to 20.
	Segments int
}

// Text represents text that can be drawn at any point in the world, looking like a normal entity nametag
// without actually being attached to an entity.
type Text struct {
	shape

	// Colour is the colour that will be used for the actual text, not affecting the always-black background.
	// If empty, the text will default to white.
	Colour color.RGBA
	// Position is the origin position of the shape in the world.
	Position mgl64.Vec3
	// Text is the text to be displayed on the shape. The background automatically scales to fit the text.
	Text string
}
