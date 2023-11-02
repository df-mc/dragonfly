package biome

// CherryGrove ...
type CherryGrove struct{}

// Temperature ...
func (CherryGrove) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (CherryGrove) Rainfall() float64 {
	return 0.8
}

// String ...
func (CherryGrove) String() string {
	return "cherry_grove"
}

// EncodeBiome ...
func (CherryGrove) EncodeBiome() int {
	return 192
}
