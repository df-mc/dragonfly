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

// Ash ...
func (DeepFrozenOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DeepFrozenOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DeepFrozenOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DeepFrozenOcean) RedSpores() float64 {
	return 0
}

// String ...
func (DeepFrozenOcean) String() string {
	return "deep_frozen_ocean"
}

// EncodeBiome ...
func (DeepFrozenOcean) EncodeBiome() int {
	return 46
}
