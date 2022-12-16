package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/constraints"
	"time"
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

// MapItem converts an item's name, count, damage (and properties when it is a block) in a map obtained by decoding NBT
// to a world.Item.
func MapItem(x map[string]any, k string) item.Stack {
	if m, ok := x[k].(map[string]any); ok {
		tag, ok := m["tag"].(map[string]any)
		if !ok {
			tag = map[string]any{}
		}

		s := readItemStack(m, tag)
		readDamage(tag, &s, true)
		readEnchantments(tag, &s)
		readDisplay(tag, &s)
		readDragonflyData(tag, &s)
		return s
	}
	return item.Stack{}
}

// Item decodes the data of an item into an item stack.
func Item(data map[string]any, s *item.Stack) item.Stack {
	disk, tag := s == nil, data
	if disk {
		t, ok := data["tag"].(map[string]any)
		if !ok {
			t = map[string]any{}
		}
		tag = t

		a := readItemStack(data, tag)
		s = &a
	}

	readAnvilCost(tag, s)
	readDamage(tag, s, disk)
	readDisplay(tag, s)
	readDragonflyData(tag, s)
	readEnchantments(tag, s)
	return *s
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

// readItemStack reads an item.Stack from the NBT in the map passed.
func readItemStack(m, t map[string]any) item.Stack {
	var it world.Item
	if blockItem, ok := Block(m, "Block").(world.Item); ok {
		it = blockItem
	}
	if v, ok := world.ItemByName(String(m, "Name"), Int16(m, "Damage")); ok {
		it = v
	}
	if it == nil {
		return item.Stack{}
	}
	if n, ok := it.(world.NBTer); ok {
		it = n.DecodeNBT(t).(world.Item)
	}
	return item.NewStack(it, int(Uint8(m, "Count")))
}

// readDamage reads the damage value stored in the NBT with the Damage tag and saves it to the item.Stack passed.
func readDamage(m map[string]any, s *item.Stack, disk bool) {
	if disk {
		*s = s.Damage(int(Int16(m, "Damage")))
		return
	}
	*s = s.Damage(int(Int32(m, "Damage")))
}

// readAnvilCost ...
func readAnvilCost(m map[string]any, s *item.Stack) {
	*s = s.WithAnvilCost(int(Int32(m, "RepairCost")))
}

// readEnchantments reads the enchantments stored in the ench tag of the NBT passed and stores it into an item.Stack.
func readEnchantments(m map[string]any, s *item.Stack) {
	enchantments, ok := m["ench"].([]map[string]any)
	if !ok {
		for _, e := range Slice(m, "ench") {
			if v, ok := e.(map[string]any); ok {
				enchantments = append(enchantments, v)
			}
		}
	}
	for _, ench := range enchantments {
		if t, ok := item.EnchantmentByID(int(Int16(ench, "id"))); ok {
			*s = s.WithEnchantments(item.NewEnchantment(t, int(Int16(ench, "lvl"))))
		}
	}
}

// readDisplay reads the display data present in the display field in the NBT. It includes a custom name of the item
// and the lore.
func readDisplay(m map[string]any, s *item.Stack) {
	if display, ok := m["display"].(map[string]any); ok {
		if name, ok := display["Name"].(string); ok {
			// Only add the custom name if actually set.
			*s = s.WithCustomName(name)
		}
		if lore, ok := display["Lore"].([]string); ok {
			*s = s.WithLore(lore...)
		} else if lore, ok := display["Lore"].([]any); ok {
			loreLines := make([]string, 0, len(lore))
			for _, l := range lore {
				loreLines = append(loreLines, l.(string))
			}
			*s = s.WithLore(loreLines...)
		}
	}
}

// readDragonflyData reads data written to the dragonflyData field in the NBT of an item and adds it to the item.Stack
// passed.
func readDragonflyData(m map[string]any, s *item.Stack) {
	if customData, ok := m["dragonflyData"]; ok {
		d, ok := customData.([]byte)
		if !ok {
			if itf, ok := customData.([]any); ok {
				for _, v := range itf {
					b, _ := v.(byte)
					d = append(d, b)
				}
			}
		}
		var values []mapValue
		if err := gob.NewDecoder(bytes.NewBuffer(d)).Decode(&values); err != nil {
			panic("error decoding item user data: " + err.Error())
		}
		for _, val := range values {
			*s = s.WithValue(val.K, val.V)
		}
	}
}
