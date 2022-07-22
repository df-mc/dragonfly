package customblock

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// Model represents the model of a custom block. It can contain multiple materials applied to different parts of the
// model, as well as a reference to its geometry.
type Model struct {
	materials  map[MaterialTarget]Material
	geometries []Geometry
	origin     mgl64.Vec3
	size       mgl64.Vec3
}

// NewModel returns a new Model with the provided information. If the size is larger than 16x16x16, this method will
// panic since the client does not allow models larger than a single block.
func NewModel(geometries []Geometry, origin, size mgl64.Vec3) Model {
	if size.X() > 16 || size.Y() > 16 || size.Z() > 16 {
		panic(fmt.Errorf("model size cannot exceed 16x16x16, got %v", size))
	}
	return Model{
		materials:  make(map[MaterialTarget]Material),
		geometries: geometries,
		origin:     origin,
		size:       size,
	}
}

// WithMaterial returns a copy of the Model with the provided material.
func (m Model) WithMaterial(target MaterialTarget, material Material) Model {
	m.materials[target] = material
	return m
}

// Encode returns the model encoded as a map[string]any.
func (m Model) Encode() map[string]any {
	materials := map[string]any{}
	for target, material := range m.materials {
		materials[target.String()] = material.Encode()
	}
	origin := vec64To32(m.origin)
	size := vec64To32(m.size)
	model := map[string]any{
		"minecraft:material_instances": map[string]any{
			"mappings":  map[string]any{},
			"materials": materials,
		},
		"minecraft:pick_collision": map[string]any{
			"enabled": uint8(1),
			"origin":  origin[:],
			"size":    size[:],
		},
	}
	if len(m.geometries) > 0 {
		model["minecraft:geometry"] = map[string]any{
			"value": m.geometries[0].Description.Identifier,
		}
	} else {
		model["minecraft:unit_cube"] = map[string]any{}
	}
	return model
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}
