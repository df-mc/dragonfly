package item

import (
	_ "embed"
	"encoding/base64"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	_ "unsafe" // Imported for compiler directives.
)

// CreativeItems returns a list with all items that have been registered as a creative item. These items will
// be accessible by players in-game who have creative mode enabled.
func CreativeItems() []Stack {
	return creativeItemStacks
}

// RegisterCreativeItem registers an item as a creative item, exposing it in the creative inventory.
func RegisterCreativeItem(item Stack) {
	creativeItemStacks = append(creativeItemStacks, item)
}

var (
	//go:embed creative_items.nbt
	creativeItemData []byte
	// creativeItemStacks holds a list of all item stacks that were registered to the creative inventory using
	// RegisterCreativeItem.
	creativeItemStacks []Stack
)

//lint:ignore U1000 Type is used using compiler directives.
type creativeItemEntry struct {
	Name string `json:"name" nbt:"name"`
	Meta int16  `json:"meta" nbt:"meta"`
	NBT  string `json:"nbt" nbt:"nbt"`
}

// registerVanillaCreativeItems initialises the creative items, registering all creative items that have also
// been registered as normal items and are present in vanilla.
// TODO: Call this without awful linkname directives. It's currently done this way because the block package is
//  loaded before the item package, so registering vanilla items here will lead to none of the blocks being in
//  the inventory.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func registerVanillaCreativeItems() {
	var temp map[string]interface{}

	var m []creativeItemEntry
	if err := nbt.Unmarshal(creativeItemData, &m); err != nil {
		panic(err)
	}
	for _, data := range m {
		it, found := world.ItemByName(data.Name, data.Meta)
		if !found {
			// The item wasn't registered, so don't register it as a creative item.
			continue
		}
		_, _, resultingMeta := it.EncodeItem()
		if resultingMeta != data.Meta {
			// We found an item registered with that ID and a meta of 0, but we only need items with strictly
			// the same meta here.
			continue
		}
		if n, ok := it.(world.NBTer); ok {
			nbtData, _ := base64.StdEncoding.DecodeString(data.NBT)
			if err := nbt.Unmarshal(nbtData, &temp); err != nil {
				panic(err)
			}
			if len(temp) != 0 {
				it = n.DecodeNBT(temp).(world.Item)
			}
		}
		RegisterCreativeItem(NewStack(it, 1))
	}
}
