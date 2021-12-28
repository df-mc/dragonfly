package biome

// Taiga ...
type Taiga struct{}

// Temperature ...
func (Taiga) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (Taiga) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (Taiga) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Taiga) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Taiga) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Taiga) RedSpores() float64 {
	return 0
}

// String ...
func (Taiga) String() string {
	return "taiga"
}

// EncodeBiome ...
func (Taiga) EncodeBiome() int {
	return 5
}
