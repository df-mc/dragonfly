package recipe

// Shape make up the shape of a shaped recipe. It consists of a width and a height.
type Shape [2]int

// Width returns the width of the shape.
func (s Shape) Width() int {
	return s[0]
}

// Height returns the height of the shape.
func (s Shape) Height() int {
	return s[1]
}

// NewShape creates a new shape using the provided width and height.
func NewShape(width, height int) Shape {
	return Shape{width, height}
}
