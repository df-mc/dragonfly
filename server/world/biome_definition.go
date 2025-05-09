package world

import (
	"bytes"
	_ "embed"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// BiomeDefinition ...
type BiomeDefinition struct {
	BiomeName string `nbt:"name"`
	BiomeID   uint16 `nbt:"id,omitempty"`

	Temperature      float32 `nbt:"temperature"`
	Downfall         float32 `nbt:"downfall"`
	RedSporeDensity  float32 `nbt:"redSporeDensity"`
	BlueSporeDensity float32 `nbt:"blueSporeDensity"`
	AshDensity       float32 `nbt:"ashDensity"`
	WhiteAshDensity  float32 `nbt:"whiteAshDensity"`

	Depth          float32 `nbt:"depth"`
	Scale          float32 `nbt:"scale"`
	MapWaterColour int32   `nbt:"mapWaterColour"`

	Rain bool     `nbt:"rain"`
	Tags []string `nbt:"tags"`
}

var (
	//go:embed biome_definitions.nbt
	biomeDefinitionData []byte
	biomeDefinitions    []BiomeDefinition
)

func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(biomeDefinitionData))
	if err := dec.Decode(&biomeDefinitions); err != nil {
		panic(err)
	}
}

// BiomeDefinitions ...
func BiomeDefinitions() []BiomeDefinition {
	return biomeDefinitions
}
