package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

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
		it := ReadItem(data, nil)
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
