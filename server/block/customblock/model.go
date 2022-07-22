package customblock

import (
	"fmt"
)

// Model represents the model of a custom block. It can contain multiple materials applied to different parts of the
// model, as well as a reference to its geometry.
type Model struct {
	materials map[Target]Material
	geometry  Geometry
}

// NewModel returns a new Model with the provided information. If the size is larger than 16x16x16, this method will
// panic since the client does not allow models larger than a single block.
func NewModel(geometry Geometry) Model {
	if size := geometry.Size(); size.X() > 16 || size.Y() > 16 || size.Z() > 16 {
		panic(fmt.Errorf("model size cannot exceed 16x16x16, got %v", size))
	}
	return Model{
		materials: make(map[Target]Material),
		geometry:  geometry,
	}
}

// WithMaterial returns a copy of the Model with the provided material.
func (m Model) WithMaterial(target Target, material Material) Model {
	m.materials[target] = material
	return m
}

// Encode returns the model encoded as a map[string]any.
func (m Model) Encode() map[string]any {
	materials := map[string]any{}
	for target, material := range m.materials {
		materials[target.String()] = material.Encode()
	}
	model := map[string]any{"minecraft:material_instances": map[string]any{
		"mappings":  map[string]any{},
		"materials": materials,
	}}
	if identifier := m.geometry.Description.Identifier; len(identifier) > 0 {
		for k, v := range m.geometry.Encode() {
			model[k] = v
		}
		model["minecraft:geometry"] = map[string]any{"value": identifier}
	} else {
		model["minecraft:unit_cube"] = map[string]any{}
	}
	return model
}
