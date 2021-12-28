package biome

// FrozenOcean ...
type FrozenOcean struct{}

// Temperature ...
func (FrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (FrozenOcean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (FrozenOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (FrozenOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (FrozenOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (FrozenOcean) RedSpores() float64 {
	return 0
}

// String ...
func (FrozenOcean) String() string {
	return "frozen_ocean"
}

// EncodeBiome ...
func (FrozenOcean) EncodeBiome() int {
	return 10
}
