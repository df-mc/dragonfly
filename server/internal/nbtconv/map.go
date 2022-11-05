package nbtconv

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Map is a wrapper around a map[string]any that provides methods for reading
// and writing values to a map obtained from or prepared for NBT.
type Map map[string]any

func (m Map) Bool(k string) bool {
	return m.Uint8(k) == 1
}

func (m Map) Uint8(k string) uint8 {
	v, _ := m[k].(uint8)
	return v
}

func (m Map) String(k string) string {
	v, _ := m[k].(string)
	return v
}

func (m Map) Int16(k string) int16 {
	v, _ := m[k].(int16)
	return v
}

func (m Map) Int32(k string) int32 {
	v, _ := m[k].(int32)
	return v
}

func (m Map) Int64(k string) int64 {
	v, _ := m[k].(int64)
	return v
}

func (m Map) TickDuration(k string) time.Duration {
	return time.Duration(m.Int32(k)) * time.Millisecond * 50
}

func (m Map) Pos(k string) cube.Pos {
	if i, ok := m[k].([]any); ok {
		if len(i) != 3 {
			return cube.Pos{}
		}
		var v cube.Pos
		for index, f := range i {
			f32, _ := f.(int32)
			v[index] = int(f32)
		}
		return v
	} else if i, ok := m[k].([]int32); ok {
		if len(i) != 3 {
			return cube.Pos{}
		}
		return cube.Pos{int(i[0]), int(i[1]), int(i[2])}
	}
	return cube.Pos{}
}

func (m Map) Float32(k string) float32 {
	v, _ := m[k].(float32)
	return v
}

func (m Map) Float64(k string) float64 {
	v, _ := m[k].(float64)
	return v
}

func (m Map) Vec3(k string) mgl64.Vec3 {
	if i, ok := m[k].([]any); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		var v mgl64.Vec3
		for index, f := range i {
			f32, _ := f.(float32)
			v[index] = float64(f32)
		}
		return v
	} else if i, ok := m[k].([]float32); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		return mgl64.Vec3{float64(i[0]), float64(i[1]), float64(i[2])}
	}
	return mgl64.Vec3{}
}

func (m Map) Block(k string) world.Block {
	if mk, ok := m[k].(map[string]any); ok {
		name, _ := mk["name"].(string)
		properties, _ := mk["states"].(map[string]any)
		b, _ := world.BlockByName(name, properties)
		return b
	}
	return nil
}

func (m Map) Item(k string) item.Stack {
	if mk, ok := m[k].(map[string]any); ok {
		s := readItemStack(mk)
		readDamage(mk, &s, true)
		readEnchantments(mk, &s)
		readDisplay(mk, &s)
		readDragonflyData(mk, &s)
		return s
	}
	return item.Stack{}
}
