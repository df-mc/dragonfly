package biome

// End ...
type End struct{}

// Temperature ...
func (End) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (End) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (End) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (End) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (End) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (End) RedSpores() float64 {
	return 0
}

// String ...
func (End) String() string {
	return "the_end"
}

// EncodeBiome ...
func (End) EncodeBiome() int {
	return 9
}
