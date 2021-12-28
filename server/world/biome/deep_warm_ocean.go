package biome

// DeepWarmOcean ...
type DeepWarmOcean struct{}

// Temperature ...
func (DeepWarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepWarmOcean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (DeepWarmOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DeepWarmOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DeepWarmOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DeepWarmOcean) RedSpores() float64 {
	return 0
}

// String ...
func (DeepWarmOcean) String() string {
	return "deep_warm_ocean"
}

// EncodeBiome ...
func (DeepWarmOcean) EncodeBiome() int {
	return 43
}
