package world

import (
	"encoding/binary"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

var (
	// maxVanillaBiomeID is the highest ID used by vanilla biomes.
	maxVanillaBiomeID int
)

// finaliseBiomeRegistry is called after all vanilla biomes have been registered.
// It sets maxVanillaBiomeID to the highest ID found among them.
// noinspection GoUnusedFunction
//
//lint:ignore U1000 Function is used through compiler directives.
func finaliseBiomeRegistry() {
	for _, b := range biomes {
		id := b.EncodeBiome()
		if id > maxVanillaBiomeID {
			maxVanillaBiomeID = id
		}
	}
}

// ashyBiome represents a biome that has any form of ash.
type ashyBiome interface {
	// Ash returns the ash and white ash of the biome.
	Ash() (ash float64, whiteAsh float64)
}

// sporingBiome represents a biome that has blue or red spores.
type sporingBiome interface {
	// Spores returns the blue and red spores of the biome.
	Spores() (blueSpores float64, redSpores float64)
}

// BiomeDefinitions returns the list of biome definitions along with the associated StringList.
func BiomeDefinitions() ([]protocol.BiomeDefinition, []string) {
	var (
		internedStrings     []string
		internedStringIndex = make(map[string]int)
	)

	intern := func(s string) int {
		if index, exists := internedStringIndex[s]; exists {
			return index
		}
		index := len(internedStrings)
		internedStrings = append(internedStrings, s)
		internedStringIndex[s] = index
		return index
	}

	encodedBiomes := make([]protocol.BiomeDefinition, 0, len(biomes))
	for _, b := range biomes {
		nameIndex := intern(b.String())

		tags := b.Tags()
		tagIndices := make([]uint16, len(tags))
		for i, tag := range tags {
			tagIndices[i] = uint16(intern(tag))
		}

		var biomeID int16 = -1
		id := b.EncodeBiome()
		if id > maxVanillaBiomeID {
			biomeID = int16(id)
		}

		def := protocol.BiomeDefinition{
			NameIndex:   int16(nameIndex),
			BiomeID:     biomeID,
			Temperature: float32(b.Temperature()),
			Downfall:    float32(b.Rainfall()),
			Depth:       float32(b.Depth()),
			Scale:       float32(b.Scale()),
			MapWaterColour: int32(binary.BigEndian.Uint32([]byte{
				b.WaterColour().A,
				b.WaterColour().R,
				b.WaterColour().G,
				b.WaterColour().B,
			})),
			Rain: b.Rainfall() > 0,
			Tags: protocol.Option[[]uint16](tagIndices),
		}

		if a, ok := b.(ashyBiome); ok {
			ash, whiteAsh := a.Ash()
			def.AshDensity = float32(ash)
			def.WhiteAshDensity = float32(whiteAsh)
		}

		if s, ok := b.(sporingBiome); ok {
			blueSpores, redSpores := s.Spores()
			def.BlueSporeDensity = float32(blueSpores)
			def.RedSporeDensity = float32(redSpores)
		}

		encodedBiomes = append(encodedBiomes, def)
	}

	return encodedBiomes, internedStrings
}
