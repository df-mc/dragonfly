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

// Ash ...
func (FrozenRiver) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (FrozenRiver) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (FrozenRiver) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (FrozenRiver) RedSpores() float64 {
	return 0
}

// String ...
func (FrozenRiver) String() string {
	return "frozen_river"
}

// EncodeBiome ...
func (FrozenRiver) EncodeBiome() int {
	return 11
}
