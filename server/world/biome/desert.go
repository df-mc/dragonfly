package biome

// Desert ...
type Desert struct{}

// Temperature ...
func (Desert) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Desert) Rainfall() float64 {
	return 0
}

// Ash ...
func (Desert) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Desert) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Desert) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Desert) RedSpores() float64 {
	return 0
}

// String ...
func (Desert) String() string {
	return "desert"
}

// EncodeBiome ...
func (Desert) EncodeBiome() int {
	return 2
}
