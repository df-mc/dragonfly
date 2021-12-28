package biome

// WarpedForest ...
type WarpedForest struct{}

// Temperature ...
func (WarpedForest) Temperature() float64 {
	return 2
}

// Rainfall ...
func (WarpedForest) Rainfall() float64 {
	return 0
}

// BlueSpores ...
func (WarpedForest) BlueSpores() float64 {
	return 0.25
}

// RedSpores ...
func (WarpedForest) RedSpores() float64 {
	return 0
}

// String ...
func (WarpedForest) String() string {
	return "warped_forest"
}

// EncodeBiome ...
func (WarpedForest) EncodeBiome() int {
	return 180
}
