package biome

// MountainEdge ...
type MountainEdge struct{}

// Temperature ...
func (MountainEdge) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (MountainEdge) Rainfall() float64 {
	return 0.3
}

// Ash ...
func (MountainEdge) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (MountainEdge) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (MountainEdge) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (MountainEdge) RedSpores() float64 {
	return 0
}

// String ...
func (MountainEdge) String() string {
	return "extreme_hills_edge"
}

// EncodeBiome ...
func (MountainEdge) EncodeBiome() int {
	return 20
}
