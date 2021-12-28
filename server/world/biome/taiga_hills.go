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

// Ash ...
func (TaigaHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (TaigaHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (TaigaHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (TaigaHills) RedSpores() float64 {
	return 0
}

// String ...
func (TaigaHills) String() string {
	return "taiga_hills"
}

// EncodeBiome ...
func (TaigaHills) EncodeBiome() int {
	return 19
}
