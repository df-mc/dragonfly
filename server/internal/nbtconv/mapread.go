package nbtconv

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// MapSlice reads an interface slice from a map at the key passed.
func MapSlice(m map[string]any, key string) []any {
	return safeRead[[]any](m, key)
}

// MapString reads a string from a map at the key passed.
func MapString(m map[string]any, key string) string {
	return safeRead[string](m, key)
}

// MapInt16 reads an int16 from a map at the key passed.
func MapInt16(m map[string]any, key string) int16 {
	return safeRead[int16](m, key)
}

// MapInt32 reads an int32 from a map at the key passed.
func MapInt32(m map[string]any, key string) int32 {
	return safeRead[int32](m, key)
}

// MapInt64 reads an int64 from a map at the key passed.
func MapInt64(m map[string]any, key string) int64 {
	return safeRead[int64](m, key)
}

// MapByte reads a byte from a map at the key passed.
func MapByte(m map[string]any, key string) byte {
	return safeRead[byte](m, key)
}

// MapFloat32 reads a float32 from a map at the key passed.
func MapFloat32(m map[string]any, key string) float32 {
	return safeRead[float32](m, key)
}

// safeRead reads a value of the type T from the map passed. safeRead never panics. If the key was not found in the map
// or if the value was of a different type, the default value of type T is returned.
func safeRead[T any](m map[string]any, key string) T {
	v, _ := m[key].(T)
	return v
}

// MapVec3 converts x, y and z values in an NBT map to an mgl64.Vec3.
func MapVec3(x map[string]any, k string) mgl64.Vec3 {
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

// MapPos converts x, y and z values in an NBT map to a cube.Pos.
func MapPos(x map[string]any, k string) cube.Pos {
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

// MapBlock converts a block's name and properties in a map obtained by decoding NBT to a world.Block.
func MapBlock(x map[string]any, k string) world.Block {
	if m, ok := x[k].(map[string]any); ok {
		return ReadBlock(m)
	}
	return nil
}

// MapItem converts an item's name, count, damage (and properties when it is a block) in a map obtained by decoding NBT
// to a world.Item.
func MapItem(x map[string]any, k string) item.Stack {
	if m, ok := x[k].(map[string]any); ok {
		s := readItemStack(m)
		readDamage(m, &s, true)
		readEnchantments(m, &s)
		readDisplay(m, &s)
		readDragonflyData(m, &s)
		return s
	}
	return item.Stack{}
}
