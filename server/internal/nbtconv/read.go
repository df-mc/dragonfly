package nbtconv

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/constraints"
)

// Bool reads a uint8 value from a map at key k and returns true if it equals 1.
func Bool(m map[string]any, k string) bool {
	return Uint8(m, k) == 1
}

// Uint8 reads a uint8 value from a map at key k.
func Uint8(m map[string]any, k string) uint8 {
	v, _ := m[k].(uint8)
	return v
}

// String reads a string value from a map at key k.
func String(m map[string]any, k string) string {
	v, _ := m[k].(string)
	return v
}

// Int16 reads an int16 value from a map at key k.
func Int16(m map[string]any, k string) int16 {
	v, _ := m[k].(int16)
	return v
}

// Int32 reads an int32 value from a map at key k.
func Int32(m map[string]any, k string) int32 {
	v, _ := m[k].(int32)
	return v
}

// Int64 reads an int16 value from a map at key k.
func Int64(m map[string]any, k string) int64 {
	v, _ := m[k].(int64)
	return v
}

// TickDuration reads a uint8/int16/in32 value from a map at key k and converts
// it from ticks to a time.Duration.
func TickDuration[T constraints.Integer](m map[string]any, k string) time.Duration {
	var v time.Duration
	switch any(*new(T)).(type) {
	case uint8:
		v = time.Duration(Uint8(m, k))
	case int16:
		v = time.Duration(Int16(m, k))
	case int32:
		v = time.Duration(Int32(m, k))
	default:
		panic("invalid tick duration value type")
	}
	return v * time.Millisecond * 50
}

// Float32 reads a float32 value from a map at key k.
func Float32(m map[string]any, k string) float32 {
	v, _ := m[k].(float32)
	return v
}

// Rotation reads a cube.Rotation from the map passed.
func Rotation(m map[string]any) cube.Rotation {
	return cube.Rotation{float64(Float32(m, "Yaw")), float64(Float32(m, "Pitch"))}
}

// Float64 reads a float64 value from a map at key k.
func Float64(m map[string]any, k string) float64 {
	v, _ := m[k].(float64)
	return v
}

// Slice reads a []any value from a map at key k.
func Slice(m map[string]any, k string) []any {
	v, _ := m[k].([]any)
	return v
}

// Vec3 converts x, y and z values in an NBT map to an mgl64.Vec3.
func Vec3(x map[string]any, k string) mgl64.Vec3 {
	if i, ok := x[k].([]any); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		var v mgl64.Vec3
		for index, f := range i {
			f32, _ := f.(float32)
			v[index] = float64(f32)
		}
		return v
	} else if i, ok := x[k].([]float32); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		return mgl64.Vec3{float64(i[0]), float64(i[1]), float64(i[2])}
	}
	return mgl64.Vec3{}
}

// Vec3ToFloat32Slice converts an mgl64.Vec3 to a []float32 with 3 elements.
func Vec3ToFloat32Slice(x mgl64.Vec3) []float32 {
	return []float32{float32(x[0]), float32(x[1]), float32(x[2])}
}

// Pos converts x, y and z values in an NBT map to a cube.Pos.
func Pos(x map[string]any, k string) cube.Pos {
	if i, ok := x[k].([]any); ok {
		if len(i) != 3 {
			return cube.Pos{}
		}
		var v cube.Pos
		for index, f := range i {
			f32, _ := f.(int32)
			v[index] = int(f32)
		}
		return v
	} else if i, ok := x[k].([]int32); ok {
		if len(i) != 3 {
			return cube.Pos{}
		}
		return cube.Pos{int(i[0]), int(i[1]), int(i[2])}
	}
	return cube.Pos{}
}

// PosToInt32Slice converts a cube.Pos to a []int32 with 3 elements.
func PosToInt32Slice(x cube.Pos) []int32 {
	return []int32{int32(x[0]), int32(x[1]), int32(x[2])}
}

// Block decodes the data of a block into a world.Block.
func Block(m map[string]any, k string) world.Block {
	if mk, ok := m[k].(map[string]any); ok {
		name, _ := mk["name"].(string)
		properties, _ := mk["states"].(map[string]any)
		b, _ := world.BlockByName(name, properties)
		return b
	}
	return nil
}
