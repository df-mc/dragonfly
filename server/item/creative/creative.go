package creative

import (
	"bytes"
	_ "embed"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	// The following three imports are essential for this package: They make sure this package is loaded after
	// all these imports. This ensures that all items are registered before the creative items are registered
	// in the init function in this package.
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// Items returns a list with all items that have been registered as a creative item. These items will
// be accessible by players in-game who have creative mode enabled.
func Items() []item.Stack {
	return creativeItemStacks
}

// RegisterItem registers an item as a creative item, exposing it in the creative inventory.
func RegisterItem(item item.Stack) {
	creativeItemStacks = append(creativeItemStacks, item)
}

var (
	//go:embed creative_items.nbt
	creativeItemData []byte
	// creativeItemStacks holds a list of all item stacks that were registered to the creative inventory using
	// RegisterItem.
	creativeItemStacks []item.Stack
)

// creativeItemEntry holds data of a creative item as present in the creative inventory.
type creativeItemEntry struct {
	Name  string         `nbt:"name"`
	Meta  int16          `nbt:"meta"`
	NBT   map[string]any `nbt:"nbt,omitempty"`
	Block struct {
		Name       string         `nbt:"name"`
		Properties map[string]any `nbt:"states"`
		Version    int32          `nbt:"version"`
	} `nbt:"block,omitempty"`
}

// init initialises the creative items, registering all creative items that have also been registered as
// normal items and are present in vanilla.
func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(creativeItemData))

	// Register all creative items present in the creative_items.nbt file.
	for {
		var data creativeItemEntry
		if err := dec.Decode(&data); err != nil {
			break
		}
		var (
			it world.Item
			ok bool
		)
		if data.Block.Version > 0 {
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
			if len(data.NBT) > 0 {
				it = n.DecodeNBT(data.NBT).(world.Item)
			}
		}

		st := item.NewStack(it, 1)
		if len(data.NBT) > 0 {
			var invalid bool
			for _, e := range nbtconv.Map[[]any](data.NBT, "ench") {
				if v, ok := e.(map[string]any); ok {
					t, ok := item.EnchantmentByID(int(nbtconv.Map[int16](v, "id")))
					if !ok {
						invalid = true
						break
					}
					st = st.WithEnchantments(item.NewEnchantment(t, int(nbtconv.Map[int16](v, "lvl"))))
				}
			}
			if invalid {
				// Invalid enchantment, skip this item.
				continue
			}
		}
		RegisterItem(st)
	}
}
