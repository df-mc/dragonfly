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

// String ...
func (Taiga) String() string {
	return "taiga"
}

// EncodeBiome ...
func (Taiga) EncodeBiome() int {
	return 5
}
