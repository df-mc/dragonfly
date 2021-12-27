package biome

// JungleEdge ...
type JungleEdge struct{}

// Temperature ...
func (JungleEdge) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (JungleEdge) Rainfall() float64 {
	return 0.8
}

// String ...
func (JungleEdge) String() string {
	return "Jungle Edge"
}

// EncodeBiome ...
func (JungleEdge) EncodeBiome() int {
	return 23
}
