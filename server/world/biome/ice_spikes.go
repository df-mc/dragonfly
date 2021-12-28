package biome

// IceSpikes ...
type IceSpikes struct{}

// Temperature ...
func (IceSpikes) Temperature() float64 {
	return 0
}

// Rainfall ...
func (IceSpikes) Rainfall() float64 {
	return 1
}

// String ...
func (IceSpikes) String() string {
	return "ice_plains_spikes"
}

// EncodeBiome ...
func (IceSpikes) EncodeBiome() int {
	return 140
}
