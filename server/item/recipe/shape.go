package recipe

import "fmt"

// Shape make up the shape of a shaped recipe.
type Shape [2]int

// Width returns the width of the shape.
func (s Shape) Width() int {
	return s[0]
}

// Height returns the height of the shape.
func (s Shape) Height() int {
	return s[1]
}

// NewShape returns Shape from a shape.
func NewShape(shape []string) (Shape, error) {
	height := len(shape)
	if height > 3 || height <= 0 {
		return Shape{}, fmt.Errorf("shaped recipes may only have 1, 2 or 3 rows, got %v", height)
	}
	width := len(shape[0])
	if width > 3 || width <= 0 {
		return Shape{}, fmt.Errorf("shaped recipes may only have 1, 2 or 3 columns, got %v", width)
	}
	for _, row := range shape {
		if len(row) != width {
			return Shape{}, fmt.Errorf("shaped recipe rows must all have the same width (expected width, got %v)", len(row))
		}
	}
	return Shape{width, height}, nil
}
