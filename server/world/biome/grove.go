package biome

// Grove ...
type Grove struct{}

// Temperature ...
func (Grove) Temperature() float64 {
	return -0.2
}

// Rainfall ...
func (Grove) Rainfall() float64 {
	return 0.8
}

// String ...
func (Grove) String() string {
	return "grove"
}

// EncodeBiome ...
func (Grove) EncodeBiome() int {
	return 185
}
