package biome

// NetherWastes ...
type NetherWastes struct{}

// Temperature ...
func (NetherWastes) Temperature() float64 {
	return 2
}

// Rainfall ...
func (NetherWastes) Rainfall() float64 {
	return 0
}

// String ...
func (NetherWastes) String() string {
	return "hell"
}

// EncodeBiome ...
func (NetherWastes) EncodeBiome() int {
	return 8
}
