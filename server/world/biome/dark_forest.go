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

// Ash ...
func (DarkForest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DarkForest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DarkForest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DarkForest) RedSpores() float64 {
	return 0
}

// String ...
func (DarkForest) String() string {
	return "roofed_forest"
}

// EncodeBiome ...
func (DarkForest) EncodeBiome() int {
	return 29
}
