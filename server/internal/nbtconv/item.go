package nbtconv

import (
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// InvFromNBT decodes the data of an NBT slice into the inventory passed.
func InvFromNBT(inv *inventory.Inventory, items []any) {
	for _, itemData := range items {
		data, _ := itemData.(map[string]any)
		it := Item(data, nil)
		if it.Empty() {
			continue
		}
		_ = inv.SetItem(int(Uint8(data, "Slot")), it)
	}
}

// InvToNBT encodes an inventory to a data slice which may be encoded as NBT.
func InvToNBT(inv *inventory.Inventory) []map[string]any {
	var items []map[string]any
	for index, i := range inv.Slots() {
		if i.Empty() {
			continue
		}
		data := WriteItem(i, true)
		data["Slot"] = byte(index)
		items = append(items, data)
	}
	return items
}
