package biome

// WindsweptGravellyHills ...
type WindsweptGravellyHills struct{}

// Temperature ...
func (WindsweptGravellyHills) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptGravellyHills) Rainfall() float64 {
	return 0.3
}

// String ...
func (WindsweptGravellyHills) String() string {
	return "extreme_hills_mutated"
}

// EncodeBiome ...
func (WindsweptGravellyHills) EncodeBiome() int {
	return 131
}
