package biome

// TaigaMountains ...
type TaigaMountains struct{}

// Temperature ...
func (TaigaMountains) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (TaigaMountains) Rainfall() float64 {
	return 0.8
}

// String ...
func (TaigaMountains) String() string {
	return "taiga_mutated"
}

// EncodeBiome ...
func (TaigaMountains) EncodeBiome() int {
	return 133
}
