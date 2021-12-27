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

// String ...
func (BasaltDeltas) String() string {
	return "Basalt Deltas"
}

// EncodeBiome ...
func (BasaltDeltas) EncodeBiome() int {
	return 181
}
