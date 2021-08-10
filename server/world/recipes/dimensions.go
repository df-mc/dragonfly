package recipes

// Dimensions make up the size of a shaped recipe.
type Dimensions struct {
	// Width is the width of the recipe's shape.
	Width int32
	// Height is the height of the recipe's shape.
	Height int32
}

// DimensionsFrom returns Dimensions from a shape.
func DimensionsFrom(shape []string) Dimensions {
	return Dimensions{
		Width:  int32(len(shape[0])),
		Height: int32(len(shape)),
	}
}
