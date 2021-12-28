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

// String ...
func (SnowyTaigaHills) String() string {
	return "cold_taiga_hills"
}

// EncodeBiome ...
func (SnowyTaigaHills) EncodeBiome() int {
	return 31
}
