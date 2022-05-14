package biome

// TaigaHills ...
type TaigaHills struct{}

// Temperature ...
func (TaigaHills) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (TaigaHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (TaigaHills) String() string {
	return "taiga_hills"
}

// EncodeBiome ...
func (TaigaHills) EncodeBiome() int {
	return 19
}
