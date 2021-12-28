package biome

// WarmOcean ...
type WarmOcean struct{}

// Temperature ...
func (WarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (WarmOcean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (WarmOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (WarmOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (WarmOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (WarmOcean) RedSpores() float64 {
	return 0
}

// String ...
func (WarmOcean) String() string {
	return "warm_ocean"
}

// EncodeBiome ...
func (WarmOcean) EncodeBiome() int {
	return 40
}
