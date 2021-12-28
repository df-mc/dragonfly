package biome

// River ...
type River struct{}

// Temperature ...
func (River) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (River) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (River) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (River) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (River) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (River) RedSpores() float64 {
	return 0
}

// String ...
func (River) String() string {
	return "river"
}

// EncodeBiome ...
func (River) EncodeBiome() int {
	return 7
}
