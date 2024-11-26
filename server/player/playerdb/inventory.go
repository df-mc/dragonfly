package playerdb

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// InventoryData is a struct that contains all data of the player inventories.
type InventoryData struct {
	// Items contains all the items in the player's main inventory.
	// This excludes armor and offhand.
	Items []item.Stack
	// Boots, Leggings, Chestplate, Helmet are armor pieces that belong to the slot corresponding to the name.
	Boots      item.Stack
	Leggings   item.Stack
	Chestplate item.Stack
	Helmet     item.Stack
	// OffHand is what the player is carrying in their non-main hand, like a shield or arrows.
	OffHand item.Stack
	// MainHandSlot saves the slot in the hotbar that the player is currently switched to.
	// Should be between 0-8.
	MainHandSlot uint32
}

func invToData(data InventoryData) jsonInventoryData {
	d := jsonInventoryData{
		MainHandSlot: data.MainHandSlot,
		OffHand:      encodeItem(data.OffHand),
	}
	d.Items = encodeItems(data.Items)
	d.Boots = encodeItem(data.Boots)
	d.Leggings = encodeItem(data.Leggings)
	d.Chestplate = encodeItem(data.Chestplate)
	d.Helmet = encodeItem(data.Helmet)
	return d
}

func dataToInv(data jsonInventoryData) InventoryData {
	d := InventoryData{
		MainHandSlot: data.MainHandSlot,
		OffHand:      decodeItem(data.OffHand),
		Items:        make([]item.Stack, 36),
	}
	decodeItems(data.Items, d.Items)
	d.Boots = decodeItem(data.Boots)
	d.Leggings = decodeItem(data.Leggings)
	d.Chestplate = decodeItem(data.Chestplate)
	d.Helmet = decodeItem(data.Helmet)
	return d
}

func encodeItems(items []item.Stack) (encoded []jsonSlot) {
	encoded = make([]jsonSlot, 0, len(items))
	for slot, i := range items {
		data := encodeItem(i)
		if data == nil {
			continue
		}
		encoded = append(encoded, jsonSlot{Slot: slot, Item: data})
	}
	return
}

func decodeItems(encoded []jsonSlot, items []item.Stack) {
	for _, i := range encoded {
		items[i.Slot] = decodeItem(i.Item)
	}
}

func encodeItem(item item.Stack) []byte {
	if item.Empty() {
		return nil
	}

	var b bytes.Buffer
	itemNBT := nbtconv.WriteItem(item, true)
	encoder := nbt.NewEncoderWithEncoding(&b, nbt.LittleEndian)
	err := encoder.Encode(itemNBT)
	if err != nil {
		return nil
	}
	return b.Bytes()
}

func decodeItem(data []byte) item.Stack {
	var itemNBT map[string]any
	decoder := nbt.NewDecoderWithEncoding(bytes.NewBuffer(data), nbt.LittleEndian)
	err := decoder.Decode(&itemNBT)
	if err != nil {
		return item.Stack{}
	}
	return nbtconv.Item(itemNBT, nil)
}
