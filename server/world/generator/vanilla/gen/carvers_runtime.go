package gen

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type CarverRegistry struct {
	mu              sync.Mutex
	configuredCache map[string]configuredCarverCacheEntry
}

type configuredCarverCacheEntry struct {
	loaded bool
	def    ConfiguredCarverDef
	err    error
}

func NewCarverRegistry() *CarverRegistry {
	return &CarverRegistry{
		configuredCache: make(map[string]configuredCarverCacheEntry),
	}
}

func (r *CarverRegistry) BiomeCarvers(biomeName string) []string {
	carvers, ok := biomeCarversByName[biomeName]
	if !ok {
		return nil
	}
	return append([]string(nil), carvers...)
}

func (r *CarverRegistry) Configured(name string) (ConfiguredCarverDef, error) {
	key := normalizeIdentifier(name)

	r.mu.Lock()
	if entry, ok := r.configuredCache[key]; ok && entry.loaded {
		r.mu.Unlock()
		return entry.def, entry.err
	}
	raw, ok := configuredCarverJSONByName[key]
	if !ok {
		r.mu.Unlock()
		return ConfiguredCarverDef{}, fmt.Errorf("unknown configured carver %q", name)
	}

	var def ConfiguredCarverDef
	err := json.Unmarshal([]byte(raw), &def)
	r.configuredCache[key] = configuredCarverCacheEntry{loaded: true, def: def, err: err}
	r.mu.Unlock()
	return def, err
}

type ConfiguredCarverDef struct {
	Type   string
	Config json.RawMessage
}

func (f *ConfiguredCarverDef) UnmarshalJSON(data []byte) error {
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

func (f ConfiguredCarverDef) Cave() (CaveCarverConfig, error) {
	return decodeCarverConfig[CaveCarverConfig](f, "cave", "nether_cave")
}

func (f ConfiguredCarverDef) Canyon() (CanyonCarverConfig, error) {
	return decodeCarverConfig[CanyonCarverConfig](f, "canyon")
}

type CaveCarverConfig struct {
	Probability                float64        `json:"probability"`
	Y                          HeightProvider `json:"y"`
	YScale                     FloatProvider  `json:"yScale"`
	LavaLevel                  VerticalAnchor `json:"lava_level"`
	Replaceable                string         `json:"replaceable"`
	HorizontalRadiusMultiplier FloatProvider  `json:"horizontal_radius_multiplier"`
	VerticalRadiusMultiplier   FloatProvider  `json:"vertical_radius_multiplier"`
	FloorLevel                 FloatProvider  `json:"floor_level"`
}

type CanyonCarverConfig struct {
	Probability      float64           `json:"probability"`
	Y                HeightProvider    `json:"y"`
	YScale           FloatProvider     `json:"yScale"`
	LavaLevel        VerticalAnchor    `json:"lava_level"`
	Replaceable      string            `json:"replaceable"`
	VerticalRotation FloatProvider     `json:"vertical_rotation"`
	Shape            CanyonShapeConfig `json:"shape"`
}

type CanyonShapeConfig struct {
	DistanceFactor              FloatProvider `json:"distance_factor"`
	Thickness                   FloatProvider `json:"thickness"`
	WidthSmoothness             int           `json:"width_smoothness"`
	HorizontalRadiusFactor      FloatProvider `json:"horizontal_radius_factor"`
	VerticalRadiusDefaultFactor float64       `json:"vertical_radius_default_factor"`
	VerticalRadiusCenterFactor  float64       `json:"vertical_radius_center_factor"`
}

type FloatProvider struct {
	Kind     string
	Constant *float64
	Min      float64
	Max      float64
	Plateau  float64
	Raw      json.RawMessage
}

func (p *FloatProvider) UnmarshalJSON(data []byte) error {
	p.Raw = append(json.RawMessage(nil), data...)

	var constant float64
	if err := json.Unmarshal(data, &constant); err == nil {
		p.Kind = "constant"
		p.Constant = &constant
		p.Min = constant
		p.Max = constant
		p.Plateau = 0
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
	p.Plateau = 0

	switch p.Kind {
	case "uniform":
		var raw struct {
			MinInclusive float64  `json:"min_inclusive"`
			MaxExclusive *float64 `json:"max_exclusive"`
			MaxInclusive *float64 `json:"max_inclusive"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.Min = raw.MinInclusive
		switch {
		case raw.MaxExclusive != nil:
			p.Max = *raw.MaxExclusive
		case raw.MaxInclusive != nil:
			p.Max = *raw.MaxInclusive
		default:
			p.Max = raw.MinInclusive
		}
	case "trapezoid":
		var raw struct {
			Min     float64 `json:"min"`
			Max     float64 `json:"max"`
			Plateau float64 `json:"plateau"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		p.Min = raw.Min
		p.Max = raw.Max
		p.Plateau = raw.Plateau
	default:
		return fmt.Errorf("unsupported float provider type %q", probe.Type)
	}
	return nil
}

func decodeCarverConfig[T any](f ConfiguredCarverDef, expectedTypes ...string) (T, error) {
	var out T
	for _, expectedType := range expectedTypes {
		if f.Type == expectedType {
			if err := json.Unmarshal(f.Config, &out); err != nil {
				return out, err
			}
			return out, nil
		}
	}
	return out, fmt.Errorf("expected %s, got %s", strings.Join(expectedTypes, "/"), f.Type)
}
