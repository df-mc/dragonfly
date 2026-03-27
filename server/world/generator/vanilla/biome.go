package vanilla

import (
	"sort"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	gen "github.com/df-mc/dragonfly/server/world/generator/vanilla/gen"
)

var biomeRuntimeIDs = map[gen.Biome]uint32{
	gen.BiomeOcean:                uint32(biome.Ocean{}.EncodeBiome()),
	gen.BiomePlains:               uint32(biome.Plains{}.EncodeBiome()),
	gen.BiomeDesert:               uint32(biome.Desert{}.EncodeBiome()),
	gen.BiomeWindsweptHills:       uint32(biome.WindsweptHills{}.EncodeBiome()),
	gen.BiomeForest:               uint32(biome.Forest{}.EncodeBiome()),
	gen.BiomeTaiga:                uint32(biome.Taiga{}.EncodeBiome()),
	gen.BiomeSwamp:                uint32(biome.Swamp{}.EncodeBiome()),
	gen.BiomeRiver:                uint32(biome.River{}.EncodeBiome()),
	gen.BiomeFrozenOcean:          uint32(biome.FrozenOcean{}.EncodeBiome()),
	gen.BiomeFrozenRiver:          uint32(biome.FrozenRiver{}.EncodeBiome()),
	gen.BiomeSnowyPlains:          uint32(biome.SnowyPlains{}.EncodeBiome()),
	gen.BiomeSnowyMountains:       uint32(biome.SnowyMountains{}.EncodeBiome()),
	gen.BiomeMushroomFields:       uint32(biome.MushroomFields{}.EncodeBiome()),
	gen.BiomeBeach:                uint32(biome.Beach{}.EncodeBiome()),
	gen.BiomeJungle:               uint32(biome.Jungle{}.EncodeBiome()),
	gen.BiomeDeepOcean:            uint32(biome.DeepOcean{}.EncodeBiome()),
	gen.BiomeStonyShore:           uint32(biome.StonyShore{}.EncodeBiome()),
	gen.BiomeSnowyBeach:           uint32(biome.SnowyBeach{}.EncodeBiome()),
	gen.BiomeBirchForest:          uint32(biome.BirchForest{}.EncodeBiome()),
	gen.BiomeDarkForest:           uint32(biome.DarkForest{}.EncodeBiome()),
	gen.BiomeSnowyTaiga:           uint32(biome.SnowyTaiga{}.EncodeBiome()),
	gen.BiomeOldGrowthPineTaiga:   uint32(biome.OldGrowthPineTaiga{}.EncodeBiome()),
	gen.BiomeWindsweptForest:      uint32(biome.WindsweptForest{}.EncodeBiome()),
	gen.BiomeSavanna:              uint32(biome.Savanna{}.EncodeBiome()),
	gen.BiomeSavannaPlateau:       uint32(biome.SavannaPlateau{}.EncodeBiome()),
	gen.BiomeBadlands:             uint32(biome.Badlands{}.EncodeBiome()),
	gen.BiomeWoodedBadlands:       uint32(biome.WoodedBadlandsPlateau{}.EncodeBiome()),
	gen.BiomeWarmOcean:            uint32(biome.WarmOcean{}.EncodeBiome()),
	gen.BiomeLukewarmOcean:        uint32(biome.LukewarmOcean{}.EncodeBiome()),
	gen.BiomeColdOcean:            uint32(biome.ColdOcean{}.EncodeBiome()),
	gen.BiomeDeepLukewarmOcean:    uint32(biome.DeepLukewarmOcean{}.EncodeBiome()),
	gen.BiomeDeepColdOcean:        uint32(biome.DeepColdOcean{}.EncodeBiome()),
	gen.BiomeDeepFrozenOcean:      uint32(biome.DeepFrozenOcean{}.EncodeBiome()),
	gen.BiomeSunflowerPlains:      uint32(biome.SunflowerPlains{}.EncodeBiome()),
	gen.BiomeGravellyMountains:    uint32(biome.WindsweptGravellyHills{}.EncodeBiome()),
	gen.BiomeFlowerForest:         uint32(biome.FlowerForest{}.EncodeBiome()),
	gen.BiomeIceSpikes:            uint32(biome.IceSpikes{}.EncodeBiome()),
	gen.BiomeTallBirchForest:      uint32(biome.OldGrowthBirchForest{}.EncodeBiome()),
	gen.BiomeOldGrowthSpruceTaiga: uint32(biome.OldGrowthSpruceTaiga{}.EncodeBiome()),
	gen.BiomeWindsweptSavanna:     uint32(biome.WindsweptSavanna{}.EncodeBiome()),
	gen.BiomeErodedBadlands:       uint32(biome.ErodedBadlands{}.EncodeBiome()),
	gen.BiomeBambooJungle:         uint32(biome.BambooJungle{}.EncodeBiome()),
	gen.BiomeDripstoneCaves:       uint32(biome.DripstoneCaves{}.EncodeBiome()),
	gen.BiomeLushCaves:            uint32(biome.LushCaves{}.EncodeBiome()),
	gen.BiomeMeadow:               uint32(biome.Meadow{}.EncodeBiome()),
	gen.BiomeGrove:                uint32(biome.Grove{}.EncodeBiome()),
	gen.BiomeSnowySlopes:          uint32(biome.SnowySlopes{}.EncodeBiome()),
	gen.BiomeJaggedPeaks:          uint32(biome.JaggedPeaks{}.EncodeBiome()),
	gen.BiomeFrozenPeaks:          uint32(biome.FrozenPeaks{}.EncodeBiome()),
	gen.BiomeStonyPeaks:           uint32(biome.StonyPeaks{}.EncodeBiome()),
	gen.BiomeDeepDark:             uint32(biome.DeepDark{}.EncodeBiome()),
	gen.BiomeMangroveSwamp:        uint32(biome.MangroveSwamp{}.EncodeBiome()),
	gen.BiomeCherryGrove:          uint32(biome.CherryGrove{}.EncodeBiome()),
	gen.BiomePaleGarden:           uint32(biome.PaleGarden{}.EncodeBiome()),
	gen.BiomeSparseJungle:         uint32(biome.JungleEdge{}.EncodeBiome()),
	gen.BiomeNetherWastes:         uint32(biome.NetherWastes{}.EncodeBiome()),
	gen.BiomeSoulSandValley:       uint32(biome.SoulSandValley{}.EncodeBiome()),
	gen.BiomeCrimsonForest:        uint32(biome.CrimsonForest{}.EncodeBiome()),
	gen.BiomeWarpedForest:         uint32(biome.WarpedForest{}.EncodeBiome()),
	gen.BiomeBasaltDeltas:         uint32(biome.BasaltDeltas{}.EncodeBiome()),
	gen.BiomeTheEnd:               uint32(biome.End{}.EncodeBiome()),
	gen.BiomeSmallEndIslands:      uint32(biome.End{}.EncodeBiome()),
	gen.BiomeEndMidlands:          uint32(biome.End{}.EncodeBiome()),
	gen.BiomeEndHighlands:         uint32(biome.End{}.EncodeBiome()),
	gen.BiomeEndBarrens:           uint32(biome.End{}.EncodeBiome()),
}

var runtimeIDBiomes = func() map[uint32]gen.Biome {
	out := make(map[uint32]gen.Biome, len(biomeRuntimeIDs))
	for biome, rid := range biomeRuntimeIDs {
		if _, exists := out[rid]; !exists {
			out[rid] = biome
		}
	}
	return out
}()

var biomeKeysByID = func() [256]string {
	var out [256]string
	for id := 0; id < len(out); id++ {
		key := resolveBiomeKey(gen.Biome(id))
		if key != "" {
			out[id] = key
		}
	}
	return out
}()

var sortedBiomesByKey = func() []gen.Biome {
	biomes := make([]gen.Biome, 0, len(biomeRuntimeIDs))
	for id, key := range biomeKeysByID {
		if key != "" {
			biomes = append(biomes, gen.Biome(id))
		}
	}
	sort.Slice(biomes, func(i, j int) bool {
		return biomeKeysByID[biomes[i]] < biomeKeysByID[biomes[j]]
	})
	return biomes
}()

func biomeRuntimeID(b gen.Biome) uint32 {
	if rid, ok := biomeRuntimeIDs[b]; ok {
		return rid
	}
	return uint32(biome.Plains{}.EncodeBiome())
}

func biomeFromRuntimeID(rid uint32) gen.Biome {
	if biome, ok := runtimeIDBiomes[rid]; ok {
		return biome
	}
	return gen.BiomePlains
}

func biomeKey(b gen.Biome) string {
	if key := biomeKeysByID[b]; key != "" {
		return key
	}
	return "plains"
}

func resolveBiomeKey(b gen.Biome) string {
	switch b {
	case gen.BiomeNetherWastes:
		return "nether_wastes"
	case gen.BiomeSoulSandValley:
		return "soul_sand_valley"
	case gen.BiomeTheEnd:
		return "the_end"
	case gen.BiomeSmallEndIslands:
		return "small_end_islands"
	case gen.BiomeEndMidlands:
		return "end_midlands"
	case gen.BiomeEndHighlands:
		return "end_highlands"
	case gen.BiomeEndBarrens:
		return "end_barrens"
	case gen.BiomeSparseJungle:
		return "sparse_jungle"
	case gen.BiomeTallBirchForest:
		return "old_growth_birch_forest"
	case gen.BiomeGravellyMountains:
		return "windswept_gravelly_hills"
	case gen.BiomeSnowyMountains:
		return "snowy_mountains"
	}
	rid, ok := biomeRuntimeIDs[b]
	if !ok {
		return ""
	}
	if biome, ok := world.BiomeByID(int(rid)); ok {
		return biome.String()
	}
	return ""
}

func isFrozenSurfaceBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeSnowyPlains,
		gen.BiomeSnowyTaiga,
		gen.BiomeSnowyMountains,
		gen.BiomeSnowySlopes,
		gen.BiomeFrozenPeaks,
		gen.BiomeIceSpikes,
		gen.BiomeGrove:
		return true
	default:
		return false
	}
}

func isSandyBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeDesert, gen.BiomeBeach, gen.BiomeSnowyBeach:
		return true
	default:
		return false
	}
}

func isBadlandsBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeBadlands, gen.BiomeWoodedBadlands, gen.BiomeErodedBadlands:
		return true
	default:
		return false
	}
}

func isRockyBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeWindsweptHills,
		gen.BiomeWindsweptForest,
		gen.BiomeGravellyMountains,
		gen.BiomeStonyPeaks,
		gen.BiomeJaggedPeaks,
		gen.BiomeStonyShore:
		return true
	default:
		return false
	}
}

func isOceanBiome(b gen.Biome) bool {
	switch b {
	case gen.BiomeOcean,
		gen.BiomeDeepOcean,
		gen.BiomeWarmOcean,
		gen.BiomeLukewarmOcean,
		gen.BiomeColdOcean,
		gen.BiomeFrozenOcean,
		gen.BiomeDeepLukewarmOcean,
		gen.BiomeDeepColdOcean,
		gen.BiomeDeepFrozenOcean:
		return true
	default:
		return false
	}
}

func isRiverBiome(b gen.Biome) bool {
	return b == gen.BiomeRiver || b == gen.BiomeFrozenRiver
}
