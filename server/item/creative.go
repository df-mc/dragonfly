package item

import (
	_ "embed"
	"encoding/base64"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
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

// init initialises the creative items, registering all creative items that have also been registered as
// normal items and are present in vanilla.
func init() {
	registerCreativeItems()
}

// registerCreativeItems registers all creative items stored in the creativeItemData for which an item has previously
// been registered.
func registerCreativeItems() {
	// creativeItemEntry holds data of a creative item as present in the creative inventory.
	type creativeItemEntry struct {
		Name  string `nbt:"name"`
		Meta  int16  `nbt:"meta"`
		NBT   string `nbt:"nbt"`
		Block struct {
			Name       string                 `nbt:"name"`
			Properties map[string]interface{} `nbt:"states"`
			Version    int32                  `nbt:"version"`
		} `nbt:"block"`
	}
	var temp map[string]interface{}

	var m []creativeItemEntry
	if err := nbt.Unmarshal(creativeItemData, &m); err != nil {
		panic(err)
	}
	for _, data := range m {
		var (
			it world.Item
			ok bool
		)
		if data.Block.Version != 0 {
			// Item with a block, try parsing the block, then try asserting that to an item. Blocks no longer
			// have their metadata sent, but we still need to get that metadata in order to be able to register
			// different block states as different items.
			if b, ok := world.BlockByName(data.Block.Name, data.Block.Properties); ok {
				if it, ok = b.(world.Item); !ok {
					continue
				}
			}
		} else {
			if it, ok = world.ItemByName(data.Name, data.Meta); !ok {
				// The item wasn't registered, so don't register it as a creative item.
				continue
			}
			if _, resultingMeta := it.EncodeItem(); resultingMeta != data.Meta {
				// We found an item registered with that ID and a meta of 0, but we only need items with strictly
				// the same meta here.
				continue
			}
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
