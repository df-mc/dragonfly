package biome

// ErodedBadlands ...
type ErodedBadlands struct{}

// Temperature ...
func (ErodedBadlands) Temperature() float64 {
	return 2
}

// Rainfall ...
func (ErodedBadlands) Rainfall() float64 {
	return 0
}

// String ...
func (ErodedBadlands) String() string {
	return "mesa_bryce"
}

// EncodeBiome ...
func (ErodedBadlands) EncodeBiome() int {
	return 165
}
