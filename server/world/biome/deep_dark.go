package biome

// DeepDark ...
type DeepDark struct{}

// Temperature ...
func (DeepDark) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (DeepDark) Rainfall() float64 {
	return 0.4
}

// String ...
func (DeepDark) String() string {
	return "deep_dark"
}

// EncodeBiome ...
func (DeepDark) EncodeBiome() int {
	return 190
}
