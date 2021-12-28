package biome

// Beach ...
type Beach struct{}

// Temperature ...
func (Beach) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Beach) Rainfall() float64 {
	return 0.4
}

// Ash ...
func (Beach) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Beach) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Beach) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Beach) RedSpores() float64 {
	return 0
}

// String ...
func (Beach) String() string {
	return "beach"
}

// EncodeBiome ...
func (Beach) EncodeBiome() int {
	return 16
}
