package nbtconv

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// MapSlice reads an interface slice from a map at the key passed.
func MapSlice(m map[string]interface{}, key string) []interface{} {
	b, _ := m[key].([]interface{})
	return b
}

// MapString reads a string from a map at the key passed.
func MapString(m map[string]interface{}, key string) string {
	b, _ := m[key].(string)
	return b
}

// MapInt16 reads an int16 from a map at the key passed.
func MapInt16(m map[string]interface{}, key string) int16 {
	b, _ := m[key].(int16)
	return b
}

// MapInt32 reads an int32 from a map at the key passed.
func MapInt32(m map[string]interface{}, key string) int32 {
	b, _ := m[key].(int32)
	return b
}

// MapInt64 reads an int64 from a map at the key passed.
func MapInt64(m map[string]interface{}, key string) int64 {
	b, _ := m[key].(int64)
	return b
}

// MapByte reads a byte from a map at the key passed.
//noinspection GoCommentLeadingSpace
func MapByte(m map[string]interface{}, key string) byte {
	b, _ := m[key].(byte)
	return b
}

// MapFloat32 reads an float32 from a map at the key passed.
//noinspection GoCommentLeadingSpace
func MapFloat32(m map[string]interface{}, key string) float32 {
	b, _ := m[key].(float32)
	return b
}

// MapVec3 converts x, y and z values in an NBT map to an mgl64.Vec3.
func MapVec3(x map[string]interface{}, k string) mgl64.Vec3 {
	if i, ok := x[k].([]interface{}); ok {
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

// MapBlock converts a block's name and properties in a map obtained by decoding NBT to a world.Block.
func MapBlock(x map[string]interface{}, k string) world.Block {
	if m, ok := x[k].(map[string]interface{}); ok {
		return ReadBlock(m)
	}
	return nil
}

// MapItem converts an item's name, count, damage (and properties when it is a block) in a map obtained by decoding NBT
// to a world.Item.
func MapItem(x map[string]interface{}, k string) item.Stack {
	if m, ok := x[k].(map[string]interface{}); ok {
		s := readItemStack(m)
		readDamage(m, &s, true)
		readEnchantments(m, &s)
		readDisplay(m, &s)
		readDragonflyData(m, &s)
		return s
	}
	return item.Stack{}
}
