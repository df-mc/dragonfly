package world

import "image/color"

// Biome is a region in a world with distinct geographical features, flora, temperatures, humidity ratings,
// and sky, water, grass and foliage colours.
type Biome interface {
	// Temperature returns the temperature of the biome.
	Temperature() float64
	// Rainfall returns the rainfall of the biome.
	Rainfall() float64
	// Depth returns the depth of the biome.
	Depth() float64
	// Scale returns the scale of the biome.
	Scale() float64
	// WaterColour returns the water colour of the biome.
	WaterColour() color.RGBA
	// Tags returns the tags for the biome.
	Tags() []string
	// String returns the biome name as a string.
	String() string
	// EncodeBiome encodes the biome into an int value that is used to identify the biome over the network.
	EncodeBiome() int
}

// biomes holds a map of id => Biome to be used for looking up the biome by an ID. It is registered
// to when calling RegisterBiome.
var biomes = map[int]Biome{}

var biomeByName = map[string]Biome{}

// RegisterBiome registers a biome to the map so that it can be saved and loaded with the world.
func RegisterBiome(b Biome) {
	id := b.EncodeBiome()
	if _, ok := biomes[id]; ok {
		panic("cannot register the same biome (" + b.String() + ") twice")
	}
	biomes[id] = b
	biomeByName[b.String()] = b
}

// BiomeByID looks up a biome by the ID and returns it if found.
func BiomeByID(id int) (Biome, bool) {
	e, ok := biomes[id]
	return e, ok
}

// BiomeByName looks up a biome by the name and returns it if found.
func BiomeByName(name string) (Biome, bool) {
	e, ok := biomeByName[name]
	return e, ok
}

// Biomes returns a slice of all registered biomes.
func Biomes() []Biome {
	bs := make([]Biome, 0, len(biomes))
	for _, b := range biomes {
		bs = append(bs, b)
	}
	return bs
}

// ocean returns an ocean biome.
func ocean() Biome {
	o, _ := BiomeByID(0)
	return o
}
