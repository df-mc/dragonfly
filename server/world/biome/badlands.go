package biome

// Badlands ...
type Badlands struct{}

// Temperature ...
func (Badlands) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Badlands) Rainfall() float64 {
	return 0
}

// String ...
func (Badlands) String() string {
	return "mesa"
}

// EncodeBiome ...
func (Badlands) EncodeBiome() int {
	return 37
}
