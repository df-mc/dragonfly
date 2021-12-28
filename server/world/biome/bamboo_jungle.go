package biome

// BambooJungle ...
type BambooJungle struct{}

// Temperature ...
func (BambooJungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (BambooJungle) Rainfall() float64 {
	return 0.9
}

// String ...
func (BambooJungle) String() string {
	return "bamboo_jungle"
}

// EncodeBiome ...
func (BambooJungle) EncodeBiome() int {
	return 168
}
