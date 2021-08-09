package nbtconv

import (
	"github.com/df-mc/dragonfly/server/item/inventory"
)

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
		data := WriteItem(i, true)
		data["Slot"] = byte(index)
		items = append(items, data)
	}
	return items
}
