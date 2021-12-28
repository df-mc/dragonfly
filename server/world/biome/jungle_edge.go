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

// Ash ...
func (JungleEdge) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (JungleEdge) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (JungleEdge) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (JungleEdge) RedSpores() float64 {
	return 0
}

// String ...
func (JungleEdge) String() string {
	return "jungle_edge"
}

// EncodeBiome ...
func (JungleEdge) EncodeBiome() int {
	return 23
}
