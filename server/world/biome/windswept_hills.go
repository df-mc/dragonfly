package biome

// WindsweptHills ...
type WindsweptHills struct{}

// Temperature ...
func (WindsweptHills) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptHills) Rainfall() float64 {
	return 0.3
}

// String ...
func (WindsweptHills) String() string {
	return "extreme_hills"
}

// EncodeBiome ...
func (WindsweptHills) EncodeBiome() int {
	return 3
}
