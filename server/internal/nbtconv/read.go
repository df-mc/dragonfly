package nbtconv

import (
	"bytes"
	"encoding/gob"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// ReadItem decodes the data of an item into an item stack.
func ReadItem(data map[string]any, s *item.Stack) item.Stack {
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

// ReadBlock decodes the data of a block into a world.Block.
func ReadBlock(m map[string]any) world.Block {
	name, _ := m["name"].(string)
	properties, _ := m["states"].(map[string]any)
	b, _ := world.BlockByName(name, properties)
	return b
}

// readItemStack reads an item.Stack from the NBT in the map passed.
func readItemStack(m map[string]any) item.Stack {
	var it world.Item
	if blockItem, ok := MapBlock(m, "Block").(world.Item); ok {
		it = blockItem
	}
	if v, ok := world.ItemByName(Map[string](m, "Name"), Map[int16](m, "Damage")); ok {
		it = v
	}
	if it == nil {
		return item.Stack{}
	}
	if n, ok := it.(world.NBTer); ok {
		it = n.DecodeNBT(m).(world.Item)
	}
	return item.NewStack(it, int(Map[byte](m, "Count")))
}

// readDamage reads the damage value stored in the NBT with the Damage tag and saves it to the item.Stack passed.
func readDamage(m map[string]any, s *item.Stack, disk bool) {
	if disk {
		*s = s.Damage(int(Map[int16](m, "Damage")))
		return
	}
	*s = s.Damage(int(Map[int32](m, "Damage")))
}

// readAnvilCost ...
func readAnvilCost(m map[string]any, s *item.Stack) {
	*s = s.WithAnvilCost(int(Map[int32](m, "RepairCost")))
}

// readEnchantments reads the enchantments stored in the ench tag of the NBT passed and stores it into an item.Stack.
func readEnchantments(m map[string]any, s *item.Stack) {
	enchantments, ok := m["ench"].([]map[string]any)
	if !ok {
		for _, e := range Map[[]any](m, "ench") {
			if v, ok := e.(map[string]any); ok {
				enchantments = append(enchantments, v)
			}
		}
	}
	for _, ench := range enchantments {
		if t, ok := item.EnchantmentByID(int(Map[int16](ench, "id"))); ok {
			*s = s.WithEnchantments(item.NewEnchantment(t, int(Map[int16](ench, "lvl"))))
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
