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
	return "Snowy Taiga Hills"
}

// EncodeBiome ...
func (SnowyTaigaHills) EncodeBiome() int {
	return 31
}
