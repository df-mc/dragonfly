package world

import (
	"bytes"
	_ "embed"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

var (
	//go:embed biome_definitions.nbt
	biomeDefinitionData []byte

	cachedBiomeDefinitions []protocol.BiomeDefinition
	cachedBiomeStringList  []string
)

func init() {
	type biomeNBT struct {
		BiomeName        string   `nbt:"name"`
		BiomeID          uint16   `nbt:"id,omitempty"`
		Temperature      float32  `nbt:"temperature"`
		Downfall         float32  `nbt:"downfall"`
		RedSporeDensity  float32  `nbt:"redSporeDensity"`
		BlueSporeDensity float32  `nbt:"blueSporeDensity"`
		AshDensity       float32  `nbt:"ashDensity"`
		WhiteAshDensity  float32  `nbt:"whiteAshDensity"`
		Depth            float32  `nbt:"depth"`
		Scale            float32  `nbt:"scale"`
		MapWaterColour   int32    `nbt:"mapWaterColour"`
		Rain             bool     `nbt:"rain"`
		Tags             []string `nbt:"tags"`
	}

	var rawBiomes []biomeNBT
	if err := nbt.NewDecoder(bytes.NewReader(biomeDefinitionData)).Decode(&rawBiomes); err != nil {
		panic(err)
	}

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

	encodedBiomes := make([]protocol.BiomeDefinition, 0, len(rawBiomes))
	for _, biome := range rawBiomes {
		nameIndex := intern(biome.BiomeName)

		tagIndices := make([]uint16, len(biome.Tags))
		for i, tag := range biome.Tags {
			tagIndices[i] = uint16(intern(tag))
		}

		var biomeID protocol.Optional[uint16]
		if biome.BiomeID > 0 {
			biomeID = protocol.Option[uint16](biome.BiomeID)
		}

		encodedBiomes = append(encodedBiomes, protocol.BiomeDefinition{
			NameIndex:        int16(nameIndex),
			BiomeID:          biomeID,
			Temperature:      biome.Temperature,
			Downfall:         biome.Downfall,
			RedSporeDensity:  biome.RedSporeDensity,
			BlueSporeDensity: biome.BlueSporeDensity,
			AshDensity:       biome.AshDensity,
			WhiteAshDensity:  biome.WhiteAshDensity,
			Depth:            biome.Depth,
			Scale:            biome.Scale,
			MapWaterColour:   biome.MapWaterColour,
			Rain:             biome.Rain,
			Tags:             protocol.Option[[]uint16](tagIndices),
		})
	}

	cachedBiomeDefinitions = encodedBiomes
	cachedBiomeStringList = internedStrings
}

// BiomeDefinitions returns cached biome data and string list.
func BiomeDefinitions() ([]protocol.BiomeDefinition, []string) {
	return cachedBiomeDefinitions, cachedBiomeStringList
}
