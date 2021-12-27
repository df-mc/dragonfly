package biome

// JungleHills ...
type JungleHills struct{}

// Temperature ...
func (JungleHills) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (JungleHills) Rainfall() float64 {
	return 0.9
}

// String ...
func (JungleHills) String() string {
	return "Jungle Hills"
}

// EncodeBiome ...
func (JungleHills) EncodeBiome() int {
	return 22
}
