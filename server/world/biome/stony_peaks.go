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

// String ...
func (StonyPeaks) String() string {
	return "Stony Peaks"
}

// EncodeBiome ...
func (StonyPeaks) EncodeBiome() int {
	return 189
}
