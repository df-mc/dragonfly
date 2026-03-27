package gen

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type RandomSpreadPlacement struct {
	Spacing                  int     `json:"spacing"`
	Separation               int     `json:"separation"`
	SpreadType               string  `json:"spread_type"`
	Salt                     int     `json:"salt"`
	Frequency                float64 `json:"frequency"`
	FrequencyReductionMethod string  `json:"frequency_reduction_method"`
	LocateOffset             [3]int  `json:"locate_offset"`
}

func (d StructurePlacementDef) RandomSpread() (RandomSpreadPlacement, error) {
	var out RandomSpreadPlacement
	if d.Type != "random_spread" {
		return out, fmt.Errorf("expected random_spread, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.SpreadType = normalizeIdentifier(out.SpreadType)
	if out.SpreadType == "" {
		out.SpreadType = "linear"
	}
	if out.Frequency == 0 {
		out.Frequency = 1
	}
	out.FrequencyReductionMethod = normalizeIdentifier(out.FrequencyReductionMethod)
	return out, nil
}

type ConcentricRingsPlacement struct {
	Distance        int    `json:"distance"`
	Spread          int    `json:"spread"`
	Count           int    `json:"count"`
	PreferredBiomes string `json:"preferred_biomes"`
	Salt            int    `json:"salt"`
}

func (d StructurePlacementDef) ConcentricRings() (ConcentricRingsPlacement, error) {
	var out ConcentricRingsPlacement
	if d.Type != "concentric_rings" {
		return out, fmt.Errorf("expected concentric_rings, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

type JigsawStructureDef struct {
	Biomes                string                  `json:"biomes"`
	MaxDistanceFromCenter int                     `json:"max_distance_from_center"`
	ProjectStartToHeight  string                  `json:"project_start_to_heightmap"`
	Size                  int                     `json:"size"`
	StartHeight           StructureHeightProvider `json:"start_height"`
	StartJigsawName       string                  `json:"start_jigsaw_name"`
	StartPool             string                  `json:"start_pool"`
	Step                  string                  `json:"step"`
	TerrainAdaptation     string                  `json:"terrain_adaptation"`
	UseExpansionHack      bool                    `json:"use_expansion_hack"`
	PoolAliases           []PoolAliasDef          `json:"pool_aliases"`
	LiquidSettings        string                  `json:"liquid_settings"`
	DimensionPadding      int                     `json:"dimension_padding"`
	SpawnOverrides        map[string]any          `json:"spawn_overrides"`
}

func (d StructureDef) Jigsaw() (JigsawStructureDef, error) {
	var out JigsawStructureDef
	if d.Type != "jigsaw" {
		return out, fmt.Errorf("expected jigsaw, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.StartJigsawName = normalizeIdentifier(out.StartJigsawName)
	out.StartPool = normalizeIdentifier(out.StartPool)
	out.ProjectStartToHeight = strings.ToUpper(out.ProjectStartToHeight)
	out.Step = normalizeIdentifier(out.Step)
	out.TerrainAdaptation = normalizeIdentifier(out.TerrainAdaptation)
	out.LiquidSettings = normalizeIdentifier(out.LiquidSettings)
	return out, nil
}

type GenericStructureDef struct {
	Biomes            string         `json:"biomes"`
	MineshaftType     string         `json:"mineshaft_type"`
	SpawnOverrides    map[string]any `json:"spawn_overrides"`
	Step              string         `json:"step"`
	TerrainAdaptation string         `json:"terrain_adaptation"`
}

func (d StructureDef) Generic() (GenericStructureDef, error) {
	var out GenericStructureDef
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.Biomes = normalizeIdentifier(out.Biomes)
	out.MineshaftType = normalizeIdentifier(out.MineshaftType)
	out.Step = normalizeIdentifier(out.Step)
	out.TerrainAdaptation = normalizeIdentifier(out.TerrainAdaptation)
	return out, nil
}

type NetherFossilStructureDef struct {
	Biomes            string                  `json:"biomes"`
	Height            StructureHeightProvider `json:"height"`
	SpawnOverrides    map[string]any          `json:"spawn_overrides"`
	Step              string                  `json:"step"`
	TerrainAdaptation string                  `json:"terrain_adaptation"`
}

func (d StructureDef) NetherFossil() (NetherFossilStructureDef, error) {
	var out NetherFossilStructureDef
	if d.Type != "nether_fossil" {
		return out, fmt.Errorf("expected nether_fossil, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.Biomes = normalizeIdentifier(out.Biomes)
	out.Step = normalizeIdentifier(out.Step)
	out.TerrainAdaptation = normalizeIdentifier(out.TerrainAdaptation)
	return out, nil
}

type ShipwreckStructureDef struct {
	Biomes         string         `json:"biomes"`
	IsBeached      bool           `json:"is_beached"`
	SpawnOverrides map[string]any `json:"spawn_overrides"`
	Step           string         `json:"step"`
}

func (d StructureDef) Shipwreck() (ShipwreckStructureDef, error) {
	var out ShipwreckStructureDef
	if d.Type != "shipwreck" {
		return out, fmt.Errorf("expected shipwreck, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.Biomes = normalizeIdentifier(out.Biomes)
	out.Step = normalizeIdentifier(out.Step)
	return out, nil
}

type OceanRuinStructureDef struct {
	BiomeTemp          string         `json:"biome_temp"`
	Biomes             string         `json:"biomes"`
	ClusterProbability float64        `json:"cluster_probability"`
	LargeProbability   float64        `json:"large_probability"`
	SpawnOverrides     map[string]any `json:"spawn_overrides"`
	Step               string         `json:"step"`
}

func (d StructureDef) OceanRuin() (OceanRuinStructureDef, error) {
	var out OceanRuinStructureDef
	if d.Type != "ocean_ruin" {
		return out, fmt.Errorf("expected ocean_ruin, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.BiomeTemp = normalizeIdentifier(out.BiomeTemp)
	out.Biomes = normalizeIdentifier(out.Biomes)
	out.Step = normalizeIdentifier(out.Step)
	return out, nil
}

type RuinedPortalSetupDef struct {
	AirPocketProbability  float64 `json:"air_pocket_probability"`
	CanBeCold             bool    `json:"can_be_cold"`
	Mossiness             float64 `json:"mossiness"`
	Overgrown             bool    `json:"overgrown"`
	Placement             string  `json:"placement"`
	ReplaceWithBlackstone bool    `json:"replace_with_blackstone"`
	Vines                 bool    `json:"vines"`
	Weight                float64 `json:"weight"`
}

type RuinedPortalStructureDef struct {
	Biomes         string                 `json:"biomes"`
	Setups         []RuinedPortalSetupDef `json:"setups"`
	SpawnOverrides map[string]any         `json:"spawn_overrides"`
	Step           string                 `json:"step"`
}

func (d StructureDef) RuinedPortal() (RuinedPortalStructureDef, error) {
	var out RuinedPortalStructureDef
	if d.Type != "ruined_portal" {
		return out, fmt.Errorf("expected ruined_portal, got %s", d.Type)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.Biomes = normalizeIdentifier(out.Biomes)
	out.Step = normalizeIdentifier(out.Step)
	for i := range out.Setups {
		out.Setups[i].Placement = normalizeIdentifier(out.Setups[i].Placement)
	}
	return out, nil
}

type StructureHeightProvider struct {
	Kind         string
	Anchor       VerticalAnchor
	MinInclusive VerticalAnchor
	MaxInclusive VerticalAnchor
	Mean         float64
	Deviation    float64
}

func (p *StructureHeightProvider) UnmarshalJSON(data []byte) error {
	var anchorMap map[string]int
	if err := json.Unmarshal(data, &anchorMap); err == nil && len(anchorMap) == 1 {
		for kind, value := range anchorMap {
			p.Kind = "constant"
			p.Anchor = VerticalAnchor{Kind: normalizeIdentifier(kind), Value: value}
			return nil
		}
	}

	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	p.Kind = normalizeIdentifier(probe.Type)
	switch p.Kind {
	case "uniform", "trapezoid", "biased_to_bottom", "very_biased_to_bottom":
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
	default:
		return fmt.Errorf("unsupported structure height provider type %q", probe.Type)
	}
	return nil
}

type PoolAliasDef struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

func (d *PoolAliasDef) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	d.Type = normalizeIdentifier(raw.Type)
	d.Raw = append(json.RawMessage(nil), data...)
	return nil
}

type ProcessorListRef struct {
	Name   string
	Inline *ProcessorListDef
}

func (r *ProcessorListRef) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		r.Name = normalizeIdentifier(name)
		r.Inline = nil
		return nil
	}

	var inline ProcessorListDef
	if err := json.Unmarshal(data, &inline); err != nil {
		return err
	}
	r.Name = ""
	r.Inline = &inline
	return nil
}

type StructureProcessorDef struct {
	Type string
	Raw  json.RawMessage
}

func (d *StructureProcessorDef) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type string `json:"processor_type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	d.Type = normalizeIdentifier(raw.Type)
	d.Raw = append(json.RawMessage(nil), data...)
	return nil
}

type SinglePoolElementDef struct {
	Location   string           `json:"location"`
	Processors ProcessorListRef `json:"processors"`
	Projection string           `json:"projection"`
}

func (d TemplatePoolElementDef) Single() (SinglePoolElementDef, error) {
	var out SinglePoolElementDef
	switch d.ElementType {
	case "legacy_single_pool_element", "single_pool_element":
	default:
		return out, fmt.Errorf("expected single pool element, got %s", d.ElementType)
	}
	if err := json.Unmarshal(d.Raw, &out); err != nil {
		return out, err
	}
	out.Location = normalizeIdentifier(out.Location)
	out.Projection = normalizeIdentifier(out.Projection)
	return out, nil
}

type StructureTemplate struct {
	Size    [3]int
	Palette []StructureTemplateBlockState
	Blocks  []StructureTemplateBlock
}

type StructureTemplateBlockState struct {
	Name       string
	Properties map[string]any
}

type StructureTemplateBlock struct {
	Pos   [3]int
	State int
	NBT   map[string]any
}

type structureTemplateNBT struct {
	Author      string                             `nbt:"author"`
	DataVersion int32                              `nbt:"DataVersion"`
	Size        []int32                            `nbt:"size"`
	Palette     []structureTemplateBlockStateNBT   `nbt:"palette"`
	Palettes    [][]structureTemplateBlockStateNBT `nbt:"palettes"`
	Blocks      []structureTemplateBlockNBT        `nbt:"blocks"`
	Entities    []map[string]any                   `nbt:"entities"`
}

type structureTemplateBlockStateNBT struct {
	Name       string         `nbt:"Name"`
	Properties map[string]any `nbt:"Properties"`
}

type structureTemplateBlockNBT struct {
	Pos   []int32        `nbt:"pos"`
	State int32          `nbt:"state"`
	NBT   map[string]any `nbt:"nbt"`
}

type StructureTemplateRegistry struct {
	worldgen *WorldgenRegistry

	mu    sync.Mutex
	cache map[string]structureTemplateEntry
}

type structureTemplateEntry struct {
	loaded bool
	def    StructureTemplate
	err    error
}

func NewStructureTemplateRegistry(worldgen *WorldgenRegistry) *StructureTemplateRegistry {
	if worldgen == nil {
		worldgen = NewWorldgenRegistry()
	}
	return &StructureTemplateRegistry{
		worldgen: worldgen,
		cache:    make(map[string]structureTemplateEntry),
	}
}

func (r *StructureTemplateRegistry) Template(name string) (StructureTemplate, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.cache[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}

	data, err := r.worldgen.StructureTemplate(key)
	if err != nil {
		r.cache[key] = structureTemplateEntry{loaded: true, err: err}
		r.mu.Unlock()
		return StructureTemplate{}, err
	}

	def, parseErr := decodeStructureTemplate(data)
	r.cache[key] = structureTemplateEntry{loaded: true, def: def, err: parseErr}
	r.mu.Unlock()
	return def, parseErr
}

func decodeStructureTemplate(data []byte) (StructureTemplate, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return StructureTemplate{}, err
	}
	defer reader.Close()

	var raw structureTemplateNBT
	if err := nbt.NewDecoderWithEncoding(reader, nbt.BigEndian).Decode(&raw); err != nil {
		if strings.Contains(err.Error(), "maximum nesting depth") {
			return decodeStructureTemplateFallback(data)
		}
		return StructureTemplate{}, err
	}

	var palette []structureTemplateBlockStateNBT
	switch {
	case len(raw.Palette) != 0:
		palette = raw.Palette
	case len(raw.Palettes) != 0:
		palette = raw.Palettes[0]
	default:
		palette = nil
	}

	var size [3]int
	if len(raw.Size) >= 3 {
		size = [3]int{int(raw.Size[0]), int(raw.Size[1]), int(raw.Size[2])}
	}

	out := StructureTemplate{
		Size:    size,
		Palette: make([]StructureTemplateBlockState, 0, len(palette)),
		Blocks:  make([]StructureTemplateBlock, 0, len(raw.Blocks)),
	}
	for _, state := range palette {
		out.Palette = append(out.Palette, StructureTemplateBlockState{
			Name:       state.Name,
			Properties: state.Properties,
		})
	}
	for _, block := range raw.Blocks {
		if len(block.Pos) < 3 {
			continue
		}
		out.Blocks = append(out.Blocks, StructureTemplateBlock{
			Pos:   [3]int{int(block.Pos[0]), int(block.Pos[1]), int(block.Pos[2])},
			State: int(block.State),
			NBT:   block.NBT,
		})
	}
	return out, nil
}
