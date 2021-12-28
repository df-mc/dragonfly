package biome

// FrozenPeaks ...
type FrozenPeaks struct{}

// Temperature ...
func (FrozenPeaks) Temperature() float64 {
	return -0.7
}

// Rainfall ...
func (FrozenPeaks) Rainfall() float64 {
	return 0.9
}

// String ...
func (FrozenPeaks) String() string {
	return "frozen_peaks"
}

// EncodeBiome ...
func (FrozenPeaks) EncodeBiome() int {
	return 183
}
