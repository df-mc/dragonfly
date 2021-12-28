package biome

// ModifiedJungleEdge ...
type ModifiedJungleEdge struct{}

// Temperature ...
func (ModifiedJungleEdge) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (ModifiedJungleEdge) Rainfall() float64 {
	return 0.8
}

// String ...
func (ModifiedJungleEdge) String() string {
	return "jungle_edge_mutated"
}

// EncodeBiome ...
func (ModifiedJungleEdge) EncodeBiome() int {
	return 151
}
