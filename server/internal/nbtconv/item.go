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
		name := MapString(data, "Name")
		var it world.Item
		if states, ok := data["States"].(map[string]interface{}); ok {
			block, ok := world.BlockByName(name, states)
			if !ok {
				return item.Stack{}
			}
			it, ok = block.(world.Item)
			if !ok {
				return item.Stack{}
			}
		} else {
			it, ok = world.ItemByName(name, MapInt16(data, "Damage"))
			if !ok {
				return item.Stack{}
			}
		}
		if nbt, ok := it.(world.NBTer); ok {
			it = nbt.DecodeNBT(data).(world.Item)
		}
		a := item.NewStack(it, int(MapByte(data, "Count")))
		s = &a
		if _, ok := s.Item().(item.Durable); ok {
			*s = s.Damage(int(MapInt16(data, "Damage")))
		}
	} else {
		if _, ok := s.Item().(item.Durable); ok {
			*s = s.Damage(int(MapInt32(data, "Damage")))
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
				} else if lore, ok := loreInterface.([]interface{}); ok {
					loreLines := make([]string, 0, len(lore))
					for _, l := range lore {
						loreLines = append(loreLines, l.(string))
					}
					*s = s.WithLore(loreLines...)
				}
			}
		}
	}
	if enchantmentList, ok := data["ench"]; ok {
		enchantments, ok := enchantmentList.([]map[string]interface{})
		if !ok {
			if enchantments2, ok := enchantmentList.([]interface{}); ok {
				for _, e := range enchantments2 {
					enchantments = append(enchantments, e.(map[string]interface{}))
				}
			}
		}
		for _, ench := range enchantments {
			if e, ok := item.EnchantmentByID(int(MapInt16(ench, "id"))); ok {
				e = e.WithLevel(int(MapInt16(ench, "lvl")))
				*s = s.WithEnchantment(e)
			}
		}
	}
	if customData, ok := data["dragonflyData"]; ok {
		d, ok := customData.([]byte)
		if !ok {
			if itf, ok := customData.([]interface{}); ok {
				for _, v := range itf {
					b, _ := v.(byte)
					d = append(d, b)
				}
			}
		}
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
		if b, ok := s.Item().(world.Block); ok {
			m["Name"], m["States"] = b.EncodeBlock()
		} else {
			m["Name"], m["Damage"] = s.Item().EncodeItem()
		}
		m["Count"] = byte(s.Count())
		if _, ok := s.Item().(item.Durable); ok {
			m["Damage"] = int16(s.MaxDurability() - s.Durability())
		}
	} else {
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
		_ = inv.SetItem(int(MapByte(data, "Slot")), it)
	}
}

// InvToNBT encodes an inventory to a data slice which may be encoded as NBT.
func InvToNBT(inv *inventory.Inventory) []map[string]interface{} {
	var items []map[string]interface{}
	for index, i := range inv.Items() {
		if i.Empty() {
			continue
		}
		data := ItemToNBT(i, false)
		data["Slot"] = byte(index)
		items = append(items, data)
	}
	return items
}
