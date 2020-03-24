package nbtconv

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	_ "unsafe" // Imported for compiler directives.
)

//go:linkname world_itemByName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.itemByName
//noinspection ALL
func world_itemByName(name string, meta int16) (world.Item, bool)

//go:linkname world_itemToName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.itemToName
//noinspection ALL
func world_itemToName(it world.Item) (name string, meta int16)

// ItemFromNBT decodes the data of an item into an item stack.
func ItemFromNBT(data map[string]interface{}, s *item.Stack) item.Stack {
	if s == nil {
		it, ok := world_itemByName(readString(data, "Name"), readInt16(data, "Damage"))
		if !ok {
			return item.Stack{}
		}
		if nbt, ok := it.(world.NBTer); ok {
			it = nbt.DecodeNBT(data).(world.Item)
		}
		a := item.NewStack(it, int(readByte(data, "Count")))
		s = &a
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
				if lore, ok := loreInterface.([]interface{}); ok {
					loreLines := make([]string, 0, len(lore))
					for _, l := range lore {
						loreLines = append(loreLines, l.(string))
					}
					*s = s.WithLore(loreLines...)
				}
			}
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
		m["Name"], m["Damage"] = world_itemToName(s.Item())
		m["Count"] = byte(s.Count())
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
	v, _ := m[key]
	b, _ := v.(byte)
	return b
}

// readInt16 reads an int16 from a map at the key passed.
func readInt16(m map[string]interface{}, key string) int16 {
	v, _ := m[key]
	b, _ := v.(int16)
	return b
}

// readString reads a string from a map at the key passed.
func readString(m map[string]interface{}, key string) string {
	v, _ := m[key]
	b, _ := v.(string)
	return b
}
