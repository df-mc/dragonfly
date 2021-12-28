package biome

// JaggedPeaks ...
type JaggedPeaks struct{}

// Temperature ...
func (JaggedPeaks) Temperature() float64 {
	return -0.7
}

// Rainfall ...
func (JaggedPeaks) Rainfall() float64 {
	return 0.9
}

// String ...
func (JaggedPeaks) String() string {
	return "jagged_peaks"
}

// EncodeBiome ...
func (JaggedPeaks) EncodeBiome() int {
	return 182
}
