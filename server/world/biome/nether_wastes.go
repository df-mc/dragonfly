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

// Ash ...
func (NetherWastes) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (NetherWastes) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (NetherWastes) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (NetherWastes) RedSpores() float64 {
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
