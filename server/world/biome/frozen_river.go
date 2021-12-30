package biome

// FrozenRiver ...
type FrozenRiver struct{}

// Temperature ...
func (FrozenRiver) Temperature() float64 {
	return 0
}

// Rainfall ...
func (FrozenRiver) Rainfall() float64 {
	return 0.5
}

// String ...
func (FrozenRiver) String() string {
	return "frozen_river"
}

// EncodeBiome ...
func (FrozenRiver) EncodeBiome() int {
	return 11
}
