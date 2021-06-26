package provider

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

func invToData(data player.InventoryData) jsonInventoryData {
	d := jsonInventoryData{
		MainHand: data.MainHand,
		OffHand:  itemToData(data.OffHand),
	}
	for slot, i := range data.Items {
		itemData := itemToData(i)
		if itemData == nil {
			continue
		}
		itemData["slot"] = slot
		d.Items = append(d.Items, itemData)
	}
	d.Boots = itemToData(data.Boots)
	d.Leggings = itemToData(data.Leggings)
	d.Chestplate = itemToData(data.Chestplate)
	d.Helmet = itemToData(data.Helmet)
	return d
}

func dataToInv(data jsonInventoryData) player.InventoryData {
	d := player.InventoryData{
		MainHand: data.MainHand,
		OffHand:  dataToItem(data.OffHand),
		Items:    make([]item.Stack, 36),
	}
	for _, i := range data.Items {
		slot, ok := readInt("slot", i)
		if !ok {
			continue
		}
		d.Items[slot] = dataToItem(i)
	}
	d.Boots = dataToItem(data.Boots)
	d.Leggings = dataToItem(data.Leggings)
	d.Chestplate = dataToItem(data.Chestplate)
	d.Helmet = dataToItem(data.Helmet)
	return d
}

func itemToData(stack item.Stack) map[string]interface{} {
	data := make(map[string]interface{})
	if stack.Empty() {
		return nil
	}

	if b, ok := stack.Item().(world.Block); ok {
		data["block"], data["block_properties"] = b.EncodeBlock()
	} else {
		data["item"], data["item_meta"] = stack.Item().EncodeItem()
	}

	data["count"] = stack.Count()
	if len(stack.Values()) > 0 {
		data["dragonflyData"] = stack.Values()
	}
	if stack.CustomName() != "" {
		data["customname"] = stack.CustomName()
	}
	if len(stack.Lore()) > 0 {
		data["lore"] = stack.Lore()
	}
	if len(stack.Enchantments()) != 0 {
		enchantments := make(map[int]int)
		for _, ench := range stack.Enchantments() {
			if enchType, ok := item.EnchantmentID(ench); ok {
				enchantments[enchType] = ench.Level()
			}
		}
		data["ench"] = enchantments
	}

	return data
}

func dataToItem(data map[string]interface{}) item.Stack {
	if data == nil {
		return item.Stack{}
	}
	var i world.Item
	if name, ok := data["block"].(string); ok {
		properties, ok := data["block_properties"].(map[string]interface{})
		if !ok {
			properties = make(map[string]interface{})
		}
		// parseInts is used here since all numeric values in the unmarshalled data
		// are float64 by default.
		b, ok := world.BlockByName(name, parseInts(properties))
		if !ok {
			return item.Stack{}
		}
		i, ok = b.(world.Item)
		if !ok {
			return item.Stack{}
		}
	} else if name, ok := data["item"].(string); ok {
		meta, ok := data["item"].(int16)
		if !ok {
			meta = 0
		}
		i, ok = world.ItemByName(name, meta)
		if !ok {
			return item.Stack{}
		}
	} else {
		return item.Stack{}
	}

	count, ok := readInt("count", data)
	if !ok {
		count = 1
	}
	stack := item.NewStack(i, count)

	if customname, ok := data["customname"].(string); ok {
		stack = stack.WithCustomName(customname)
	}
	if lore, ok := data["customname"].([]string); ok {
		stack = stack.WithLore(lore...)
	}
	if values, ok := data["dragonflyData"].(map[string]interface{}); ok {
		for key, value := range values {
			stack = stack.WithValue(key, value)
		}
	}
	if enchants, ok := data["ench"].(map[int]int); ok {
		for id, lvl := range enchants {
			enchant, ok := item.EnchantmentByID(id)
			if !ok {
				continue
			}
			stack = stack.WithEnchantment(enchant.WithLevel(lvl))
		}
	}

	return stack
}

// readInt checks if the given value in the map can be converted to an int and returns it.
func readInt(key string, data map[string]interface{}) (int, bool) {
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

// parseInts is used to convert all floating point values in unmarshalled data to int32.
func parseInts(data map[string]interface{}) map[string]interface{} {
	for key, v := range data {
		f, ok := v.(float64)
		if !ok {
			continue
		}
		data[key] = int32(f)
	}
	return data
}
