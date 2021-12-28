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

// Ash ...
func (JaggedPeaks) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (JaggedPeaks) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (JaggedPeaks) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (JaggedPeaks) RedSpores() float64 {
	return 0
}

// String ...
func (JaggedPeaks) String() string {
	return "jagged_peaks"
}

// EncodeBiome ...
func (JaggedPeaks) EncodeBiome() int {
	return 182
}
