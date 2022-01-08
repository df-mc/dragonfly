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

// String ...
func (End) String() string {
	return "the_end"
}

// EncodeBiome ...
func (End) EncodeBiome() int {
	return 9
}
