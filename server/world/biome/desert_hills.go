package biome

// DesertHills ...
type DesertHills struct{}

// Temperature ...
func (DesertHills) Temperature() float64 {
	return 2
}

// Rainfall ...
func (DesertHills) Rainfall() float64 {
	return 0
}

// String ...
func (DesertHills) String() string {
	return "desert_hills"
}

// EncodeBiome ...
func (DesertHills) EncodeBiome() int {
	return 17
}
