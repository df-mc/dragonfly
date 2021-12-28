package biome

// ColdOcean ...
type ColdOcean struct{}

// Temperature ...
func (ColdOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (ColdOcean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (ColdOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (ColdOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (ColdOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (ColdOcean) RedSpores() float64 {
	return 0
}

// String ...
func (ColdOcean) String() string {
	return "cold_ocean"
}

// EncodeBiome ...
func (ColdOcean) EncodeBiome() int {
	return 42
}
