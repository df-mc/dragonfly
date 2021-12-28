package biome

// BadlandsPlateau ...
type BadlandsPlateau struct{}

// Temperature ...
func (BadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (BadlandsPlateau) Rainfall() float64 {
	return 0
}

// String ...
func (BadlandsPlateau) String() string {
	return "mesa_plateau"
}

// EncodeBiome ...
func (BadlandsPlateau) EncodeBiome() int {
	return 38
}
