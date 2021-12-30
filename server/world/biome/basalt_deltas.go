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
func (BasaltDeltas) Ash() (ash float64, whiteAsh float64) {
	return 0, 2
}

// String ...
func (BasaltDeltas) String() string {
	return "basalt_deltas"
}

// EncodeBiome ...
func (BasaltDeltas) EncodeBiome() int {
	return 181
}
