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

// String ...
func (WarpedForest) String() string {
	return "Warped Forest"
}

// EncodeBiome ...
func (WarpedForest) EncodeBiome() int {
	return 180
}
