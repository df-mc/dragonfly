package gen

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type GenerationStep uint8

const (
	GenerationStepRawGeneration GenerationStep = iota
	GenerationStepLakes
	GenerationStepLocalModifications
	GenerationStepUndergroundStructures
	GenerationStepSurfaceStructures
	GenerationStepStrongholds
	GenerationStepUndergroundOres
	GenerationStepUndergroundDecoration
	GenerationStepFluidSprings
	GenerationStepVegetalDecoration
	GenerationStepTopLayerModification
)

type FeatureRegistry struct {
	mu              sync.Mutex
	placedCache     map[string]placedCacheEntry
	configuredCache map[string]configuredCacheEntry
}

type placedCacheEntry struct {
	loaded bool
	def    PlacedFeatureDef
	err    error
}

type configuredCacheEntry struct {
	loaded bool
	def    ConfiguredFeatureDef
	err    error
}

func NewFeatureRegistry() *FeatureRegistry {
	return &FeatureRegistry{
		placedCache:     make(map[string]placedCacheEntry),
		configuredCache: make(map[string]configuredCacheEntry),
	}
}

func (r *FeatureRegistry) BiomePlacedFeatures(biomeName string, step GenerationStep) []string {
	steps, ok := biomePlacedFeaturesByName[biomeName]
	if !ok {
		return nil
	}
	features, ok := steps[step]
	if !ok {
		return nil
	}
	return append([]string(nil), features...)
}

func (r *FeatureRegistry) Placed(name string) (PlacedFeatureDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.placedCache[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := placedFeatureJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return PlacedFeatureDef{}, fmt.Errorf("unknown placed feature %q", name)
	}

	var def PlacedFeatureDef
	err := json.Unmarshal([]byte(raw), &def)
	r.placedCache[key] = placedCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

func (r *FeatureRegistry) Configured(name string) (ConfiguredFeatureDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.configuredCache[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := configuredFeatureJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return ConfiguredFeatureDef{}, fmt.Errorf("unknown configured feature %q", name)
	}

	var def ConfiguredFeatureDef
	err := json.Unmarshal([]byte(raw), &def)
	r.configuredCache[key] = configuredCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

func (r *FeatureRegistry) ResolvePlaced(ref PlacedFeatureRef) (PlacedFeatureDef, error) {
	if ref.Inline != nil {
		return *ref.Inline, nil
	}
	return r.Placed(ref.Name)
}

func (r *FeatureRegistry) ResolveConfigured(ref ConfiguredFeatureRef) (ConfiguredFeatureDef, error) {
	if ref.Inline != nil {
		return *ref.Inline, nil
	}
	return r.Configured(ref.Name)
}

type PlacedFeatureDef struct {
	Feature   ConfiguredFeatureRef `json:"feature"`
	Placement []PlacementModifier  `json:"placement"`
}

type ConfiguredFeatureRef struct {
	Name   string
	Inline *ConfiguredFeatureDef
}

func (r *ConfiguredFeatureRef) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		r.Name = normalizeIdentifier(name)
		r.Inline = nil
		return nil
	}

	var inline ConfiguredFeatureDef
	if err := json.Unmarshal(data, &inline); err != nil {
		return err
	}
	r.Name = ""
	r.Inline = &inline
	return nil
}

type PlacedFeatureRef struct {
	Name   string
	Inline *PlacedFeatureDef
}

func (r *PlacedFeatureRef) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		r.Name = normalizeIdentifier(name)
		r.Inline = nil
		return nil
	}

	var inline PlacedFeatureDef
	if err := json.Unmarshal(data, &inline); err != nil {
		return err
	}
	r.Name = ""
	r.Inline = &inline
	return nil
}

type ConfiguredFeatureDef struct {
	Type   string
	Config json.RawMessage
}

func (f *ConfiguredFeatureDef) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type   string          `json:"type"`
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Type = normalizeIdentifier(raw.Type)
	f.Config = append(json.RawMessage(nil), raw.Config...)
	return nil
}

func (f ConfiguredFeatureDef) RandomSelector() (RandomSelectorConfig, error) {
	return decodeFeatureConfig[RandomSelectorConfig](f, "random_selector")
}

func (f ConfiguredFeatureDef) SimpleRandomSelector() (SimpleRandomSelectorConfig, error) {
	return decodeFeatureConfig[SimpleRandomSelectorConfig](f, "simple_random_selector")
}

func (f ConfiguredFeatureDef) RandomBooleanSelector() (RandomBooleanSelectorConfig, error) {
	return decodeFeatureConfig[RandomBooleanSelectorConfig](f, "random_boolean_selector")
}

func (f ConfiguredFeatureDef) RandomPatch() (RandomPatchConfig, error) {
	return decodeFeatureConfig[RandomPatchConfig](f, "random_patch")
}

func (f ConfiguredFeatureDef) Flower() (RandomPatchConfig, error) {
	return decodeFeatureConfig[RandomPatchConfig](f, "flower")
}

func (f ConfiguredFeatureDef) SimpleBlock() (SimpleBlockConfig, error) {
	return decodeFeatureConfig[SimpleBlockConfig](f, "simple_block")
}

func (f ConfiguredFeatureDef) BlockColumn() (BlockColumnConfig, error) {
	return decodeFeatureConfig[BlockColumnConfig](f, "block_column")
}

func (f ConfiguredFeatureDef) Ore() (OreConfig, error) {
	return decodeFeatureConfig[OreConfig](f, "ore")
}

func (f ConfiguredFeatureDef) ScatteredOre() (OreConfig, error) {
	return decodeFeatureConfig[OreConfig](f, "scattered_ore")
}

func (f ConfiguredFeatureDef) Disk() (DiskConfig, error) {
	return decodeFeatureConfig[DiskConfig](f, "disk")
}

func (f ConfiguredFeatureDef) SpringFeature() (SpringFeatureConfig, error) {
	return decodeFeatureConfig[SpringFeatureConfig](f, "spring_feature")
}

func (f ConfiguredFeatureDef) UnderwaterMagma() (UnderwaterMagmaConfig, error) {
	return decodeFeatureConfig[UnderwaterMagmaConfig](f, "underwater_magma")
}

func (f ConfiguredFeatureDef) Seagrass() (SeagrassConfig, error) {
	return decodeFeatureConfig[SeagrassConfig](f, "seagrass")
}

func (f ConfiguredFeatureDef) Kelp() (KelpConfig, error) {
	return decodeFeatureConfig[KelpConfig](f, "kelp")
}

func (f ConfiguredFeatureDef) MultifaceGrowth() (MultifaceGrowthConfig, error) {
	return decodeFeatureConfig[MultifaceGrowthConfig](f, "multiface_growth")
}

func (f ConfiguredFeatureDef) SculkPatch() (SculkPatchConfig, error) {
	return decodeFeatureConfig[SculkPatchConfig](f, "sculk_patch")
}

func (f ConfiguredFeatureDef) PointedDripstone() (PointedDripstoneConfig, error) {
	return decodeFeatureConfig[PointedDripstoneConfig](f, "pointed_dripstone")
}

func (f ConfiguredFeatureDef) DripstoneCluster() (DripstoneClusterConfig, error) {
	return decodeFeatureConfig[DripstoneClusterConfig](f, "dripstone_cluster")
}

func (f ConfiguredFeatureDef) Vines() (VinesConfig, error) {
	return decodeFeatureConfig[VinesConfig](f, "vines")
}

func (f ConfiguredFeatureDef) SeaPickle() (SeaPickleConfig, error) {
	return decodeFeatureConfig[SeaPickleConfig](f, "sea_pickle")
}

func (f ConfiguredFeatureDef) Lake() (LakeConfig, error) {
	return decodeFeatureConfig[LakeConfig](f, "lake")
}

func (f ConfiguredFeatureDef) FreezeTopLayer() (FreezeTopLayerConfig, error) {
	return decodeFeatureConfig[FreezeTopLayerConfig](f, "freeze_top_layer")
}

func (f ConfiguredFeatureDef) FallenTree() (FallenTreeConfig, error) {
	return decodeFeatureConfig[FallenTreeConfig](f, "fallen_tree")
}

func (f ConfiguredFeatureDef) Tree() (TreeConfig, error) {
	return decodeFeatureConfig[TreeConfig](f, "tree")
}

func (f ConfiguredFeatureDef) Bamboo() (BambooConfig, error) {
	return decodeFeatureConfig[BambooConfig](f, "bamboo")
}

func (f ConfiguredFeatureDef) VegetationPatch() (VegetationPatchConfig, error) {
	return decodeFeatureConfig[VegetationPatchConfig](f, "vegetation_patch")
}

func (f ConfiguredFeatureDef) WaterloggedVegetationPatch() (VegetationPatchConfig, error) {
	return decodeFeatureConfig[VegetationPatchConfig](f, "waterlogged_vegetation_patch")
}

func (f ConfiguredFeatureDef) RootSystem() (RootSystemConfig, error) {
	return decodeFeatureConfig[RootSystemConfig](f, "root_system")
}

func (f ConfiguredFeatureDef) HugeFungus() (HugeFungusConfig, error) {
	return decodeFeatureConfig[HugeFungusConfig](f, "huge_fungus")
}

func (f ConfiguredFeatureDef) NetherForestVegetation() (NetherForestVegetationConfig, error) {
	return decodeFeatureConfig[NetherForestVegetationConfig](f, "nether_forest_vegetation")
}

func (f ConfiguredFeatureDef) TwistingVines() (TwistingVinesConfig, error) {
	return decodeFeatureConfig[TwistingVinesConfig](f, "twisting_vines")
}

func (f ConfiguredFeatureDef) WeepingVines() (WeepingVinesConfig, error) {
	return decodeFeatureConfig[WeepingVinesConfig](f, "weeping_vines")
}

func (f ConfiguredFeatureDef) NetherrackReplaceBlobs() (NetherrackReplaceBlobsConfig, error) {
	return decodeFeatureConfig[NetherrackReplaceBlobsConfig](f, "netherrack_replace_blobs")
}

func (f ConfiguredFeatureDef) GlowstoneBlob() (GlowstoneBlobConfig, error) {
	return decodeFeatureConfig[GlowstoneBlobConfig](f, "glowstone_blob")
}

func (f ConfiguredFeatureDef) BasaltPillar() (BasaltPillarConfig, error) {
	return decodeFeatureConfig[BasaltPillarConfig](f, "basalt_pillar")
}

func (f ConfiguredFeatureDef) BasaltColumns() (BasaltColumnsConfig, error) {
	return decodeFeatureConfig[BasaltColumnsConfig](f, "basalt_columns")
}

func (f ConfiguredFeatureDef) DeltaFeature() (DeltaFeatureConfig, error) {
	return decodeFeatureConfig[DeltaFeatureConfig](f, "delta_feature")
}

func (f ConfiguredFeatureDef) ChorusPlant() (ChorusPlantConfig, error) {
	return decodeFeatureConfig[ChorusPlantConfig](f, "chorus_plant")
}

func (f ConfiguredFeatureDef) EndIsland() (EndIslandConfig, error) {
	return decodeFeatureConfig[EndIslandConfig](f, "end_island")
}

func (f ConfiguredFeatureDef) EndSpike() (EndSpikeConfig, error) {
	return decodeFeatureConfig[EndSpikeConfig](f, "end_spike")
}

func (f ConfiguredFeatureDef) EndPlatform() (EndPlatformConfig, error) {
	return decodeFeatureConfig[EndPlatformConfig](f, "end_platform")
}

func (f ConfiguredFeatureDef) EndGateway() (EndGatewayConfig, error) {
	return decodeFeatureConfig[EndGatewayConfig](f, "end_gateway")
}

type PlacementModifier struct {
	Type string
	Data json.RawMessage
}

func (m *PlacementModifier) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	m.Type = normalizeIdentifier(probe.Type)
	m.Data = append(json.RawMessage(nil), data...)
	return nil
}

func (m PlacementModifier) Count() (CountPlacement, error) {
	return decodePlacement[CountPlacement](m, "count")
}

func (m PlacementModifier) CountOnEveryLayer() (CountPlacement, error) {
	return decodePlacement[CountPlacement](m, "count_on_every_layer")
}

func (m PlacementModifier) Heightmap() (HeightmapPlacement, error) {
	return decodePlacement[HeightmapPlacement](m, "heightmap")
}

func (m PlacementModifier) HeightRange() (HeightRangePlacement, error) {
	return decodePlacement[HeightRangePlacement](m, "height_range")
}

func (m PlacementModifier) SurfaceWaterDepthFilter() (SurfaceWaterDepthFilterPlacement, error) {
	return decodePlacement[SurfaceWaterDepthFilterPlacement](m, "surface_water_depth_filter")
}

func (m PlacementModifier) BlockPredicateFilter() (BlockPredicateFilterPlacement, error) {
	return decodePlacement[BlockPredicateFilterPlacement](m, "block_predicate_filter")
}

func (m PlacementModifier) RarityFilter() (RarityFilterPlacement, error) {
	return decodePlacement[RarityFilterPlacement](m, "rarity_filter")
}

func (m PlacementModifier) NoiseThresholdCount() (NoiseThresholdCountPlacement, error) {
	return decodePlacement[NoiseThresholdCountPlacement](m, "noise_threshold_count")
}

func (m PlacementModifier) NoiseBasedCount() (NoiseBasedCountPlacement, error) {
	return decodePlacement[NoiseBasedCountPlacement](m, "noise_based_count")
}

func (m PlacementModifier) RandomOffset() (RandomOffsetPlacement, error) {
	return decodePlacement[RandomOffsetPlacement](m, "random_offset")
}

func (m PlacementModifier) FixedPlacement() (FixedPlacementPlacement, error) {
	return decodePlacement[FixedPlacementPlacement](m, "fixed_placement")
}

func (m PlacementModifier) EnvironmentScan() (EnvironmentScanPlacement, error) {
	return decodePlacement[EnvironmentScanPlacement](m, "environment_scan")
}

func (m PlacementModifier) SurfaceRelativeThresholdFilter() (SurfaceRelativeThresholdFilterPlacement, error) {
	return decodePlacement[SurfaceRelativeThresholdFilterPlacement](m, "surface_relative_threshold_filter")
}

type CountPlacement struct {
	Count IntProvider `json:"count"`
}

type HeightmapPlacement struct {
	Heightmap string `json:"heightmap"`
}

type HeightRangePlacement struct {
	Height HeightProvider `json:"height"`
}

type SurfaceWaterDepthFilterPlacement struct {
	MaxWaterDepth int `json:"max_water_depth"`
}

type BlockPredicateFilterPlacement struct {
	Predicate BlockPredicate `json:"predicate"`
}

type RarityFilterPlacement struct {
	Chance int `json:"chance"`
}

type NoiseThresholdCountPlacement struct {
	AboveNoise int     `json:"above_noise"`
	BelowNoise int     `json:"below_noise"`
	NoiseLevel float64 `json:"noise_level"`
}

type NoiseBasedCountPlacement struct {
	NoiseFactor       float64 `json:"noise_factor"`
	NoiseOffset       float64 `json:"noise_offset"`
	NoiseToCountRatio int     `json:"noise_to_count_ratio"`
}

type RandomOffsetPlacement struct {
	XZSpread IntProvider `json:"xz_spread"`
	YSpread  IntProvider `json:"y_spread"`
}

type FixedPlacementPlacement struct {
	Positions []BlockPos `json:"positions"`
}

type EnvironmentScanPlacement struct {
	AllowedSearchCondition *BlockPredicate `json:"allowed_search_condition"`
	DirectionOfSearch      string          `json:"direction_of_search"`
	MaxSteps               int             `json:"max_steps"`
	TargetCondition        BlockPredicate  `json:"target_condition"`
}

type SurfaceRelativeThresholdFilterPlacement struct {
	Heightmap    string `json:"heightmap"`
	MinInclusive *int   `json:"min_inclusive"`
	MaxInclusive *int   `json:"max_inclusive"`
}

type RandomSelectorConfig struct {
	Default  PlacedFeatureRef         `json:"default"`
	Features []RandomSelectorEntryDef `json:"features"`
}

type SimpleRandomSelectorConfig struct {
	Features []PlacedFeatureRef `json:"features"`
}

type RandomBooleanSelectorConfig struct {
	FeatureFalse PlacedFeatureRef `json:"feature_false"`
	FeatureTrue  PlacedFeatureRef `json:"feature_true"`
}

type RandomSelectorEntryDef struct {
	Chance  float64          `json:"chance"`
	Feature PlacedFeatureRef `json:"feature"`
}

type RandomPatchConfig struct {
	Feature  PlacedFeatureRef `json:"feature"`
	Tries    int              `json:"tries"`
	XZSpread int              `json:"xz_spread"`
	YSpread  int              `json:"y_spread"`
}

type SimpleBlockConfig struct {
	ToPlace StateProvider `json:"to_place"`
}

type BlockColumnConfig struct {
	AllowedPlacement BlockPredicate     `json:"allowed_placement"`
	Direction        string             `json:"direction"`
	Layers           []BlockColumnLayer `json:"layers"`
	PrioritizeTip    bool               `json:"prioritize_tip"`
}

type BlockColumnLayer struct {
	Height   IntProvider   `json:"height"`
	Provider StateProvider `json:"provider"`
}

type OreConfig struct {
	DiscardChanceOnAirExposure float64           `json:"discard_chance_on_air_exposure"`
	Size                       int               `json:"size"`
	Targets                    []OreTargetConfig `json:"targets"`
}

type OreTargetConfig struct {
	State  BlockState         `json:"state"`
	Target OreTargetPredicate `json:"target"`
}

type DiskConfig struct {
	HalfHeight    int            `json:"half_height"`
	Radius        IntProvider    `json:"radius"`
	StateProvider StateProvider  `json:"state_provider"`
	Target        BlockPredicate `json:"target"`
}

type SpringFeatureConfig struct {
	HoleCount          int             `json:"hole_count"`
	RequiresBlockBelow bool            `json:"requires_block_below"`
	RockCount          int             `json:"rock_count"`
	State              BlockState      `json:"state"`
	ValidBlocks        StringOrStrings `json:"valid_blocks"`
}

type UnderwaterMagmaConfig struct {
	FloorSearchRange                     int     `json:"floor_search_range"`
	PlacementProbabilityPerValidPosition float64 `json:"placement_probability_per_valid_position"`
	PlacementRadiusAroundFloor           int     `json:"placement_radius_around_floor"`
}

type SeagrassConfig struct {
	Probability float64 `json:"probability"`
}

type KelpConfig struct{}

type MultifaceGrowthConfig struct {
	Block             string   `json:"block"`
	CanBePlacedOn     []string `json:"can_be_placed_on"`
	CanPlaceOnCeiling bool     `json:"can_place_on_ceiling"`
	CanPlaceOnFloor   bool     `json:"can_place_on_floor"`
	CanPlaceOnWall    bool     `json:"can_place_on_wall"`
	ChanceOfSpreading float64  `json:"chance_of_spreading"`
	SearchRange       int      `json:"search_range"`
}

func (c *MultifaceGrowthConfig) UnmarshalJSON(data []byte) error {
	type multifaceGrowthConfig MultifaceGrowthConfig
	var raw multifaceGrowthConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw.Block = normalizeIdentifier(raw.Block)
	for i, value := range raw.CanBePlacedOn {
		raw.CanBePlacedOn[i] = normalizeIdentifier(value)
	}
	*c = MultifaceGrowthConfig(raw)
	return nil
}

type SculkPatchConfig struct {
	AmountPerCharge  int     `json:"amount_per_charge"`
	CatalystChance   float64 `json:"catalyst_chance"`
	ChargeCount      int     `json:"charge_count"`
	ExtraRareGrowths int     `json:"extra_rare_growths"`
	GrowthRounds     int     `json:"growth_rounds"`
	SpreadAttempts   int     `json:"spread_attempts"`
	SpreadRounds     int     `json:"spread_rounds"`
}

type PointedDripstoneConfig struct {
	ChanceOfDirectionalSpread float64 `json:"chance_of_directional_spread"`
	ChanceOfSpreadRadius2     float64 `json:"chance_of_spread_radius2"`
	ChanceOfSpreadRadius3     float64 `json:"chance_of_spread_radius3"`
	ChanceOfTallerDripstone   float64 `json:"chance_of_taller_dripstone"`
}

type DripstoneClusterConfig struct {
	FloorToCeilingSearchRange    int         `json:"floor_to_ceiling_search_range"`
	Height                       IntProvider `json:"height"`
	HeightDeviation              int         `json:"height_deviation"`
	Radius                       IntProvider `json:"radius"`
	DripstoneBlockLayerThickness IntProvider `json:"dripstone_block_layer_thickness"`
}

type VinesConfig struct{}

type SeaPickleConfig struct {
	Count int `json:"count"`
}

type LakeConfig struct {
	Barrier StateProvider `json:"barrier"`
	Fluid   StateProvider `json:"fluid"`
}

type FreezeTopLayerConfig struct{}

type OreTargetPredicate struct {
	PredicateType string `json:"predicate_type"`
	Tag           string `json:"tag"`
	Block         string `json:"block"`
}

func (p *OreTargetPredicate) UnmarshalJSON(data []byte) error {
	type oreTargetPredicate OreTargetPredicate
	var raw oreTargetPredicate
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw.PredicateType = normalizeIdentifier(raw.PredicateType)
	raw.Tag = normalizeIdentifier(raw.Tag)
	raw.Block = normalizeIdentifier(raw.Block)
	*p = OreTargetPredicate(raw)
	return nil
}

type TreeConfig struct {
	Decorators      []FeatureDecorator `json:"decorators"`
	DirtProvider    StateProvider      `json:"dirt_provider"`
	FoliagePlacer   TypedJSONValue     `json:"foliage_placer"`
	FoliageProvider StateProvider      `json:"foliage_provider"`
	ForceDirt       bool               `json:"force_dirt"`
	IgnoreVines     bool               `json:"ignore_vines"`
	MinimumSize     TypedJSONValue     `json:"minimum_size"`
	TrunkPlacer     TypedJSONValue     `json:"trunk_placer"`
	TrunkProvider   StateProvider      `json:"trunk_provider"`
}

type FallenTreeConfig struct {
	LogDecorators   []FeatureDecorator `json:"log_decorators"`
	LogLength       IntProvider        `json:"log_length"`
	StumpDecorators []FeatureDecorator `json:"stump_decorators"`
	TrunkProvider   StateProvider      `json:"trunk_provider"`
}

type BambooConfig struct {
	Probability float64 `json:"probability"`
}

type VegetationPatchConfig struct {
	Depth                  IntProvider      `json:"depth"`
	ExtraBottomBlockChance float64          `json:"extra_bottom_block_chance"`
	ExtraEdgeColumnChance  float64          `json:"extra_edge_column_chance"`
	GroundState            StateProvider    `json:"ground_state"`
	Replaceable            string           `json:"replaceable"`
	Surface                string           `json:"surface"`
	VegetationChance       float64          `json:"vegetation_chance"`
	VegetationFeature      PlacedFeatureRef `json:"vegetation_feature"`
	VerticalRange          int              `json:"vertical_range"`
	XZRadius               IntProvider      `json:"xz_radius"`
}

type RootSystemConfig struct {
	AllowedTreePosition          BlockPredicate   `json:"allowed_tree_position"`
	AllowedVerticalWaterForTree  int              `json:"allowed_vertical_water_for_tree"`
	Feature                      PlacedFeatureRef `json:"feature"`
	HangingRootPlacementAttempts int              `json:"hanging_root_placement_attempts"`
	HangingRootRadius            int              `json:"hanging_root_radius"`
	HangingRootStateProvider     StateProvider    `json:"hanging_root_state_provider"`
	HangingRootsVerticalSpan     int              `json:"hanging_roots_vertical_span"`
	RequiredVerticalSpaceForTree int              `json:"required_vertical_space_for_tree"`
	RootColumnMaxHeight          int              `json:"root_column_max_height"`
	RootPlacementAttempts        int              `json:"root_placement_attempts"`
	RootRadius                   int              `json:"root_radius"`
	RootReplaceable              string           `json:"root_replaceable"`
	RootStateProvider            StateProvider    `json:"root_state_provider"`
}

type HugeFungusConfig struct {
	DecorState        BlockState     `json:"decor_state"`
	HatState          BlockState     `json:"hat_state"`
	Planted           bool           `json:"planted"`
	ReplaceableBlocks BlockPredicate `json:"replaceable_blocks"`
	StemState         BlockState     `json:"stem_state"`
	ValidBaseBlock    BlockState     `json:"valid_base_block"`
}

type NetherForestVegetationConfig struct {
	SpreadHeight  int           `json:"spread_height"`
	SpreadWidth   int           `json:"spread_width"`
	StateProvider StateProvider `json:"state_provider"`
}

type TwistingVinesConfig struct {
	MaxHeight    int `json:"max_height"`
	SpreadHeight int `json:"spread_height"`
	SpreadWidth  int `json:"spread_width"`
}

type WeepingVinesConfig struct{}

type NetherrackReplaceBlobsConfig struct {
	Radius IntProvider `json:"radius"`
	State  BlockState  `json:"state"`
	Target BlockState  `json:"target"`
}

type GlowstoneBlobConfig struct{}

type BasaltPillarConfig struct{}

type BasaltColumnsConfig struct {
	Height IntProvider `json:"height"`
	Reach  IntProvider `json:"reach"`
}

type DeltaFeatureConfig struct {
	Contents BlockState  `json:"contents"`
	Rim      BlockState  `json:"rim"`
	RimSize  IntProvider `json:"rim_size"`
	Size     IntProvider `json:"size"`
}

type ChorusPlantConfig struct{}

type EndIslandConfig struct{}

type EndSpikeConfig struct {
	CrystalInvulnerable bool              `json:"crystal_invulnerable"`
	Spikes              []json.RawMessage `json:"spikes"`
}

type EndPlatformConfig struct{}

type EndGatewayConfig struct {
	Exact bool      `json:"exact"`
	Exit  *BlockPos `json:"exit"`
}

type FeatureDecorator struct {
	Type string
	Data json.RawMessage
}

func (d *FeatureDecorator) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	d.Type = normalizeIdentifier(probe.Type)
	d.Data = append(json.RawMessage(nil), data...)
	return nil
}

type TypedJSONValue struct {
	Type string
	Data json.RawMessage
}

func (v *TypedJSONValue) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	v.Type = normalizeIdentifier(probe.Type)
	v.Data = append(json.RawMessage(nil), data...)
	return nil
}

type IntProvider struct {
	Kind         string
	Constant     *int
	MinInclusive int
	MaxInclusive int
	Mean         float64
	Deviation    float64
	Distribution []WeightedInt
	Source       *IntProvider
	Raw          json.RawMessage
}

type WeightedInt struct {
	Data   int `json:"data"`
	Weight int `json:"weight"`
}

func (p *IntProvider) UnmarshalJSON(data []byte) error {
	p.Raw = append(json.RawMessage(nil), data...)

	var constant int
	if err := json.Unmarshal(data, &constant); err == nil {
		p.Kind = "constant"
		p.Constant = &constant
		p.MinInclusive = constant
		p.MaxInclusive = constant
		p.Distribution = nil
		return nil
	}

	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	p.Kind = normalizeIdentifier(probe.Type)
	p.Constant = nil
	p.Distribution = nil

	switch p.Kind {
	case "uniform":
		var raw struct {
			MinInclusive int `json:"min_inclusive"`
			MaxInclusive int `json:"max_inclusive"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
	case "biased_to_bottom":
		var raw struct {
			MinInclusive int `json:"min_inclusive"`
			MaxInclusive int `json:"max_inclusive"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
	case "weighted_list":
		var raw struct {
			Distribution []WeightedInt `json:"distribution"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.Distribution = raw.Distribution
	case "clamped":
		var raw struct {
			MinInclusive int         `json:"min_inclusive"`
			MaxInclusive int         `json:"max_inclusive"`
			Source       IntProvider `json:"source"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
		p.Source = &raw.Source
	case "clamped_normal":
		var raw struct {
			MinInclusive int     `json:"min_inclusive"`
			MaxInclusive int     `json:"max_inclusive"`
			Mean         float64 `json:"mean"`
			Deviation    float64 `json:"deviation"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
		p.Mean = raw.Mean
		p.Deviation = raw.Deviation
	}
	return nil
}

type HeightProvider struct {
	Kind         string
	MinInclusive VerticalAnchor
	MaxInclusive VerticalAnchor
	Mean         float64
	Deviation    float64
	Raw          json.RawMessage
}

func (p *HeightProvider) UnmarshalJSON(data []byte) error {
	p.Raw = append(json.RawMessage(nil), data...)

	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	p.Kind = normalizeIdentifier(probe.Type)
	switch p.Kind {
	case "uniform":
		var raw struct {
			MinInclusive VerticalAnchor `json:"min_inclusive"`
			MaxInclusive VerticalAnchor `json:"max_inclusive"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
	case "trapezoid", "biased_to_bottom", "very_biased_to_bottom":
		var raw struct {
			MinInclusive VerticalAnchor `json:"min_inclusive"`
			MaxInclusive VerticalAnchor `json:"max_inclusive"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
	case "clamped_normal":
		var raw struct {
			MinInclusive VerticalAnchor `json:"min_inclusive"`
			MaxInclusive VerticalAnchor `json:"max_inclusive"`
			Mean         float64        `json:"mean"`
			Deviation    float64        `json:"deviation"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.MinInclusive = raw.MinInclusive
		p.MaxInclusive = raw.MaxInclusive
		p.Mean = raw.Mean
		p.Deviation = raw.Deviation
	}
	return nil
}

type VerticalAnchor struct {
	Kind  string
	Value int
}

func (a *VerticalAnchor) UnmarshalJSON(data []byte) error {
	var raw map[string]int
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 1 {
		return fmt.Errorf("expected one vertical anchor field, got %d", len(raw))
	}
	for kind, value := range raw {
		a.Kind = normalizeIdentifier(kind)
		a.Value = value
	}
	return nil
}

type BlockState struct {
	Name       string            `json:"Name"`
	Properties map[string]string `json:"Properties,omitempty"`
}

func (s *BlockState) UnmarshalJSON(data []byte) error {
	type blockState BlockState
	var raw blockState
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw.Name = normalizeIdentifier(raw.Name)
	*s = BlockState(raw)
	return nil
}

type StateProvider struct {
	Type string
	Data json.RawMessage
}

func (p *StateProvider) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe.Type == "" {
		var ruleProbe struct {
			Fallback json.RawMessage   `json:"fallback"`
			Rules    []json.RawMessage `json:"rules"`
		}
		if err := json.Unmarshal(data, &ruleProbe); err != nil {
			return err
		}
		if len(ruleProbe.Fallback) != 0 || len(ruleProbe.Rules) != 0 {
			p.Type = "rule_based_state_provider"
		}
	} else {
		p.Type = normalizeIdentifier(probe.Type)
	}
	p.Data = append(json.RawMessage(nil), data...)
	return nil
}

func (p StateProvider) SimpleState() (SimpleStateProviderConfig, error) {
	return decodeTypedJSON[SimpleStateProviderConfig](p.Type, "simple_state_provider", p.Data)
}

func (p StateProvider) WeightedState() (WeightedStateProviderConfig, error) {
	return decodeTypedJSON[WeightedStateProviderConfig](p.Type, "weighted_state_provider", p.Data)
}

func (p StateProvider) RandomizedIntState() (RandomizedIntStateProviderConfig, error) {
	return decodeTypedJSON[RandomizedIntStateProviderConfig](p.Type, "randomized_int_state_provider", p.Data)
}

func (p StateProvider) RuleBasedState() (RuleBasedStateProviderConfig, error) {
	return decodeTypedJSON[RuleBasedStateProviderConfig](p.Type, "rule_based_state_provider", p.Data)
}

func (p StateProvider) NoiseThreshold() (NoiseThresholdStateProviderConfig, error) {
	return decodeTypedJSON[NoiseThresholdStateProviderConfig](p.Type, "noise_threshold_provider", p.Data)
}

type SimpleStateProviderConfig struct {
	State BlockState `json:"state"`
}

type WeightedStateProviderConfig struct {
	Entries []WeightedStateProviderEntry `json:"entries"`
}

type WeightedStateProviderEntry struct {
	Data   BlockState `json:"data"`
	Weight int        `json:"weight"`
}

type RandomizedIntStateProviderConfig struct {
	Property string        `json:"property"`
	Source   StateProvider `json:"source"`
	Values   IntProvider   `json:"values"`
}

type RuleBasedStateProviderConfig struct {
	Fallback StateProvider           `json:"fallback"`
	Rules    []RuleBasedStateRuleDef `json:"rules"`
}

type RuleBasedStateRuleDef struct {
	IfTrue BlockPredicate `json:"if_true"`
	Then   StateProvider  `json:"then"`
}

type NoiseThresholdStateProviderConfig struct {
	DefaultState BlockState      `json:"default_state"`
	HighChance   float64         `json:"high_chance"`
	HighStates   []BlockState    `json:"high_states"`
	LowStates    []BlockState    `json:"low_states"`
	Noise        NoiseParamsData `json:"noise"`
	Scale        float64         `json:"scale"`
	Seed         int64           `json:"seed"`
	Threshold    float64         `json:"threshold"`
}

type BlockPredicate struct {
	Type string
	Data json.RawMessage
}

func (p *BlockPredicate) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	p.Type = normalizeIdentifier(probe.Type)
	p.Data = append(json.RawMessage(nil), data...)
	return nil
}

func (p BlockPredicate) MatchingBlocks() (MatchingBlocksPredicateConfig, error) {
	return decodeTypedJSON[MatchingBlocksPredicateConfig](p.Type, "matching_blocks", p.Data)
}

func (p BlockPredicate) MatchingFluids() (MatchingFluidsPredicateConfig, error) {
	return decodeTypedJSON[MatchingFluidsPredicateConfig](p.Type, "matching_fluids", p.Data)
}

func (p BlockPredicate) Not() (NotPredicateConfig, error) {
	return decodeTypedJSON[NotPredicateConfig](p.Type, "not", p.Data)
}

func (p BlockPredicate) WouldSurvive() (WouldSurvivePredicateConfig, error) {
	return decodeTypedJSON[WouldSurvivePredicateConfig](p.Type, "would_survive", p.Data)
}

func (p BlockPredicate) MatchingBlockTag() (MatchingBlockTagPredicateConfig, error) {
	return decodeTypedJSON[MatchingBlockTagPredicateConfig](p.Type, "matching_block_tag", p.Data)
}

type MatchingBlocksPredicateConfig struct {
	Blocks StringOrStrings `json:"blocks"`
	Offset BlockPos        `json:"offset"`
}

type MatchingFluidsPredicateConfig struct {
	Fluids StringOrStrings `json:"fluids"`
	Offset BlockPos        `json:"offset"`
}

type NotPredicateConfig struct {
	Predicate BlockPredicate `json:"predicate"`
}

type WouldSurvivePredicateConfig struct {
	State BlockState `json:"state"`
}

type MatchingBlockTagPredicateConfig struct {
	Tag    string   `json:"tag"`
	Offset BlockPos `json:"offset"`
}

type StringOrStrings struct {
	Values []string
}

func (s *StringOrStrings) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		s.Values = []string{normalizeIdentifier(single)}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	s.Values = s.Values[:0]
	for _, value := range many {
		s.Values = append(s.Values, normalizeIdentifier(value))
	}
	return nil
}

type BlockPos [3]int

func (p *BlockPos) UnmarshalJSON(data []byte) error {
	var raw [3]int
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = BlockPos(raw)
	return nil
}

func decodeFeatureConfig[T any](feature ConfiguredFeatureDef, expectedType string) (T, error) {
	return decodeTypedJSON[T](feature.Type, expectedType, feature.Config)
}

func decodePlacement[T any](modifier PlacementModifier, expectedType string) (T, error) {
	return decodeTypedJSON[T](modifier.Type, expectedType, modifier.Data)
}

func decodeTypedJSON[T any](actualType, expectedType string, data []byte) (T, error) {
	var out T
	if actualType != expectedType {
		return out, fmt.Errorf("expected %s, got %s", expectedType, actualType)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	return out, nil
}

func normalizeIdentifier(value string) string {
	return strings.TrimPrefix(value, "minecraft:")
}
