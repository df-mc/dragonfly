package biome

// WoodedHills ...
type WoodedHills struct{}

// Temperature ...
func (WoodedHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (WoodedHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (WoodedHills) String() string {
	return "forest_hills"
}

// EncodeBiome ...
func (WoodedHills) EncodeBiome() int {
	return 18
}
