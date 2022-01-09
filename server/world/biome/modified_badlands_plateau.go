package biome

// ModifiedBadlandsPlateau ...
type ModifiedBadlandsPlateau struct{}

// Temperature ...
func (ModifiedBadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (ModifiedBadlandsPlateau) Rainfall() float64 {
	return 0
}

// String ...
func (ModifiedBadlandsPlateau) String() string {
	return "mesa_plateau_mutated"
}

// EncodeBiome ...
func (ModifiedBadlandsPlateau) EncodeBiome() int {
	return 166
}
