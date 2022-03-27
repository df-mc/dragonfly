package world

// Biome is a region in a world with distinct geographical features, flora, temperatures, humidity ratings,
// and sky, water, grass and foliage colors.
type Biome interface {
	// Temperature returns the temperature of the biome.
	Temperature() float64
	// Rainfall returns the rainfall of the biome.
	Rainfall() float64
	// String returns the biome name as a string.
	String() string
	// EncodeBiome encodes the biome into an int value that is used to identify the biome over the network.
	EncodeBiome() int
}

// biomes holds a map of id => Biome to be used for looking up the biome by an ID. It is registered
// to when calling RegisterBiome.
var biomes = map[int]Biome{}

// RegisterBiome registers a biome to the map so that it can be saved and loaded with the world.
func RegisterBiome(b Biome) {
	id := b.EncodeBiome()
	if _, ok := biomes[id]; ok {
		panic("cannot register the same biome (" + b.String() + ") twice")
	}
	biomes[id] = b
}

// BiomeByID looks up a biome by the ID and returns it if found.
func BiomeByID(id int) (Biome, bool) {
	e, ok := biomes[id]
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
