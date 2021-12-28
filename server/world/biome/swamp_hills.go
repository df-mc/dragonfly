package biome

// SwampHills ...
type SwampHills struct{}

// Temperature ...
func (SwampHills) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (SwampHills) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (SwampHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SwampHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SwampHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SwampHills) RedSpores() float64 {
	return 0
}

// String ...
func (SwampHills) String() string {
	return "swampland_mutated"
}

// EncodeBiome ...
func (SwampHills) EncodeBiome() int {
	return 134
}
