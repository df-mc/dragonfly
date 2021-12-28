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

// Ash ...
func (GravellyMountainsPlus) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (GravellyMountainsPlus) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (GravellyMountainsPlus) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (GravellyMountainsPlus) RedSpores() float64 {
	return 0
}

// String ...
func (GravellyMountainsPlus) String() string {
	return "extreme_hills_plus_trees_mutated"
}

// EncodeBiome ...
func (GravellyMountainsPlus) EncodeBiome() int {
	return 162
}
