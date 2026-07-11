package biome

import "image/color"

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

// Depth ...
func (Beach) Depth() float64 {
	return 0
}

// Scale ...
func (Beach) Scale() float64 {
	return 0.025
}

// WaterColour ...
func (Beach) WaterColour() color.RGBA {
	return color.RGBA{R: 0x15, G: 0x7c, B: 0xab, A: 0xa5}
}

// Tags ...
func (Beach) Tags() []string {
	return []string{"beach", "monster", "overworld", "warm"}
}

// String ...
func (Beach) String() string {
	return "beach"
}

// EncodeBiome ...
func (Beach) EncodeBiome() int {
	return 16
}
