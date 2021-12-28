package biome

// StonyPeaks ...
type StonyPeaks struct{}

// Temperature ...
func (StonyPeaks) Temperature() float64 {
	return 1
}

// Rainfall ...
func (StonyPeaks) Rainfall() float64 {
	return 0.3
}

// Ash ...
func (StonyPeaks) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (StonyPeaks) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (StonyPeaks) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (StonyPeaks) RedSpores() float64 {
	return 0
}

// String ...
func (StonyPeaks) String() string {
	return "stony_peaks"
}

// EncodeBiome ...
func (StonyPeaks) EncodeBiome() int {
	return 189
}
