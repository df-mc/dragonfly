package biome

// BasaltDeltas ...
type BasaltDeltas struct{}

// Temperature ...
func (BasaltDeltas) Temperature() float64 {
	return 2
}

// Rainfall ...
func (BasaltDeltas) Rainfall() float64 {
	return 0
}

// Ash ...
func (BasaltDeltas) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (BasaltDeltas) WhiteAsh() float64 {
	return 2
}

// BlueSpores ...
func (BasaltDeltas) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (BasaltDeltas) RedSpores() float64 {
	return 0
}

// String ...
func (BasaltDeltas) String() string {
	return "basalt_deltas"
}

// EncodeBiome ...
func (BasaltDeltas) EncodeBiome() int {
	return 181
}
