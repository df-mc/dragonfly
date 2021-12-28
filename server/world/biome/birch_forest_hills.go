package biome

// BirchForestHills ...
type BirchForestHills struct{}

// Temperature ...
func (BirchForestHills) Temperature() float64 {
	return 0.6
}

// Rainfall ...
func (BirchForestHills) Rainfall() float64 {
	return 0.6
}

// Ash ...
func (BirchForestHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (BirchForestHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (BirchForestHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (BirchForestHills) RedSpores() float64 {
	return 0
}

// String ...
func (BirchForestHills) String() string {
	return "birch_forest_hills"
}

// EncodeBiome ...
func (BirchForestHills) EncodeBiome() int {
	return 28
}
