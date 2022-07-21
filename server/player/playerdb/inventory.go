package playerdb

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func invToData(data player.InventoryData) jsonInventoryData {
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

func dataToInv(data jsonInventoryData) player.InventoryData {
	d := player.InventoryData{
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
	return nbtconv.ReadItem(itemNBT, nil)
}
