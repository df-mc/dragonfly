package biome

// GiantTreeTaigaHills ...
type GiantTreeTaigaHills struct{}

// Temperature ...
func (GiantTreeTaigaHills) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (GiantTreeTaigaHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (GiantTreeTaigaHills) String() string {
	return "mega_taiga_hills"
}

// EncodeBiome ...
func (GiantTreeTaigaHills) EncodeBiome() int {
	return 33
}
