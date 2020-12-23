package item

import (
	"encoding/base64"
	"encoding/json"
	"github.com/df-mc/dragonfly/dragonfly/internal/resource"
	"github.com/df-mc/dragonfly/dragonfly/world"
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

// creativeItemStacks holds a list of all item stacks that were registered to the creative inventory using
// RegisterCreativeItem.
var creativeItemStacks []Stack

//lint:ignore U1000 Type is used using compiler directives.
type creativeItemEntry struct {
	ID   int32
	Meta int16
	NBT  string
}

// registerVanillaCreativeItems initialises the creative items, registering all creative items that have also
// been registered as normal items and are present in vanilla.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func registerVanillaCreativeItems() {
	var temp map[string]interface{}

	var m []creativeItemEntry
	if err := json.Unmarshal([]byte(resource.CreativeItems), &m); err != nil {
		panic(err)
	}
	for _, data := range m {
		it, found := world_itemByID(world_runtimeById(data.ID, data.Meta), data.Meta)
		if !found {
			// The item wasn't registered, so don't register it as a creative item.
			continue
		}
		_, resultingMeta := it.EncodeItem()
		if resultingMeta != data.Meta {
			// We found an item registered with that ID and a meta of 0, but we only need items with strictly
			// the same meta here.
			continue
		}
		//noinspection ALL
		if nbter, ok := it.(world.NBTer); ok {
			nbtData, _ := base64.StdEncoding.DecodeString(data.NBT)
			if err := nbt.Unmarshal(nbtData, &temp); err != nil {
				panic(err)
			}
			if len(temp) != 0 {
				it = nbter.DecodeNBT(temp).(world.Item)
			}
		}
		RegisterCreativeItem(NewStack(it, 1))
	}
}

//go:linkname world_itemByID github.com/df-mc/dragonfly/dragonfly/world.itemByID
//noinspection ALL
func world_itemByID(id int32, meta int16) (world.Item, bool)

//go:linkname world_runtimeById github.com/df-mc/dragonfly/dragonfly/world.runtimeById
//noinspection ALL
func world_runtimeById(id int32, meta int16) int32
