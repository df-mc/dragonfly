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

// Ash ...
func (BirchForest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (BirchForest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (BirchForest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (BirchForest) RedSpores() float64 {
	return 0
}

// String ...
func (BirchForest) String() string {
	return "birch_forest"
}

// EncodeBiome ...
func (BirchForest) EncodeBiome() int {
	return 27
}
