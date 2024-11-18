package biome

// PaleGarden ...
type PaleGarden struct{}

// Temperature ...
func (PaleGarden) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (PaleGarden) Rainfall() float64 {
	return 0.8
}

// String ...
func (PaleGarden) String() string {
	return "pale_garden"
}

// EncodeBiome ...
func (PaleGarden) EncodeBiome() int {
	return 193
}
