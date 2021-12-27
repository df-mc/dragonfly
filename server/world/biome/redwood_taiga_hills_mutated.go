package biome

// GiantSpruceTaigaHills ...
type GiantSpruceTaigaHills struct{}

// Temperature ...
func (GiantSpruceTaigaHills) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (GiantSpruceTaigaHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (GiantSpruceTaigaHills) String() string {
	return "Giant Spruce Taiga Hills"
}

// EncodeBiome ...
func (GiantSpruceTaigaHills) EncodeBiome() int {
	return 161
}
