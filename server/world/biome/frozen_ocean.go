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

// String ...
func (FrozenOcean) String() string {
	return "frozen_ocean"
}

// EncodeBiome ...
func (FrozenOcean) EncodeBiome() int {
	return 10
}
