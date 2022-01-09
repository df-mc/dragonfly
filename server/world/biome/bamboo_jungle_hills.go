package biome

// BambooJungleHills ...
type BambooJungleHills struct{}

// Temperature ...
func (BambooJungleHills) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (BambooJungleHills) Rainfall() float64 {
	return 0.9
}

// String ...
func (BambooJungleHills) String() string {
	return "bamboo_jungle_hills"
}

// EncodeBiome ...
func (BambooJungleHills) EncodeBiome() int {
	return 169
}
