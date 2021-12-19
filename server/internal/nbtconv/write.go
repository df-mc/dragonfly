package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

// WriteItem encodes an item stack into a map that can be encoded using NBT.
func WriteItem(s item.Stack, disk bool) map[string]interface{} {
	m := make(map[string]interface{})
	if nbt, ok := s.Item().(world.NBTer); ok {
		for k, v := range nbt.EncodeNBT() {
			m[k] = v
		}
	}
	if disk {
		writeItemStack(m, s)
	}
	writeDamage(m, s, disk)
	writeDisplay(m, s)
	writeEnchantments(m, s)
	writeDragonflyData(m, s)
	return m
}

// WriteBlock encodes a world.Block into a map that can be encoded using NBT.
func WriteBlock(b world.Block) map[string]interface{} {
	name, properties := b.EncodeBlock()
	return map[string]interface{}{
		"name":    name,
		"states":  properties,
		"version": chunk.CurrentBlockVersion,
	}
}

// writeItemStack writes the name, metadata value, count and NBT of an item to a map ready for NBT encoding.
func writeItemStack(m map[string]interface{}, s item.Stack) {
	m["Name"], m["Damage"] = s.Item().EncodeItem()
	if b, ok := s.Item().(world.Block); ok {
		v := map[string]interface{}{}
		writeBlock(v, b)
		m["Block"] = v
	}
	m["Count"] = byte(s.Count())
}

// writeBlock writes the name, properties and version of a block to a map ready for NBT encoding.
func writeBlock(m map[string]interface{}, b world.Block) {
	m["name"], m["states"] = b.EncodeBlock()
	m["version"] = chunk.CurrentBlockVersion
}

// writeDragonflyData writes additional data associated with an item.Stack to a map for NBT encoding.
func writeDragonflyData(m map[string]interface{}, s item.Stack) {
	if len(s.Values()) != 0 {
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(s.Values()); err != nil {
			panic("error encoding item user data: " + err.Error())
		}
		m["dragonflyData"] = buf.Bytes()
	}
}

// writeEnchantments writes the enchantments of an item to a map for NBT encoding.
func writeEnchantments(m map[string]interface{}, s item.Stack) {
	if len(s.Enchantments()) != 0 {
		var enchantments []map[string]interface{}
		for _, e := range s.Enchantments() {
			if eType, ok := item.EnchantmentID(e); ok {
				enchantments = append(enchantments, map[string]interface{}{
					"id":  int16(eType),
					"lvl": int16(e.Level()),
				})
			}
		}
		m["ench"] = enchantments
	}
}

// writeDisplay writes the display name and lore of an item to a map for NBT encoding.
func writeDisplay(m map[string]interface{}, s item.Stack) {
	name, lore := s.CustomName(), s.Lore()
	v := map[string]interface{}{}
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

// writeDamage writes the damage of an item.Stack (either an int16 for disk or int32 for network) to a map for NBT
// encoding.
func writeDamage(m map[string]interface{}, s item.Stack, disk bool) {
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
