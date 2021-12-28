package biome

// BirchForest ...
type BirchForest struct{}

// Temperature ...
func (BirchForest) Temperature() float64 {
	return 0.6
}

// Rainfall ...
func (BirchForest) Rainfall() float64 {
	return 0.6
}

// String ...
func (BirchForest) String() string {
	return "birch_forest"
}

// EncodeBiome ...
func (BirchForest) EncodeBiome() int {
	return 27
}
