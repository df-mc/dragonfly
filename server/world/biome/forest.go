package biome

// Forest ...
type Forest struct{}

// Temperature ...
func (Forest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (Forest) Rainfall() float64 {
	return 0.8
}

// String ...
func (Forest) String() string {
	return "forest"
}

// EncodeBiome ...
func (Forest) EncodeBiome() int {
	return 4
}
