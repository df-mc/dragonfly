package gen

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

type WorldgenRegistry struct {
	mu sync.Mutex

	dimensionMetadata         map[string]DimensionMetadata
	biomeSourceParameterLists map[string]biomeSourceParameterListCacheEntry
	processorLists            map[string]processorListCacheEntry
	structures                map[string]structureCacheEntry
	structureSets             map[string]structureSetCacheEntry
	templatePools             map[string]templatePoolCacheEntry
	structureTemplates        map[string]structureTemplateCacheEntry
}

type biomeSourceParameterListCacheEntry struct {
	loaded bool
	def    MultiNoiseBiomeSourceParameterListDef
	err    error
}

type processorListCacheEntry struct {
	loaded bool
	def    ProcessorListDef
	err    error
}

type structureCacheEntry struct {
	loaded bool
	def    StructureDef
	err    error
}

type structureSetCacheEntry struct {
	loaded bool
	def    StructureSetDef
	err    error
}

type templatePoolCacheEntry struct {
	loaded bool
	def    TemplatePoolDef
	err    error
}

type structureTemplateCacheEntry struct {
	loaded bool
	data   []byte
	err    error
}

func NewWorldgenRegistry() *WorldgenRegistry {
	return &WorldgenRegistry{
		dimensionMetadata:         dimensionMetadataByName,
		biomeSourceParameterLists: make(map[string]biomeSourceParameterListCacheEntry),
		processorLists:            make(map[string]processorListCacheEntry),
		structures:                make(map[string]structureCacheEntry),
		structureSets:             make(map[string]structureSetCacheEntry),
		templatePools:             make(map[string]templatePoolCacheEntry),
		structureTemplates:        make(map[string]structureTemplateCacheEntry),
	}
}

func (r *WorldgenRegistry) DimensionMetadata(name string) (DimensionMetadata, error) {
	key := normalizeIdentifier(name)
	if metadata, ok := r.dimensionMetadata[key]; ok {
		return metadata, nil
	}
	if metadata, ok := r.dimensionMetadata["minecraft:"+key]; ok {
		return metadata, nil
	}
	return DimensionMetadata{}, fmt.Errorf("unknown noise settings %q", name)
}

func (r *WorldgenRegistry) StructureSetNames() []string {
	out := make([]string, 0, len(structureSetJSONByName))
	for name := range structureSetJSONByName {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

type MultiNoiseBiomeSourceParameterListDef struct {
	Preset string `json:"preset"`
}

func (r *WorldgenRegistry) BiomeSourceParameterList(name string) (MultiNoiseBiomeSourceParameterListDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.biomeSourceParameterLists[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := biomeSourceParameterListJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return MultiNoiseBiomeSourceParameterListDef{}, fmt.Errorf("unknown biome source parameter list %q", name)
	}

	var def MultiNoiseBiomeSourceParameterListDef
	err := json.Unmarshal([]byte(raw), &def)
	def.Preset = normalizeIdentifier(def.Preset)
	r.biomeSourceParameterLists[key] = biomeSourceParameterListCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

type ProcessorListDef struct {
	Processors []StructureProcessorDef `json:"processors"`
}

func (r *WorldgenRegistry) ProcessorList(name string) (ProcessorListDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.processorLists[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := processorListJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return ProcessorListDef{}, fmt.Errorf("unknown processor list %q", name)
	}

	var def ProcessorListDef
	err := json.Unmarshal([]byte(raw), &def)
	r.processorLists[key] = processorListCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

type StructureDef struct {
	Type string
	Raw  json.RawMessage
}

func (d *StructureDef) UnmarshalJSON(data []byte) error {
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

func (r *WorldgenRegistry) Structure(name string) (StructureDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.structures[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := structureJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return StructureDef{}, fmt.Errorf("unknown structure %q", name)
	}

	var def StructureDef
	err := json.Unmarshal([]byte(raw), &def)
	r.structures[key] = structureCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

type StructurePlacementDef struct {
	Type string
	Raw  json.RawMessage
}

func (d *StructurePlacementDef) UnmarshalJSON(data []byte) error {
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

type StructureSetDef struct {
	Placement  StructurePlacementDef  `json:"placement"`
	Structures []WeightedStructureRef `json:"structures"`
}

type WeightedStructureRef struct {
	Structure string `json:"structure"`
	Weight    int    `json:"weight"`
}

func (r *WorldgenRegistry) StructureSet(name string) (StructureSetDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.structureSets[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := structureSetJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return StructureSetDef{}, fmt.Errorf("unknown structure set %q", name)
	}

	var def StructureSetDef
	err := json.Unmarshal([]byte(raw), &def)
	r.structureSets[key] = structureSetCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

type TemplatePoolDef struct {
	Fallback string              `json:"fallback"`
	Elements []TemplatePoolEntry `json:"elements"`
}

type TemplatePoolEntry struct {
	Element TemplatePoolElementDef `json:"element"`
	Weight  int                    `json:"weight"`
}

type TemplatePoolElementDef struct {
	ElementType string
	Raw         json.RawMessage
}

func (d *TemplatePoolElementDef) UnmarshalJSON(data []byte) error {
	var raw struct {
		ElementType string `json:"element_type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	d.ElementType = normalizeIdentifier(raw.ElementType)
	d.Raw = append(json.RawMessage(nil), data...)
	return nil
}

func (r *WorldgenRegistry) TemplatePool(name string) (TemplatePoolDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.templatePools[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := templatePoolJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return TemplatePoolDef{}, fmt.Errorf("unknown template pool %q", name)
	}

	var def TemplatePoolDef
	err := json.Unmarshal([]byte(raw), &def)
	r.templatePools[key] = templatePoolCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

func (r *WorldgenRegistry) StructureTemplate(name string) ([]byte, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.structureTemplates[key]; ok && entry.loaded {
		r.mu.Unlock()
		return append([]byte(nil), entry.data...), entry.err
	}
	raw, ok := structureTemplateBase64ByName[key]
	if !ok {
		r.mu.Unlock()
		return nil, fmt.Errorf("unknown structure template %q", name)
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		data = append([]byte(nil), data...)
	}
	r.structureTemplates[key] = structureTemplateCacheEntry{loaded: true, data: data, err: err}
	r.mu.Unlock()
	return append([]byte(nil), data...), err
}
