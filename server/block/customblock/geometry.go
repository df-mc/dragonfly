package customblock

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Geometries represents the JSON structure of a vanilla geometry file. It contains a format version and a slice of
// unique geometries.
type Geometries struct {
	FormatVersion string     `json:"format_version"`
	Geometry      []Geometry `json:"minecraft:geometry"`
}

// Geometry represents a single geometry that contains bones and other information.
type Geometry struct {
	Description struct {
		Identifier          string     `json:"identifier"`
		TextureWidth        int        `json:"texture_width"`
		TextureHeight       int        `json:"texture_height"`
		VisibleBoundsWidth  float64    `json:"visible_bounds_width"`
		VisibleBoundsHeight float64    `json:"visible_bounds_height"`
		VisibleBoundsOffset mgl64.Vec3 `json:"visible_bounds_offset"`
	} `json:"description"`
	Bones []struct {
		Name     string     `json:"name"`
		Pivot    mgl64.Vec3 `json:"pivot,omitempty"`
		Rotation mgl64.Vec3 `json:"rotation,omitempty"`
		Cubes    []struct {
			Origin   mgl64.Vec3 `json:"origin"`
			Size     mgl64.Vec3 `json:"size"`
			UV       any        `json:"uv"`
			Pivot    mgl64.Vec3 `json:"pivot,omitempty"`
			Rotation mgl64.Vec3 `json:"rotation,omitempty"`
			Inflate  float64    `json:"inflate,omitempty"`
		} `json:"cubes"`
	} `json:"bones"`
}

// Encode encodes the geometry into a JSON component.
func (g Geometry) Encode() map[string]any {
	origin, size := vec64To32(g.Origin()), vec64To32(g.Size())
	box := map[string]any{
		"enabled": byte(0x1),
		"origin": []float32{
			origin.X(),
			origin.Y(),
			origin.Z(),
		},
		"size": []float32{
			size.X(),
			size.Y(),
			size.Z(),
		},
	}
	return map[string]any{
		"minecraft:aim_collision": box,
		"minecraft:collision_box": box,
		"minecraft:pick_collision": map[string]any{
			"enabled": uint8(1),
			"origin":  origin[:],
			"size":    size[:],
		},
	}
}

// Origin returns the origin of the geometry. It is calculated by using the smallest origin points of all cubes.
func (g Geometry) Origin() (x mgl64.Vec3) {
	for _, bone := range g.Bones {
		for _, cube := range bone.Cubes {
			x[0] = math.Min(x[0], cube.Origin.X())
			x[1] = math.Min(x[1], cube.Origin.Y())
			x[2] = math.Min(x[2], cube.Origin.Z())
		}
	}
	return
}

// Size returns the size of the geometry. It is calculated by using the largest size of all cubes.
func (g Geometry) Size() (x mgl64.Vec3) {
	for _, bone := range g.Bones {
		for _, cube := range bone.Cubes {
			x[0] = math.Max(x[0], math.Abs(cube.Size.X()))
			x[1] = math.Max(x[1], math.Abs(cube.Size.Y()))
			x[2] = math.Max(x[2], math.Abs(cube.Size.Z()))
		}
	}
	return
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}
