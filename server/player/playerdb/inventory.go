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
	for slot, i := range data.Items {
		itemData := encodeItem(i)
		if itemData == nil {
			continue
		}
		d.Items = append(d.Items, jsonSlot{
			Slot: slot,
			Item: itemData,
		})
	}
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
	for _, i := range data.Items {
		d.Items[i.Slot] = decodeItem(i.Item)
	}
	d.Boots = decodeItem(data.Boots)
	d.Leggings = decodeItem(data.Leggings)
	d.Chestplate = decodeItem(data.Chestplate)
	d.Helmet = decodeItem(data.Helmet)
	return d
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
