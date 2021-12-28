package biome

// DeepOcean ...
type DeepOcean struct{}

// Temperature ...
func (DeepOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepOcean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (DeepOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DeepOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DeepOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DeepOcean) RedSpores() float64 {
	return 0
}

// String ...
func (DeepOcean) String() string {
	return "deep_ocean"
}

// EncodeBiome ...
func (DeepOcean) EncodeBiome() int {
	return 24
}
