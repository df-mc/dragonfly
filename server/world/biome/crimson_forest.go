package biome

// CrimsonForest ...
type CrimsonForest struct{}

// Temperature ...
func (CrimsonForest) Temperature() float64 {
	return 2
}

// Rainfall ...
func (CrimsonForest) Rainfall() float64 {
	return 0
}

// Ash ...
func (CrimsonForest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (CrimsonForest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (CrimsonForest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (CrimsonForest) RedSpores() float64 {
	return 0.25
}

// String ...
func (CrimsonForest) String() string {
	return "crimson_forest"
}

// EncodeBiome ...
func (CrimsonForest) EncodeBiome() int {
	return 179
}
