package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	_ "unsafe" // Imported for compiler directives.
)

// ItemFromNBT decodes the data of an item into an item stack.
func ItemFromNBT(data map[string]interface{}, s *item.Stack) item.Stack {
	disk := s == nil
	if disk {
		it, ok := world.ItemByName(readString(data, "Name"), readInt16(data, "Damage"))
		if !ok {
			return item.Stack{}
		}
		if nbt, ok := it.(world.NBTer); ok {
			it = nbt.DecodeNBT(data).(world.Item)
		}
		a := item.NewStack(it, int(readByte(data, "Count")))
		s = &a
		if _, ok := s.Item().(item.Durable); ok {
			*s = s.Damage(int(readInt16(data, "Damage")))
		}
	} else if !disk {
		if _, ok := s.Item().(item.Durable); ok {
			*s = s.Damage(int(readInt32(data, "Damage")))
		}
	}
	if displayInterface, ok := data["display"]; ok {
		display, ok := displayInterface.(map[string]interface{})
		if ok {
			if nameInterface, ok := display["Name"]; ok {
				if name, ok := nameInterface.(string); ok {
					*s = s.WithCustomName(name)
				}
			}
			if loreInterface, ok := display["Lore"]; ok {
				if lore, ok := loreInterface.([]string); ok {
					*s = s.WithLore(lore...)
				}
			}
		}
	}
	if enchantmentList, ok := data["ench"]; ok {
		enchantments, ok := enchantmentList.([]map[string]interface{})
		if ok {
			for _, ench := range enchantments {
				if e, ok := item.EnchantmentByID(int(readInt16(ench, "id"))); ok {
					e = e.WithLevel(int(readInt16(ench, "lvl")))
					*s = s.WithEnchantment(e)
				}
			}
		}
	}
	if customData, ok := data["dragonflyData"]; ok {
		d, _ := customData.([]byte)
		var m map[string]interface{}
		if err := gob.NewDecoder(bytes.NewBuffer(d)).Decode(&m); err != nil {
			panic("error decoding item user data: " + err.Error())
		}
		for k, v := range m {
			*s = s.WithValue(k, v)
		}
	}
	return *s
}

// ItemToNBT encodes an item stack to its NBT representation.
func ItemToNBT(s item.Stack, network bool) map[string]interface{} {
	m := make(map[string]interface{})
	if nbt, ok := s.Item().(world.NBTer); ok {
		m = nbt.EncodeNBT()
	}
	if !network {
		_, name, damage := s.Item().EncodeItem()
		m["Name"], m["Damage"] = name, damage
		m["Count"] = byte(s.Count())
		if _, ok := s.Item().(item.Durable); ok {
			m["Damage"] = int16(s.MaxDurability() - s.Durability())
		}
	} else if network {
		if _, ok := s.Item().(item.Durable); ok {
			m["Damage"] = int32(s.MaxDurability() - s.Durability())
		}
	}
	if s.CustomName() != "" {
		m["display"] = map[string]interface{}{"Name": s.CustomName()}
	}
	if len(s.Lore()) != 0 {
		if display, ok := m["display"]; ok {
			display.(map[string]interface{})["Lore"] = s.Lore()
		} else {
			m["display"] = map[string]interface{}{"Lore": s.Lore()}
		}
	}
	if len(s.Enchantments()) != 0 {
		var enchantments []map[string]interface{}
		for _, ench := range s.Enchantments() {
			if enchType, ok := item.EnchantmentID(ench); ok {
				enchantments = append(enchantments, map[string]interface{}{
					"id":  int16(enchType),
					"lvl": int16(ench.Level()),
				})
			}
		}
		m["ench"] = enchantments
	}
	if len(s.Values()) != 0 {
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(s.Values()); err != nil {
			panic("error encoding item user data: " + err.Error())
		}
		m["dragonflyData"] = buf.Bytes()
	}
	return m
}

// InvFromNBT decodes the data of an NBT slice into the inventory passed.
func InvFromNBT(inv *inventory.Inventory, items []interface{}) {
	for _, itemData := range items {
		data, _ := itemData.(map[string]interface{})
		it := ItemFromNBT(data, nil)
		if it.Empty() {
			continue
		}
		_ = inv.SetItem(int(readByte(data, "Slot")), it)
	}
}

// InvToNBT encodes an inventory to a data slice which may be encoded as NBT.
func InvToNBT(inv *inventory.Inventory) []map[string]interface{} {
	var items []map[string]interface{}
	for index, i := range inv.All() {
		if i.Empty() {
			continue
		}
		data := ItemToNBT(i, false)
		data["Slot"] = byte(index)
		items = append(items, data)
	}
	return items
}

// readByte reads a byte from a map at the key passed.
func readByte(m map[string]interface{}, key string) byte {
	v := m[key]
	b, _ := v.(byte)
	return b
}

// readInt16 reads an int16 from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readInt16(m map[string]interface{}, key string) int16 {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(int16)
	return b
}

// readInt32 reads an int32 from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readInt32(m map[string]interface{}, key string) int32 {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(int32)
	return b
}

// readString reads a string from a map at the key passed.
//noinspection GoCommentLeadingSpace
func readString(m map[string]interface{}, key string) string {
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	v, _ := m[key]
	b, _ := v.(string)
	return b
}
