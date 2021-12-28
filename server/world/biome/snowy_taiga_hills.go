package biome

// SnowyTaigaHills ...
type SnowyTaigaHills struct{}

// Temperature ...
func (SnowyTaigaHills) Temperature() float64 {
	return -0.5
}

// Rainfall ...
func (SnowyTaigaHills) Rainfall() float64 {
	return 0.4
}

// Ash ...
func (SnowyTaigaHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowyTaigaHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowyTaigaHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowyTaigaHills) RedSpores() float64 {
	return 0
}

// String ...
func (SnowyTaigaHills) String() string {
	return "cold_taiga_hills"
}

// EncodeBiome ...
func (SnowyTaigaHills) EncodeBiome() int {
	return 31
}
