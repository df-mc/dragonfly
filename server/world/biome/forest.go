package biome

import "image/color"

// Forest ...
type Forest struct{}

// Temperature ...
func (Forest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (Forest) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (Forest) Depth() float64 {
	return 0.1
}

// Scale ...
func (Forest) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (Forest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1e, G: 0x97, B: 0xf2, A: 0xa5}
}

// Tags ...
func (Forest) Tags() []string {
	return []string{"animal", "forest", "monster", "overworld", "bee_habitat"}
}

// String ...
func (Forest) String() string {
	return "forest"
}

// EncodeBiome ...
func (Forest) EncodeBiome() int {
	return 4
}
