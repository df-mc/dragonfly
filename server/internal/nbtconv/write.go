package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"sort"
)

// WriteItem encodes an item stack into a map that can be encoded using NBT.
func WriteItem(s item.Stack, disk bool) map[string]any {
	tag := make(map[string]any)
	if nbt, ok := s.Item().(world.NBTer); ok {
		for k, v := range nbt.EncodeNBT() {
			tag[k] = v
		}
	}
	writeAnvilCost(tag, s)
	writeDamage(tag, s, disk)
	writeDisplay(tag, s)
	writeDragonflyData(tag, s)
	writeEnchantments(tag, s)

	data := make(map[string]any)
	if disk {
		writeItemStack(data, tag, s)
	} else {
		for k, v := range tag {
			data[k] = v
		}
	}
	return data
}

// WriteBlock encodes a world.Block into a map that can be encoded using NBT.
func WriteBlock(b world.Block) map[string]any {
	name, properties := b.EncodeBlock()
	return map[string]any{
		"name":    name,
		"states":  properties,
		"version": chunk.CurrentBlockVersion,
	}
}

// writeItemStack writes the name, metadata value, count and NBT of an item to a map ready for NBT encoding.
func writeItemStack(m, t map[string]any, s item.Stack) {
	m["Name"], m["Damage"] = s.Item().EncodeItem()
	if b, ok := s.Item().(world.Block); ok {
		v := map[string]any{}
		writeBlock(v, b)
		m["Block"] = v
	}
	m["Count"] = byte(s.Count())
	if len(t) > 0 {
		m["tag"] = t
	}
}

// writeBlock writes the name, properties and version of a block to a map ready for NBT encoding.
func writeBlock(m map[string]any, b world.Block) {
	m["name"], m["states"] = b.EncodeBlock()
	m["version"] = chunk.CurrentBlockVersion
}

// writeDragonflyData writes additional data associated with an item.Stack to a map for NBT encoding.
func writeDragonflyData(m map[string]any, s item.Stack) {
	if v := s.Values(); len(v) != 0 {
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(mapToSlice(v)); err != nil {
			panic("error encoding item user data: " + err.Error())
		}
		m["dragonflyData"] = buf.Bytes()
	}
}

// mapToSlice converts a map to a slice of the type mapValue and orders the slice by the keys in the map to ensure a
// deterministic order.
func mapToSlice(m map[string]any) []mapValue {
	values := make([]mapValue, 0, len(m))
	for k, v := range m {
		values = append(values, mapValue{K: k, V: v})
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].K < values[j].K
	})
	return values
}

// mapValue represents a value in a map. It is used to convert maps to a slice and order the slice before encoding to
// NBT to ensure a deterministic output.
type mapValue struct {
	K string
	V any
}

// writeEnchantments writes the enchantments of an item to a map for NBT encoding.
func writeEnchantments(m map[string]any, s item.Stack) {
	if len(s.Enchantments()) != 0 {
		var enchantments []map[string]any
		for _, e := range s.Enchantments() {
			if eType, ok := item.EnchantmentID(e.Type()); ok {
				enchantments = append(enchantments, map[string]any{
					"id":  int16(eType),
					"lvl": int16(e.Level()),
				})
			}
		}
		m["ench"] = enchantments
	}
}

// writeDisplay writes the display name and lore of an item to a map for NBT encoding.
func writeDisplay(m map[string]any, s item.Stack) {
	name, lore := s.CustomName(), s.Lore()
	v := map[string]any{}
	if name != "" {
		v["Name"] = name
	}
	if len(lore) != 0 {
		v["Lore"] = lore
	}
	if len(v) != 0 {
		m["display"] = v
	}
}

// writeDamage writes the damage to an item.Stack (either an int16 for disk or int32 for network) to a map for NBT
// encoding.
func writeDamage(m map[string]any, s item.Stack, disk bool) {
	if v, ok := m["Damage"]; !ok || v.(int16) == 0 {
		if _, ok := s.Item().(item.Durable); ok {
			if disk {
				m["Damage"] = int16(s.MaxDurability() - s.Durability())
			} else {
				m["Damage"] = int32(s.MaxDurability() - s.Durability())
			}
		}
	}
}

// writeAnvilCost ...
func writeAnvilCost(m map[string]any, s item.Stack) {
	if cost := s.AnvilCost(); cost > 0 {
		m["RepairCost"] = int32(cost)
	}
}
