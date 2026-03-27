package gen

import "fmt"

type BiomeSource interface {
	SampleClimate(x, y, z int) [6]int64
	GetBiome(x, y, z int) Biome
}

type presetBiomeSource struct {
	preset string
	noise  BiomeNoise
}

type endBiomeSource struct {
	erosion EndIslandDensity
}

type climateParameter struct {
	min int64
	max int64
}

type climateParameterPoint struct {
	params [6]climateParameter
	offset int64
	biome  Biome
}

func NewBiomeSource(seed int64, registry *WorldgenRegistry, name string) (BiomeSource, error) {
	if registry == nil {
		registry = NewWorldgenRegistry()
	}
	if normalizeIdentifier(name) == "end" {
		return endBiomeSource{erosion: NewEndIslandDensity(seed)}, nil
	}

	def, err := registry.BiomeSourceParameterList(name)
	if err != nil {
		return nil, err
	}

	switch def.Preset {
	case "overworld":
		return presetBiomeSource{preset: def.Preset, noise: NewBiomeNoise(seed)}, nil
	case "nether":
		return presetBiomeSource{preset: def.Preset, noise: NewBiomeNoise(seed)}, nil
	default:
		return nil, fmt.Errorf("unsupported biome source preset %q", def.Preset)
	}
}

func (s presetBiomeSource) SampleClimate(x, y, z int) [6]int64 {
	return s.noise.SampleClimate(x, y, z)
}

func (s presetBiomeSource) GetBiome(x, y, z int) Biome {
	climate := s.SampleClimate(x, y, z)
	switch s.preset {
	case "overworld":
		return lookupOverworldPresetBiome(climate)
	case "nether":
		return lookupPresetBiome(climate, netherPresetPoints)
	default:
		return BiomePlains
	}
}

func (s endBiomeSource) SampleClimate(x, y, z int) [6]int64 {
	var climate [6]int64
	climate[erosionIdx] = int64(s.erosion.Sample(x, z) * 10000.0)
	return climate
}

func (s endBiomeSource) GetBiome(x, y, z int) Biome {
	chunkX := x >> 4
	chunkZ := z >> 4
	if int64(chunkX)*int64(chunkX)+int64(chunkZ)*int64(chunkZ) <= 4096 {
		return BiomeTheEnd
	}

	weirdBlockX := ((x>>4)*2 + 1) * 8
	weirdBlockZ := ((z>>4)*2 + 1) * 8
	heightValue := s.erosion.Sample(weirdBlockX, weirdBlockZ)
	switch {
	case heightValue > 0.25:
		return BiomeEndHighlands
	case heightValue >= -0.0625:
		return BiomeEndMidlands
	case heightValue < -0.21875:
		return BiomeSmallEndIslands
	default:
		return BiomeEndBarrens
	}
}

func climateSpan(min, max float64) climateParameter {
	return climateParameter{min: int64(min * 10000.0), max: int64(max * 10000.0)}
}

func climatePoint(value float64) climateParameter {
	return climateSpan(value, value)
}

func (p climateParameter) distance(value int64) int64 {
	if value < p.min {
		return p.min - value
	}
	if value > p.max {
		return value - p.max
	}
	return 0
}

func lookupPresetBiome(climate [6]int64, points []climateParameterPoint) Biome {
	if len(points) == 0 {
		return BiomePlains
	}
	best := points[0]
	bestFitness := climatePointFitness(climate, points[0])
	for _, point := range points[1:] {
		fitness := climatePointFitness(climate, point)
		if fitness < bestFitness {
			best = point
			bestFitness = fitness
		}
	}
	return best.biome
}

func climatePointFitness(climate [6]int64, point climateParameterPoint) int64 {
	var total int64
	for i, value := range climate {
		delta := point.params[i].distance(value)
		total += delta * delta
	}
	return total + point.offset*point.offset
}

var netherPresetPoints = []climateParameterPoint{
	{
		params: [6]climateParameter{
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
		},
		biome: BiomeNetherWastes,
	},
	{
		params: [6]climateParameter{
			climatePoint(0.0),
			climatePoint(-0.5),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
		},
		biome: BiomeSoulSandValley,
	},
	{
		params: [6]climateParameter{
			climatePoint(0.4),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
		},
		biome: BiomeCrimsonForest,
	},
	{
		params: [6]climateParameter{
			climatePoint(0.0),
			climatePoint(0.5),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
		},
		offset: int64(0.375 * 10000.0),
		biome:  BiomeWarpedForest,
	},
	{
		params: [6]climateParameter{
			climatePoint(-0.5),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
			climatePoint(0.0),
		},
		offset: int64(0.175 * 10000.0),
		biome:  BiomeBasaltDeltas,
	},
}
