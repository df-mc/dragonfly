package biome

// DeepFrozenOcean ...
type DeepFrozenOcean struct{}

// Temperature ...
func (DeepFrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (DeepFrozenOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (DeepFrozenOcean) String() string {
	return "deep_frozen_ocean"
}

// EncodeBiome ...
func (DeepFrozenOcean) EncodeBiome() int {
	return 46
}
