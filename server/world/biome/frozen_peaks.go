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

// Ash ...
func (FrozenPeaks) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (FrozenPeaks) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (FrozenPeaks) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (FrozenPeaks) RedSpores() float64 {
	return 0
}

// String ...
func (FrozenPeaks) String() string {
	return "frozen_peaks"
}

// EncodeBiome ...
func (FrozenPeaks) EncodeBiome() int {
	return 183
}
