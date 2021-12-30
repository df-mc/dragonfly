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

// Spores ...
func (WarpedForest) Spores() (blueSpores float64, redSpores float64) {
	return 0.25, 0
}

// String ...
func (WarpedForest) String() string {
	return "warped_forest"
}

// EncodeBiome ...
func (WarpedForest) EncodeBiome() int {
	return 180
}
