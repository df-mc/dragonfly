package biome

// MangroveSwamp ...
type MangroveSwamp struct{}

// Temperature ...
func (MangroveSwamp) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (MangroveSwamp) Rainfall() float64 {
	return 0.9
}

// String ...
func (MangroveSwamp) String() string {
	return "mangrove_swamp"
}

// EncodeBiome ...
func (MangroveSwamp) EncodeBiome() int {
	return 191
}
