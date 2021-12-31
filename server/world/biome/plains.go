package biome

// Plains ...
type Plains struct{}

// Temperature ...
func (Plains) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Plains) Rainfall() float64 {
	return 0.4
}

// String ...
func (Plains) String() string {
	return "plains"
}

// EncodeBiome ...
func (Plains) EncodeBiome() int {
	return 1
}
