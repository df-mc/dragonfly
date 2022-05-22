package customblock

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/maps"
)

// Model represents the model of a custom block. It can contain multiple materials applied to different parts of the
// model, as well as a reference to its geometry.
type Model struct {
	materials    map[MaterialTarget]Material
	geometryName string
	origin       mgl64.Vec3
	size         mgl64.Vec3
}

// NewModel returns a new Model with the provided information. If the size is larger than 16x16x16, this method will
// panic since the client does not allow models larger than a single block.
func NewModel(geometryName string, origin, size mgl64.Vec3) Model {
	if size.X() > 16 || size.Y() > 16 || size.Z() > 16 {
		panic(fmt.Errorf("model size cannot exceed 16x16x16, got %v", size))
	}
	return Model{
		geometryName: geometryName,
		origin:       origin,
		size:         size,
	}
}

// WithMaterial returns a copy of the Model with the provided material.
func (m Model) WithMaterial(target MaterialTarget, material Material) Model {
	m.materials = maps.Clone(m.materials)
	m.materials[target] = material
	return m
}

// Encode returns the model encoded as a map[string]any.
func (m Model) Encode() map[string]any {
	materials := map[MaterialTarget]any{}
	for target, material := range m.materials {
		materials[target] = material.Encode()
	}
	origin := vec64To32(m.origin)
	size := vec64To32(m.size)
	return map[string]any{
		"minecraft:material_instances": map[string]any{
			"mappings":  map[string]any{},
			"materials": materials,
		},
		"minecraft:geometry": map[string]any{
			"value": m.geometryName,
		},
		"minecraft:pick_collision": map[string]any{
			"enabled": uint8(1),
			"origin":  origin[:],
			"size":    size[:],
		},
	}
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}
