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

// Ash ...
func (GiantTreeTaigaHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (GiantTreeTaigaHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (GiantTreeTaigaHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (GiantTreeTaigaHills) RedSpores() float64 {
	return 0
}

// String ...
func (GiantTreeTaigaHills) String() string {
	return "mega_taiga_hills"
}

// EncodeBiome ...
func (GiantTreeTaigaHills) EncodeBiome() int {
	return 33
}
