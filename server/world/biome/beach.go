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

// String ...
func (Beach) String() string {
	return "beach"
}

// EncodeBiome ...
func (Beach) EncodeBiome() int {
	return 16
}
