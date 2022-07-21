package biome

// WindsweptSavanna ...
type WindsweptSavanna struct{}

// Temperature ...
func (WindsweptSavanna) Temperature() float64 {
	return 1.1
}

// Rainfall ...
func (WindsweptSavanna) Rainfall() float64 {
	return 0.5
}

// String ...
func (WindsweptSavanna) String() string {
	return "savanna_mutated"
}

// EncodeBiome ...
func (WindsweptSavanna) EncodeBiome() int {
	return 163
}
