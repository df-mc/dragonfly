package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Bool reads a boolean value from a map at key k.
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

// Int16 reads an int16 value from a map at key k.
func Int64(m map[string]any, k string) int64 {
	v, _ := m[k].(int64)
	return v
}

// TickDuration reads an int32 value from a map at key k and converts it from
// ticks to a time.Duration.
func TickDuration(m map[string]any, k string) time.Duration {
	return time.Duration(Int32(m, k)) * time.Millisecond * 50
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

// Item decodes the data of an item into an item stack.
func Item(data map[string]any, s *item.Stack) item.Stack {
	disk := s == nil
	if disk {
		a := readItemStack(data)
		s = &a
	}
	readDamage(data, s, disk)
	readAnvilCost(data, s)
	readDisplay(data, s)
	readEnchantments(data, s)
	readDragonflyData(data, s)
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
func readItemStack(m map[string]any) item.Stack {
	var it world.Item
	if blockItem, ok := Block(m, "Block").(world.Item); ok {
		it = blockItem
	}
	if v, ok := world.ItemByName(Read[string](m, "Name"), Read[int16](m, "Damage")); ok {
		it = v
	}
	if it == nil {
		return item.Stack{}
	}
	if n, ok := it.(world.NBTer); ok {
		it = n.DecodeNBT(m).(world.Item)
	}
	return item.NewStack(it, int(Read[byte](m, "Count")))
}

// readDamage reads the damage value stored in the NBT with the Damage tag and saves it to the item.Stack passed.
func readDamage(m map[string]any, s *item.Stack, disk bool) {
	if disk {
		*s = s.Damage(int(Read[int16](m, "Damage")))
		return
	}
	*s = s.Damage(int(Read[int32](m, "Damage")))
}

// readAnvilCost ...
func readAnvilCost(m map[string]any, s *item.Stack) {
	*s = s.WithAnvilCost(int(Read[int32](m, "RepairCost")))
}

// readEnchantments reads the enchantments stored in the ench tag of the NBT passed and stores it into an item.Stack.
func readEnchantments(m map[string]any, s *item.Stack) {
	enchantments, ok := m["ench"].([]map[string]any)
	if !ok {
		for _, e := range Read[[]any](m, "ench") {
			if v, ok := e.(map[string]any); ok {
				enchantments = append(enchantments, v)
			}
		}
	}
	for _, ench := range enchantments {
		if t, ok := item.EnchantmentByID(int(Read[int16](ench, "id"))); ok {
			*s = s.WithEnchantments(item.NewEnchantment(t, int(Read[int16](ench, "lvl"))))
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
