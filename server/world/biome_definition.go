package world

import (
	"encoding/binary"
	"image/color"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

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

// MaxVanillaBiomeID ...
const MaxVanillaBiomeID = 193

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

		var biomeID protocol.Optional[uint16]
		id := b.EncodeBiome()
		if id > MaxVanillaBiomeID {
			biomeID = protocol.Option[uint16](uint16(id))
		}

		def := protocol.BiomeDefinition{
			NameIndex:      int16(nameIndex),
			BiomeID:        biomeID,
			Temperature:    float32(b.Temperature()),
			Downfall:       float32(b.Rainfall()),
			Depth:          float32(b.Depth()),
			Scale:          float32(b.Scale()),
			MapWaterColour: int32FromRGBA(b.WaterColour()),
			Rain:           b.Rainfall() > 0,
			Tags:           protocol.Option[[]uint16](tagIndices),
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

// int32FromRGBA converts a color.RGBA into an int32. These int32s are present in things such as signs and dyed leather armour.
func int32FromRGBA(x color.RGBA) int32 {
	if x.R == 0 && x.G == 0 && x.B == 0 {
		// Default to black colour. The default (0x000000) is a transparent colour. Text with this colour will not show
		// up on the sign.
		return int32(-0x1000000)
	}
	return int32(binary.BigEndian.Uint32([]byte{x.A, x.R, x.G, x.B}))
}
