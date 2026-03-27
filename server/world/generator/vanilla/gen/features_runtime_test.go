package gen

import "testing"

func TestFeatureRegistryBiomePlacedFeatures(t *testing.T) {
	registry := NewFeatureRegistry()

	features := registry.BiomePlacedFeatures("plains", GenerationStepVegetalDecoration)
	if len(features) == 0 {
		t.Fatal("expected plains vegetal decoration features")
	}
	if !containsString(features, "trees_plains") {
		t.Fatalf("expected plains vegetation to include trees_plains, got %v", features)
	}
	if !containsString(features, "patch_tall_grass_2") {
		t.Fatalf("expected plains vegetation to include patch_tall_grass_2, got %v", features)
	}
}

func TestPlacedFeatureTreesPlainsDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	placed, err := registry.Placed("trees_plains")
	if err != nil {
		t.Fatalf("failed to load trees_plains: %v", err)
	}
	if len(placed.Placement) != 6 {
		t.Fatalf("expected 6 placement modifiers, got %d", len(placed.Placement))
	}

	count, err := placed.Placement[0].Count()
	if err != nil {
		t.Fatalf("failed to decode count placement: %v", err)
	}
	if count.Count.Kind != "weighted_list" || len(count.Count.Distribution) != 2 {
		t.Fatalf("unexpected count provider: %#v", count.Count)
	}

	depth, err := placed.Placement[2].SurfaceWaterDepthFilter()
	if err != nil {
		t.Fatalf("failed to decode surface water depth filter: %v", err)
	}
	if depth.MaxWaterDepth != 0 {
		t.Fatalf("expected max water depth 0, got %d", depth.MaxWaterDepth)
	}

	predicateFilter, err := placed.Placement[4].BlockPredicateFilter()
	if err != nil {
		t.Fatalf("failed to decode block predicate filter: %v", err)
	}
	if predicateFilter.Predicate.Type != "would_survive" {
		t.Fatalf("expected would_survive predicate, got %s", predicateFilter.Predicate.Type)
	}
	wouldSurvive, err := predicateFilter.Predicate.WouldSurvive()
	if err != nil {
		t.Fatalf("failed to decode would_survive predicate: %v", err)
	}
	if wouldSurvive.State.Name != "oak_sapling" {
		t.Fatalf("expected oak_sapling survival check, got %s", wouldSurvive.State.Name)
	}

	configured, err := registry.ResolveConfigured(placed.Feature)
	if err != nil {
		t.Fatalf("failed to resolve configured feature: %v", err)
	}
	if configured.Type != "random_selector" {
		t.Fatalf("expected random_selector configured feature, got %s", configured.Type)
	}

	selector, err := configured.RandomSelector()
	if err != nil {
		t.Fatalf("failed to decode random selector config: %v", err)
	}
	defaultPlaced, err := registry.ResolvePlaced(selector.Default)
	if err != nil {
		t.Fatalf("failed to resolve default placed feature: %v", err)
	}
	if defaultPlaced.Feature.Name != "oak_bees_005" {
		t.Fatalf("expected default tree oak_bees_005, got %q", defaultPlaced.Feature.Name)
	}
	if len(selector.Features) != 2 {
		t.Fatalf("expected 2 random selector branches, got %d", len(selector.Features))
	}
}

func TestPatchTallGrassRuntimeDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	placed, err := registry.Placed("patch_tall_grass_2")
	if err != nil {
		t.Fatalf("failed to load patch_tall_grass_2: %v", err)
	}
	if len(placed.Placement) != 5 {
		t.Fatalf("expected 5 placement modifiers, got %d", len(placed.Placement))
	}

	noiseCount, err := placed.Placement[0].NoiseThresholdCount()
	if err != nil {
		t.Fatalf("failed to decode noise threshold count: %v", err)
	}
	if noiseCount.AboveNoise != 7 || noiseCount.BelowNoise != 0 {
		t.Fatalf("unexpected noise threshold count payload: %#v", noiseCount)
	}

	rarity, err := placed.Placement[1].RarityFilter()
	if err != nil {
		t.Fatalf("failed to decode rarity filter: %v", err)
	}
	if rarity.Chance != 32 {
		t.Fatalf("expected rarity chance 32, got %d", rarity.Chance)
	}

	configured, err := registry.ResolveConfigured(placed.Feature)
	if err != nil {
		t.Fatalf("failed to resolve configured feature: %v", err)
	}
	if configured.Type != "random_patch" {
		t.Fatalf("expected random_patch configured feature, got %s", configured.Type)
	}

	patch, err := configured.RandomPatch()
	if err != nil {
		t.Fatalf("failed to decode random patch: %v", err)
	}
	if patch.Tries != 96 || patch.XZSpread != 7 || patch.YSpread != 3 {
		t.Fatalf("unexpected random patch payload: %#v", patch)
	}

	inlinePlaced, err := registry.ResolvePlaced(patch.Feature)
	if err != nil {
		t.Fatalf("failed to resolve inline placed feature: %v", err)
	}
	if len(inlinePlaced.Placement) != 1 {
		t.Fatalf("expected one inline placement filter, got %d", len(inlinePlaced.Placement))
	}
	blockFilter, err := inlinePlaced.Placement[0].BlockPredicateFilter()
	if err != nil {
		t.Fatalf("failed to decode inline block predicate filter: %v", err)
	}
	matching, err := blockFilter.Predicate.MatchingBlocks()
	if err != nil {
		t.Fatalf("failed to decode matching_blocks predicate: %v", err)
	}
	if len(matching.Blocks.Values) != 1 || matching.Blocks.Values[0] != "air" {
		t.Fatalf("expected air-only predicate, got %#v", matching.Blocks.Values)
	}

	inlineConfigured, err := registry.ResolveConfigured(inlinePlaced.Feature)
	if err != nil {
		t.Fatalf("failed to resolve inline configured feature: %v", err)
	}
	if inlineConfigured.Type != "simple_block" {
		t.Fatalf("expected simple_block inline configured feature, got %s", inlineConfigured.Type)
	}
	simpleBlock, err := inlineConfigured.SimpleBlock()
	if err != nil {
		t.Fatalf("failed to decode simple_block: %v", err)
	}
	stateProvider, err := simpleBlock.ToPlace.SimpleState()
	if err != nil {
		t.Fatalf("failed to decode simple state provider: %v", err)
	}
	if stateProvider.State.Name != "tall_grass" {
		t.Fatalf("expected tall_grass placement, got %s", stateProvider.State.Name)
	}
}

func TestOreCoalUpperRuntimeDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	placed, err := registry.Placed("ore_coal_upper")
	if err != nil {
		t.Fatalf("failed to load ore_coal_upper: %v", err)
	}
	if len(placed.Placement) != 4 {
		t.Fatalf("expected 4 placement modifiers, got %d", len(placed.Placement))
	}

	heightRange, err := placed.Placement[2].HeightRange()
	if err != nil {
		t.Fatalf("failed to decode height range: %v", err)
	}
	if heightRange.Height.Kind != "uniform" {
		t.Fatalf("expected uniform height range, got %s", heightRange.Height.Kind)
	}
	if heightRange.Height.MinInclusive.Kind != "absolute" || heightRange.Height.MinInclusive.Value != 136 {
		t.Fatalf("unexpected min height anchor: %#v", heightRange.Height.MinInclusive)
	}
	if heightRange.Height.MaxInclusive.Kind != "below_top" || heightRange.Height.MaxInclusive.Value != 0 {
		t.Fatalf("unexpected max height anchor: %#v", heightRange.Height.MaxInclusive)
	}

	configured, err := registry.ResolveConfigured(placed.Feature)
	if err != nil {
		t.Fatalf("failed to resolve configured ore feature: %v", err)
	}
	if configured.Type != "ore" {
		t.Fatalf("expected ore configured feature, got %s", configured.Type)
	}
	ore, err := configured.Ore()
	if err != nil {
		t.Fatalf("failed to decode ore config: %v", err)
	}
	if ore.Size != 17 || len(ore.Targets) != 2 {
		t.Fatalf("unexpected ore config: %#v", ore)
	}
	if ore.Targets[0].State.Name != "coal_ore" || ore.Targets[1].State.Name != "deepslate_coal_ore" {
		t.Fatalf("unexpected ore targets: %#v", ore.Targets)
	}
}

func TestPatchSugarCaneRuntimeDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	placed, err := registry.Placed("patch_sugar_cane")
	if err != nil {
		t.Fatalf("failed to load patch_sugar_cane: %v", err)
	}
	if len(placed.Placement) != 4 {
		t.Fatalf("expected 4 placement modifiers, got %d", len(placed.Placement))
	}

	configured, err := registry.ResolveConfigured(placed.Feature)
	if err != nil {
		t.Fatalf("failed to resolve configured feature: %v", err)
	}
	patch, err := configured.RandomPatch()
	if err != nil {
		t.Fatalf("failed to decode sugar cane patch: %v", err)
	}
	inlinePlaced, err := registry.ResolvePlaced(patch.Feature)
	if err != nil {
		t.Fatalf("failed to resolve inline placed feature: %v", err)
	}
	inlineConfigured, err := registry.ResolveConfigured(inlinePlaced.Feature)
	if err != nil {
		t.Fatalf("failed to resolve inline configured feature: %v", err)
	}
	column, err := inlineConfigured.BlockColumn()
	if err != nil {
		t.Fatalf("failed to decode block column config: %v", err)
	}
	if column.Direction != "up" || len(column.Layers) != 1 {
		t.Fatalf("unexpected block column payload: %#v", column)
	}
	if column.Layers[0].Height.Kind != "biased_to_bottom" {
		t.Fatalf("expected biased_to_bottom height, got %s", column.Layers[0].Height.Kind)
	}

	filter, err := inlinePlaced.Placement[0].BlockPredicateFilter()
	if err != nil {
		t.Fatalf("failed to decode sugar cane predicate filter: %v", err)
	}
	if filter.Predicate.Type != "all_of" {
		t.Fatalf("expected all_of sugar cane predicate, got %s", filter.Predicate.Type)
	}
}

func TestDiskSandRuntimeDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	configured, err := registry.Configured("disk_sand")
	if err != nil {
		t.Fatalf("failed to load disk_sand: %v", err)
	}
	disk, err := configured.Disk()
	if err != nil {
		t.Fatalf("failed to decode disk_sand: %v", err)
	}
	if disk.HalfHeight != 2 || disk.Radius.Kind != "uniform" {
		t.Fatalf("unexpected disk payload: %#v", disk)
	}
	if disk.StateProvider.Type != "rule_based_state_provider" {
		t.Fatalf("expected rule_based_state_provider, got %s", disk.StateProvider.Type)
	}
}

func TestSpringWaterRuntimeDecodes(t *testing.T) {
	registry := NewFeatureRegistry()

	configured, err := registry.Configured("spring_water")
	if err != nil {
		t.Fatalf("failed to load spring_water: %v", err)
	}
	spring, err := configured.SpringFeature()
	if err != nil {
		t.Fatalf("failed to decode spring feature: %v", err)
	}
	if spring.State.Name != "water" || spring.RockCount != 4 || spring.HoleCount != 1 {
		t.Fatalf("unexpected spring config: %#v", spring)
	}
	if len(spring.ValidBlocks.Values) == 0 {
		t.Fatal("expected spring valid blocks")
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
