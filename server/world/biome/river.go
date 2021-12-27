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

// String ...
func (River) String() string {
	return "River"
}

// EncodeBiome ...
func (River) EncodeBiome() int {
	return 7
}
