package creative

import (
	_ "embed"
	"fmt"

	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	// The following four imports are essential for this package: They make sure this package is loaded after
	// all these imports. This ensures that all blocks and items are registered before the creative items are
	// registered in the init function in this package.
	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// Item represents a registered item in the creative inventory. It holds a stack of the item and a group that
// the item is part of.
type Item struct {
	// Stack is the stack of the item that is registered in the creative inventory.
	Stack item.Stack
	// Group is the name of the group that the item is part of. If two groups are registered with the same
	// name, the item will always reside in the first group that was registered.
	Group string
}

// Group represents a group of items in the creative inventory. Each group has a category, a name and an icon.
// If either the name or icon is empty, the group is considered an 'anonymous group' and will not group its
// contents together in the creative inventory.
type Group struct {
	// Category is the category of the group. It determines the tab in which the group will be displayed in the
	// creative inventory.
	Category Category
	// Name is the localised name of the group, i.e. "itemGroup.name.planks".
	Name string
	// Icon is the item that will be displayed as the icon of the group in the creative inventory.
	Icon item.Stack
}

// Groups returns a list with all groups that have been registered as a creative group. These groups will be
// accessible by players in-game who have creative mode enabled.
func Groups() []Group {
	return creativeGroups
}

// RegisterGroup registers a group as a creative group, exposing it in the creative inventory. It can then
// be referenced using its name when calling RegisterItem.
func RegisterGroup(group Group) {
	creativeGroups = append(creativeGroups, group)
}

// Items returns a list with all items that have been registered as a creative item. These items will
// be accessible by players in-game who have creative mode enabled.
func Items() []Item {
	return creativeItemStacks
}

// RegisterItem registers an item as a creative item, exposing it in the creative inventory.
func RegisterItem(item Item) {
	creativeItemStacks = append(creativeItemStacks, item)
}

var (
	//go:embed creative_items.nbt
	creativeItemData []byte

	// creativeGroups holds a list of all groups that were registered to the creative inventory using
	// RegisterGroup.
	creativeGroups []Group
	// creativeItemStacks holds a list of all item stacks that were registered to the creative inventory using
	// RegisterItem.
	creativeItemStacks []Item
)

// creativeGroupEntry holds data of a creative group as present in the creative inventory.
type creativeGroupEntry struct {
	Category int32             `nbt:"category"`
	Name     string            `nbt:"name"`
	Icon     creativeItemEntry `nbt:"icon"`
}

// creativeItemEntry holds data of a creative item as present in the creative inventory.
type creativeItemEntry struct {
	Name            string         `nbt:"name"`
	Meta            int16          `nbt:"meta"`
	NBT             map[string]any `nbt:"nbt,omitempty"`
	BlockProperties map[string]any `nbt:"block_properties,omitempty"`
	GroupIndex      int32          `nbt:"group_index,omitempty"`
}

// registerCreativeItems initialises the creative items, registering all creative items that have also been registered as
// normal items and are present in vanilla.
// noinspection GoUnusedFunction
//
//lint:ignore U1000 Function is used through compiler directives.
func registerCreativeItems() {
	var m struct {
		Groups []creativeGroupEntry `nbt:"groups"`
		Items  []creativeItemEntry  `nbt:"items"`
	}
	if err := nbt.Unmarshal(creativeItemData, &m); err != nil {
		panic(err)
	}
	for i, group := range m.Groups {
		name := group.Name
		if name == "" {
			name = fmt.Sprint("anon", i)
		}
		st, _ := itemStackFromEntry(group.Icon)
		c := Category{category(group.Category)}
		RegisterGroup(Group{Category: c, Name: name, Icon: st})
	}
	for _, data := range m.Items {
		if data.GroupIndex >= int32(len(creativeGroups)) {
			panic(fmt.Errorf("invalid group index %v for item %v", data.GroupIndex, data.Name))
		}
		st, ok := itemStackFromEntry(data)
		if !ok {
			continue
		}
		RegisterItem(Item{st, creativeGroups[data.GroupIndex].Name})
	}
}

func itemStackFromEntry(data creativeItemEntry) (item.Stack, bool) {
	var (
		it world.Item
		ok bool
	)
	if len(data.BlockProperties) > 0 {
		// Item with a block, try parsing the block, then try asserting that to an item. Blocks no longer
		// have their metadata sent, but we still need to get that metadata in order to be able to register
		// different block states as different items.
		if b, ok := world.BlockByName(data.Name, data.BlockProperties); ok {
			if it, ok = b.(world.Item); !ok {
				return item.Stack{}, false
			}
		}
	} else {
		if it, ok = world.ItemByName(data.Name, data.Meta); !ok {
			// The item wasn't registered, so don't register it as a creative item.
			return item.Stack{}, false
		}
		if _, resultingMeta := it.EncodeItem(); resultingMeta != data.Meta {
			// We found an item registered with that ID and a meta of 0, but we only need items with strictly
			// the same meta here.
			return item.Stack{}, false
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
		for _, e := range nbtconv.Slice(data.NBT, "ench") {
			if v, ok := e.(map[string]any); ok {
				t, ok := item.EnchantmentByID(int(nbtconv.Int16(v, "id")))
				if !ok {
					invalid = true
					break
				}
				st = st.WithEnchantments(item.NewEnchantment(t, int(nbtconv.Int16(v, "lvl"))))
			}
		}
		if invalid {
			// Invalid enchantment, skip this item.
			return item.Stack{}, false
		}
	}
	return st, true
}
