package recipe

import "fmt"

// Dimensions make up the size of a shaped recipe.
type Dimensions [2]int

// DimensionsFrom returns Dimensions from a shape.
func DimensionsFrom(shape []string) (Dimensions, error) {
	height := len(shape)
	if height > 3 || height <= 0 {
		return Dimensions{}, fmt.Errorf("shaped recipes may only have 1, 2 or 3 rows, not height")
	}

	width := len(shape[0])
	if width > 3 || width <= 0 {
		return Dimensions{}, fmt.Errorf("shaped recipes may only have 1, 2 or 3 columns, not width")
	}
	return Dimensions{width, height}, nil
}
