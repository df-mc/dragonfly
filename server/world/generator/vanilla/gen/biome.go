package gen

type Biome uint8

const (
	BiomeOcean Biome = iota
	BiomePlains
	BiomeDesert
	BiomeWindsweptHills
	BiomeForest
	BiomeTaiga
	BiomeSwamp
	BiomeRiver
	BiomeNetherWastes
	BiomeTheEnd
	BiomeFrozenOcean
	BiomeFrozenRiver
	BiomeSnowyPlains
	BiomeSnowyMountains
	BiomeMushroomFields
	BiomeMushroomFieldShore
	BiomeBeach
	BiomeDesertHills
	BiomeWoodedHills
	BiomeTaigaHills
	BiomeMountainEdge
	BiomeJungle
	BiomeJungleHills
	BiomeSparseJungle
	BiomeDeepOcean
	BiomeStonyShore
	BiomeSnowyBeach
	BiomeBirchForest
	BiomeBirchForestHills
	BiomeDarkForest
	BiomeSnowyTaiga
	BiomeSnowyTaigaHills
	BiomeOldGrowthPineTaiga
	BiomeOldGrowthPineTaigaHills
	BiomeWindsweptForest
	BiomeSavanna
	BiomeSavannaPlateau
	BiomeBadlands
	BiomeWoodedBadlands
	BiomeBadlandsPlateau
	BiomeSmallEndIslands
	BiomeEndMidlands
	BiomeEndHighlands
	BiomeEndBarrens
	BiomeWarmOcean
	BiomeLukewarmOcean
	BiomeColdOcean
	BiomeDeepWarmOcean
	BiomeDeepLukewarmOcean
	BiomeDeepColdOcean
	BiomeDeepFrozenOcean
)

const (
	BiomeTheVoid              Biome = 127
	BiomeSunflowerPlains      Biome = 129
	BiomeGravellyMountains    Biome = 131
	BiomeFlowerForest         Biome = 132
	BiomeIceSpikes            Biome = 140
	BiomeTallBirchForest      Biome = 155
	BiomeOldGrowthSpruceTaiga Biome = 160
	BiomeWindsweptSavanna     Biome = 163
	BiomeErodedBadlands       Biome = 165
	BiomeBambooJungle         Biome = 168
	BiomeSoulSandValley       Biome = 170
	BiomeCrimsonForest        Biome = 171
	BiomeWarpedForest         Biome = 172
	BiomeBasaltDeltas         Biome = 173
	BiomeDripstoneCaves       Biome = 174
	BiomeLushCaves            Biome = 175
	BiomeMeadow               Biome = 177
	BiomeGrove                Biome = 178
	BiomeSnowySlopes          Biome = 179
	BiomeJaggedPeaks          Biome = 180
	BiomeFrozenPeaks          Biome = 181
	BiomeStonyPeaks           Biome = 182
	BiomeDeepDark             Biome = 183
	BiomeMangroveSwamp        Biome = 184
	BiomeCherryGrove          Biome = 185
	BiomePaleGarden           Biome = 186
)

const (
	temperatureIdx = iota
	humidityIdx
	continentalnessIdx
	erosionIdx
	depthIdx
	weirdnessIdx
)

var temperatureBoundaries = [...]int64{-4500, -1500, 2000, 5500}
var humidityBoundaries = [...]int64{-3500, -1000, 1000, 3000}
var erosionBoundaries = [...]int64{-7800, -3750, -2225, 500, 4500, 5500}

const (
	mushroomCont   = -10500
	deepOceanCont  = -4550
	oceanCont      = -1900
	coastCont      = -1100
	nearInlandCont = 300
	midInlandCont  = 3000
)

var oceans = [2][5]Biome{
	{
		BiomeDeepFrozenOcean,
		BiomeDeepColdOcean,
		BiomeDeepOcean,
		BiomeDeepLukewarmOcean,
		BiomeWarmOcean,
	},
	{
		BiomeFrozenOcean,
		BiomeColdOcean,
		BiomeOcean,
		BiomeLukewarmOcean,
		BiomeWarmOcean,
	},
}

var middleBiomes = [5][5]Biome{
	{
		BiomeSnowyPlains,
		BiomeSnowyPlains,
		BiomeSnowyPlains,
		BiomeSnowyTaiga,
		BiomeTaiga,
	},
	{
		BiomePlains,
		BiomePlains,
		BiomeForest,
		BiomeTaiga,
		BiomeOldGrowthSpruceTaiga,
	},
	{
		BiomeFlowerForest,
		BiomePlains,
		BiomeForest,
		BiomeBirchForest,
		BiomeDarkForest,
	},
	{
		BiomeSavanna,
		BiomeSavanna,
		BiomeForest,
		BiomeJungle,
		BiomeJungle,
	},
	{
		BiomeDesert,
		BiomeDesert,
		BiomeDesert,
		BiomeDesert,
		BiomeDesert,
	},
}

var middleBiomeVariants = [5][5]Biome{
	{
		BiomeIceSpikes,
		BiomeTheVoid,
		BiomeSnowyTaiga,
		BiomeTheVoid,
		BiomeTheVoid,
	},
	{
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeOldGrowthPineTaiga,
	},
	{
		BiomeSunflowerPlains,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTallBirchForest,
		BiomeTheVoid,
	},
	{
		BiomeTheVoid,
		BiomeTheVoid,
		BiomePlains,
		BiomeSparseJungle,
		BiomeBambooJungle,
	},
	{
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
	},
}

var plateauBiomes = [5][5]Biome{
	{
		BiomeSnowyPlains,
		BiomeSnowyPlains,
		BiomeSnowyPlains,
		BiomeSnowyTaiga,
		BiomeSnowyTaiga,
	},
	{
		BiomeMeadow,
		BiomeMeadow,
		BiomeForest,
		BiomeTaiga,
		BiomeOldGrowthSpruceTaiga,
	},
	{
		BiomeMeadow,
		BiomeMeadow,
		BiomeMeadow,
		BiomeMeadow,
		BiomePaleGarden,
	},
	{
		BiomeSavannaPlateau,
		BiomeSavannaPlateau,
		BiomeForest,
		BiomeForest,
		BiomeJungle,
	},
	{
		BiomeBadlands,
		BiomeBadlands,
		BiomeBadlands,
		BiomeWoodedBadlands,
		BiomeWoodedBadlands,
	},
}

var plateauBiomeVariants = [5][5]Biome{
	{BiomeIceSpikes, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid},
	{
		BiomeCherryGrove,
		BiomeTheVoid,
		BiomeMeadow,
		BiomeMeadow,
		BiomeOldGrowthPineTaiga,
	},
	{
		BiomeCherryGrove,
		BiomeCherryGrove,
		BiomeForest,
		BiomeBirchForest,
		BiomeTheVoid,
	},
	{BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid},
	{
		BiomeErodedBadlands,
		BiomeErodedBadlands,
		BiomeTheVoid,
		BiomeTheVoid,
		BiomeTheVoid,
	},
}

var shatteredBiomes = [5][5]Biome{
	{
		BiomeGravellyMountains,
		BiomeGravellyMountains,
		BiomeWindsweptHills,
		BiomeWindsweptForest,
		BiomeWindsweptForest,
	},
	{
		BiomeGravellyMountains,
		BiomeGravellyMountains,
		BiomeWindsweptHills,
		BiomeWindsweptForest,
		BiomeWindsweptForest,
	},
	{
		BiomeWindsweptHills,
		BiomeWindsweptHills,
		BiomeWindsweptHills,
		BiomeWindsweptForest,
		BiomeWindsweptForest,
	},
	{BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid},
	{BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid, BiomeTheVoid},
}

type BiomeNoise struct {
	temperature     DoublePerlinNoise
	humidity        DoublePerlinNoise
	continentalness DoublePerlinNoise
	erosion         DoublePerlinNoise
	weirdness       DoublePerlinNoise
}

func NewBiomeNoise(seed int64) BiomeNoise {
	rng := NewXoroshiro128FromSeed(seed)
	return BiomeNoise{
		temperature:     newBiomeClimateNoise(&rng, NoiseTemperature),
		humidity:        newBiomeClimateNoise(&rng, NoiseVegetation),
		continentalness: newBiomeClimateNoise(&rng, NoiseContinentalness),
		erosion:         newBiomeClimateNoise(&rng, NoiseErosion),
		weirdness:       newBiomeClimateNoise(&rng, NoiseRidge),
	}
}

func newBiomeClimateNoise(rng *Xoroshiro128, ref NoiseRef) DoublePerlinNoise {
	params := NoiseParams[ref]
	return NewDoublePerlinNoise(rng, params.Amplitudes, params.FirstOctave)
}

func (b BiomeNoise) SampleClimate(x, y, z int) [6]int64 {
	const scale = 0.25
	qx := float64(x>>2) * scale
	qy := float64(y>>2) * scale
	qz := float64(z>>2) * scale

	return [6]int64{
		int64(b.temperature.Sample(qx, qy, qz) * 10000.0),
		int64(b.humidity.Sample(qx, qy, qz) * 10000.0),
		int64(b.continentalness.Sample(qx, qy, qz) * 10000.0),
		int64(b.erosion.Sample(qx, qy, qz) * 10000.0),
		depthFromY(y),
		int64(b.weirdness.Sample(qx, qy, qz) * 10000.0),
	}
}

func (b BiomeNoise) GetBiome(x, y, z int) Biome {
	climate := b.SampleClimate(x, y, z)
	return lookupOverworldPresetBiome(climate)
}

func tempIndex(temp int64) int {
	for i, boundary := range temperatureBoundaries {
		if temp < boundary {
			return i
		}
	}
	return 4
}

func humidIndex(humid int64) int {
	for i, boundary := range humidityBoundaries {
		if humid < boundary {
			return i
		}
	}
	return 4
}

func erosionIndex(erosion int64) int {
	for i, boundary := range erosionBoundaries {
		if erosion < boundary {
			return i
		}
	}
	return 6
}

func depthFromY(y int) int64 {
	depth := int64((float64(64-y) / 128.0) * 10000.0)
	if depth < -10000 {
		return -10000
	}
	if depth > 10000 {
		return 10000
	}
	return depth
}

func lookupBiome(climate [6]int64) Biome {
	temp := climate[temperatureIdx]
	humid := climate[humidityIdx]
	cont := climate[continentalnessIdx]
	erosion := climate[erosionIdx]
	depth := climate[depthIdx]
	weird := climate[weirdnessIdx]

	ti := tempIndex(temp)
	hi := humidIndex(humid)
	ei := erosionIndex(erosion)

	if cont < mushroomCont {
		return BiomeMushroomFields
	}
	if cont < deepOceanCont {
		return oceans[0][ti]
	}
	if cont < oceanCont {
		return oceans[1][ti]
	}

	if depth > 2000 {
		if cont > 8000 {
			return BiomeDripstoneCaves
		}
		if humid > 7000 {
			return BiomeLushCaves
		}
		if ei <= 1 && depth > 9000 {
			return BiomeDeepDark
		}
	}

	if cont < coastCont {
		if ei <= 2 {
			return BiomeStonyShore
		}
		if ei <= 4 {
			return pickBeach(ti)
		}
	}

	isValley := weird > -500 && weird < 500
	if isValley && cont >= coastCont && cont < midInlandCont && ei >= 2 && ei <= 5 {
		if ti == 0 {
			return BiomeFrozenRiver
		}
		return BiomeRiver
	}

	if ei == 6 && cont >= nearInlandCont {
		if ti == 1 || ti == 2 {
			return BiomeSwamp
		}
		if ti >= 3 {
			return BiomeMangroveSwamp
		}
	}

	useVariant := weird > 0
	switch ei {
	case 0:
		return pickPeakBiome(ti, hi, useVariant)
	case 1:
		if cont >= midInlandCont {
			return pickSlopeBiome(ti, hi, useVariant)
		}
		return pickMiddleOrBadlands(ti, hi, useVariant)
	case 2:
		if cont >= midInlandCont {
			return pickPlateauBiome(ti, hi, useVariant)
		}
		return pickMiddleBiome(ti, hi, useVariant)
	case 3:
		return pickMiddleOrBadlands(ti, hi, useVariant)
	case 4:
		return pickMiddleBiome(ti, hi, useVariant)
	case 5:
		if cont >= midInlandCont {
			return pickShatteredBiome(ti, hi, useVariant)
		}
		if ti > 1 && hi < 4 && useVariant {
			return BiomeWindsweptSavanna
		}
		return pickMiddleBiome(ti, hi, useVariant)
	default:
		return pickMiddleBiome(ti, hi, useVariant)
	}
}

func lookupOverworldPresetBiome(climate [6]int64) Biome {
	biome := lookupBiome(climate)
	cont := climate[continentalnessIdx]
	if cont >= oceanCont && cont < coastCont {
		return lookupPresetBiome(climate, overworldPresetPoints)
	}
	return biome
}

func pickBeach(ti int) Biome {
	switch ti {
	case 0:
		return BiomeSnowyBeach
	case 4:
		return BiomeDesert
	default:
		return BiomeBeach
	}
}

func pickMiddleBiome(ti, hi int, useVariant bool) Biome {
	if useVariant {
		if biome := middleBiomeVariants[ti][hi]; biome != BiomeTheVoid {
			return biome
		}
	}
	return middleBiomes[ti][hi]
}

func pickMiddleOrBadlands(ti, hi int, useVariant bool) Biome {
	if ti == 4 {
		return pickBadlands(hi, useVariant)
	}
	return pickMiddleBiome(ti, hi, useVariant)
}

func pickPlateauBiome(ti, hi int, useVariant bool) Biome {
	if useVariant {
		if biome := plateauBiomeVariants[ti][hi]; biome != BiomeTheVoid {
			return biome
		}
	}
	return plateauBiomes[ti][hi]
}

func pickSlopeBiome(ti, hi int, useVariant bool) Biome {
	if ti < 3 {
		if hi <= 1 {
			return BiomeSnowySlopes
		}
		return BiomeGrove
	}
	return pickPlateauBiome(ti, hi, useVariant)
}

func pickPeakBiome(ti, hi int, useVariant bool) Biome {
	if ti <= 2 {
		if useVariant {
			return BiomeFrozenPeaks
		}
		return BiomeJaggedPeaks
	}
	if ti == 3 {
		return BiomeStonyPeaks
	}
	return pickBadlands(hi, useVariant)
}

func pickBadlands(hi int, useVariant bool) Biome {
	if hi < 2 {
		if useVariant {
			return BiomeErodedBadlands
		}
		return BiomeBadlands
	}
	if hi < 3 {
		return BiomeBadlands
	}
	return BiomeWoodedBadlands
}

func pickShatteredBiome(ti, hi int, useVariant bool) Biome {
	if biome := shatteredBiomes[ti][hi]; biome != BiomeTheVoid {
		return biome
	}
	return pickMiddleBiome(ti, hi, useVariant)
}
