package biome

// GravellyMountainsPlus ...
type GravellyMountainsPlus struct{}

// Temperature ...
func (GravellyMountainsPlus) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (GravellyMountainsPlus) Rainfall() float64 {
	return 0.3
}

// String ...
func (GravellyMountainsPlus) String() string {
	return "extreme_hills_plus_trees_mutated"
}

// EncodeBiome ...
func (GravellyMountainsPlus) EncodeBiome() int {
	return 162
}
