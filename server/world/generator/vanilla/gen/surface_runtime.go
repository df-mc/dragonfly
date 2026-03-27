package gen

import "math"

const surfaceNoWaterHeight = -1 << 31

type SurfaceContext struct {
	BlockX int
	BlockY int
	BlockZ int

	SurfaceDepth     int
	SurfaceSecondary float64
	WaterHeight      int
	StoneDepthAbove  int
	StoneDepthBelow  int
	Steep            bool
	Biome            Biome
	MinSurfaceLevel  int
	MinY             int
	MaxY             int
}

type SurfaceRuntime struct {
	seed                  int64
	noises                *NoiseRegistry
	biomeSource           BiomeSource
	surfaceNoise          DoublePerlinNoise
	surfaceSecondaryNoise DoublePerlinNoise
	rule                  *surfaceRule
	useBandlands          bool
	bandlands             surfaceBandlands
}

type surfaceRuntimeLookup func(name string, properties map[string]string) uint32

type surfaceRuleKind uint8

const (
	surfaceRuleSequence surfaceRuleKind = iota
	surfaceRuleCondition
	surfaceRuleBlock
	surfaceRuleBandlands
)

type surfaceConditionKind uint8

const (
	surfaceConditionBiome surfaceConditionKind = iota
	surfaceConditionStoneDepth
	surfaceConditionYAbove
	surfaceConditionWater
	surfaceConditionNoiseThreshold
	surfaceConditionVerticalGradient
	surfaceConditionSteep
	surfaceConditionHole
	surfaceConditionAbovePreliminarySurface
	surfaceConditionTemperature
	surfaceConditionNot
)

type surfaceCaveSurface uint8

const (
	surfaceFloor surfaceCaveSurface = iota
	surfaceCeiling
)

type surfaceAnchorKind uint8

const (
	surfaceAnchorAbsolute surfaceAnchorKind = iota
	surfaceAnchorAboveBottom
	surfaceAnchorBelowTop
)

type surfaceVerticalAnchor struct {
	kind  surfaceAnchorKind
	value int
}

func (a surfaceVerticalAnchor) resolve(minY, maxY int) int {
	switch a.kind {
	case surfaceAnchorAboveBottom:
		return minY + a.value
	case surfaceAnchorBelowTop:
		return maxY - a.value
	default:
		return a.value
	}
}

type surfaceBlockState struct {
	name       string
	properties map[string]string
}

type surfaceCondition struct {
	kind surfaceConditionKind

	biomes []Biome

	offset              int
	addSurfaceDepth     bool
	secondaryDepthRange int
	caveSurface         surfaceCaveSurface

	anchor                 surfaceVerticalAnchor
	surfaceDepthMultiplier int
	addStoneDepth          bool

	noise        NoiseRef
	minThreshold float64
	maxThreshold float64

	trueAtAndBelow  surfaceVerticalAnchor
	falseAtAndAbove surfaceVerticalAnchor

	inner *surfaceCondition
}

type surfaceRule struct {
	kind surfaceRuleKind

	sequence []*surfaceRule
	ifTrue   *surfaceCondition
	thenRun  *surfaceRule
	block    surfaceBlockState
}

type surfaceBandlands struct {
	bands       [192]surfaceBlockState
	offsetNoise DoublePerlinNoise
}

func NewSurfaceRuntime(seed int64, noises *NoiseRegistry, biomeSource BiomeSource, rule *surfaceRule, useBandlands bool) *SurfaceRuntime {
	surfaceRNG := NewXoroshiro128FromSeed(seed + 0x1234567890ABCDEF)
	secondaryRNG := NewXoroshiro128FromSeed(seed + 0x0EDCBA0987654321)

	return &SurfaceRuntime{
		seed:                  seed,
		noises:                noises,
		biomeSource:           biomeSource,
		surfaceNoise:          NewDoublePerlinNoise(&surfaceRNG, []float64{1.0, 1.0, 1.0}, -6),
		surfaceSecondaryNoise: NewDoublePerlinNoise(&secondaryRNG, []float64{1.0, 1.0}, -6),
		rule:                  rule,
		useBandlands:          useBandlands,
		bandlands:             newSurfaceBandlands(seed),
	}
}

func NewOverworldSurfaceRuntime(seed int64, noises *NoiseRegistry, biomeSource BiomeSource) *SurfaceRuntime {
	return NewSurfaceRuntime(seed, noises, biomeSource, overworldSurfaceRule, true)
}

func NewNetherSurfaceRuntime(seed int64, noises *NoiseRegistry, biomeSource BiomeSource) *SurfaceRuntime {
	return NewSurfaceRuntime(seed, noises, biomeSource, netherSurfaceRule, false)
}

func NewEndSurfaceRuntime(seed int64, noises *NoiseRegistry, biomeSource BiomeSource) *SurfaceRuntime {
	return NewSurfaceRuntime(seed, noises, biomeSource, endSurfaceRule, false)
}

func (s *SurfaceRuntime) SurfaceDepth(x, z int) int {
	return int((s.surfaceNoise.Sample(float64(x), 0.0, float64(z))+1.0)*2.75 + 3.0)
}

func (s *SurfaceRuntime) SurfaceSecondary(x, z int) float64 {
	return (s.surfaceSecondaryNoise.Sample(float64(x), 0.0, float64(z)) + 1.0) * 0.5
}

func (s *SurfaceRuntime) TryApply(ctx SurfaceContext, lookup func(name string, properties map[string]string) uint32) (uint32, bool) {
	if lookup == nil {
		return 0, false
	}
	return s.evalRule(s.rule, ctx, lookup)
}

func (s *SurfaceRuntime) evalRule(rule *surfaceRule, ctx SurfaceContext, lookup surfaceRuntimeLookup) (uint32, bool) {
	if rule == nil {
		return 0, false
	}
	switch rule.kind {
	case surfaceRuleSequence:
		for _, child := range rule.sequence {
			if rid, ok := s.evalRule(child, ctx, lookup); ok {
				return rid, true
			}
		}
		return 0, false
	case surfaceRuleCondition:
		if s.evalCondition(rule.ifTrue, ctx) {
			return s.evalRule(rule.thenRun, ctx, lookup)
		}
		return 0, false
	case surfaceRuleBlock:
		return lookup(rule.block.name, rule.block.properties), true
	case surfaceRuleBandlands:
		if !s.useBandlands {
			return 0, false
		}
		state := s.bandlands.blockState(ctx)
		return lookup(state.name, state.properties), true
	default:
		return 0, false
	}
}

func (s *SurfaceRuntime) evalCondition(condition *surfaceCondition, ctx SurfaceContext) bool {
	if condition == nil {
		return false
	}
	switch condition.kind {
	case surfaceConditionBiome:
		for _, biome := range condition.biomes {
			if biome == ctx.Biome {
				return true
			}
		}
		return false
	case surfaceConditionStoneDepth:
		depth := ctx.StoneDepthAbove
		if condition.caveSurface == surfaceCeiling {
			depth = ctx.StoneDepthBelow
		}
		threshold := 1 + condition.offset
		if condition.addSurfaceDepth {
			threshold += ctx.SurfaceDepth
		}
		if condition.secondaryDepthRange > 0 {
			threshold += int(lerp(ctx.SurfaceSecondary, 0.0, float64(condition.secondaryDepthRange)))
		}
		return depth <= threshold
	case surfaceConditionYAbove:
		targetY := condition.anchor.resolve(ctx.MinY, ctx.MaxY)
		blockY := ctx.BlockY
		if condition.addStoneDepth {
			blockY += ctx.StoneDepthAbove
		}
		return blockY >= targetY+ctx.SurfaceDepth*condition.surfaceDepthMultiplier
	case surfaceConditionWater:
		if ctx.WaterHeight == surfaceNoWaterHeight {
			return true
		}
		blockY := ctx.BlockY
		if condition.addStoneDepth {
			blockY += ctx.StoneDepthAbove
		}
		return blockY >= ctx.WaterHeight+condition.offset+ctx.SurfaceDepth*condition.surfaceDepthMultiplier
	case surfaceConditionNoiseThreshold:
		if s.noises == nil {
			return false
		}
		value := s.noises.Sample(condition.noise, float64(ctx.BlockX), 0.0, float64(ctx.BlockZ))
		return value >= condition.minThreshold && value <= condition.maxThreshold
	case surfaceConditionVerticalGradient:
		trueY := condition.trueAtAndBelow.resolve(ctx.MinY, ctx.MaxY)
		falseY := condition.falseAtAndAbove.resolve(ctx.MinY, ctx.MaxY)
		if ctx.BlockY <= trueY {
			return true
		}
		if ctx.BlockY >= falseY {
			return false
		}
		probability := float64(falseY-ctx.BlockY) / float64(falseY-trueY)
		posSeed := s.seed + int64(ctx.BlockX)*341873128712 + int64(ctx.BlockY)*132897987541 + int64(ctx.BlockZ)*1664525
		rng := NewXoroshiro128FromSeed(posSeed)
		return rng.NextDouble() < probability
	case surfaceConditionSteep:
		return ctx.Steep
	case surfaceConditionHole:
		return ctx.SurfaceDepth <= 0
	case surfaceConditionAbovePreliminarySurface:
		return ctx.BlockY >= ctx.MinSurfaceLevel
	case surfaceConditionTemperature:
		return s.temperatureValue(ctx) <= 0.2
	case surfaceConditionNot:
		return !s.evalCondition(condition.inner, ctx)
	default:
		return false
	}
}

func (s *SurfaceRuntime) temperatureValue(ctx SurfaceContext) float64 {
	climate := s.biomeSource.SampleClimate(ctx.BlockX, ctx.BlockY, ctx.BlockZ)
	return float64(climate[temperatureIdx]) / 10000.0
}

func newSurfaceBandlands(seed int64) surfaceBandlands {
	rng := NewXoroshiro128FromSeed(seed + 0xBADBADBAD)
	bands := generateSurfaceBandlands(seed)
	return surfaceBandlands{
		bands:       bands,
		offsetNoise: NewDoublePerlinNoise(&rng, []float64{1.0}, -8),
	}
}

func generateSurfaceBandlands(seed int64) [192]surfaceBlockState {
	bands := [192]surfaceBlockState{}
	for i := range bands {
		bands[i] = surfaceBlockState{name: "minecraft:terracotta"}
	}

	rng := NewXoroshiro128FromSeed(seed)
	paintBandlands(&bands, &rng, "minecraft:orange_terracotta", 4, 3.0)
	paintBandlands(&bands, &rng, "minecraft:yellow_terracotta", 2, 2.0)
	paintBandlands(&bands, &rng, "minecraft:brown_terracotta", 2, 3.0)
	paintBandlands(&bands, &rng, "minecraft:red_terracotta", 2, 2.0)

	for i := 0; i < 2; i++ {
		start := int(rng.NextDouble() * 192.0)
		if start >= 0 && start < len(bands) {
			bands[start] = surfaceBlockState{name: "minecraft:white_terracotta"}
		}
	}

	paintBandlands(&bands, &rng, "minecraft:light_gray_terracotta", 2, 2.0)
	return bands
}

func paintBandlands(bands *[192]surfaceBlockState, rng *Xoroshiro128, name string, count int, maxWidth float64) {
	for i := 0; i < count; i++ {
		start := int(rng.NextDouble() * 192.0)
		width := int(rng.NextDouble()*maxWidth + 1.0)
		for offset := 0; offset < width; offset++ {
			idx := start + offset
			if idx >= 0 && idx < len(bands) {
				bands[idx] = surfaceBlockState{name: name}
			}
		}
	}
}

func (b surfaceBandlands) blockState(ctx SurfaceContext) surfaceBlockState {
	offset := int(math.Round(b.offsetNoise.Sample(float64(ctx.BlockX), 0.0, float64(ctx.BlockZ)) * 4.0))
	index := (ctx.BlockY + offset + len(b.bands)) % len(b.bands)
	if index < 0 {
		index += len(b.bands)
	}
	return b.bands[index]
}
