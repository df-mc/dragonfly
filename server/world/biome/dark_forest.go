package biome

// DarkForest ...
type DarkForest struct{}

// Temperature ...
func (DarkForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (DarkForest) Rainfall() float64 {
	return 0.8
}

// String ...
func (DarkForest) String() string {
	return "roofed_forest"
}

// EncodeBiome ...
func (DarkForest) EncodeBiome() int {
	return 29
}
