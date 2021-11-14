package recipe

// Dimensions make up the size of a shaped recipe.
type Dimensions [2]int

// DimensionsFrom returns Dimensions from a shape.
func DimensionsFrom(shape []string) Dimensions {
	return Dimensions{len(shape[0]), len(shape)}
}
