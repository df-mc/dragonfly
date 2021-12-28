package biome

// Plains ...
type Plains struct{}

// Temperature ...
func (Plains) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Plains) Rainfall() float64 {
	return 0.4
}

// Ash ...
func (Plains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Plains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Plains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Plains) RedSpores() float64 {
	return 0
}

// String ...
func (Plains) String() string {
	return "plains"
}

// EncodeBiome ...
func (Plains) EncodeBiome() int {
	return 1
}
