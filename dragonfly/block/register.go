package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	_ "unsafe" // Imported for compiler directives.
)

// init registers all blocks implemented by Dragonfly.
func init() {
	world.RegisterBlock(Air{})
	world.RegisterBlock(Stone{})
	world.RegisterBlock(Granite{}, Granite{Polished: true})
	world.RegisterBlock(Diorite{}, Diorite{Polished: true})
	world.RegisterBlock(Andesite{}, Andesite{Polished: true})
	world.RegisterBlock(Grass{})
	world.RegisterBlock(Dirt{}, Dirt{Coarse: true})
	world.RegisterBlock(allLogs()...)
	world.RegisterBlock(allLeaves()...)
	world.RegisterBlock(Bedrock{}, Bedrock{InfiniteBurning: true})
	world.RegisterBlock(Chest{Facing: world.Down}, Chest{Facing: world.Up}, Chest{Facing: world.East}, Chest{Facing: world.West}, Chest{Facing: world.North}, Chest{Facing: world.South})
	world.RegisterBlock(allConcrete()...)
}

func init() {
	world.RegisterItem("minecraft:air", Air{})
	world.RegisterItem("minecraft:stone", Stone{})
	world.RegisterItem("minecraft:stone", Granite{})
	world.RegisterItem("minecraft:stone", Granite{Polished: true})
	world.RegisterItem("minecraft:stone", Diorite{})
	world.RegisterItem("minecraft:stone", Diorite{Polished: true})
	world.RegisterItem("minecraft:stone", Andesite{})
	world.RegisterItem("minecraft:stone", Andesite{Polished: true})
	world.RegisterItem("minecraft:grass", Grass{})
	world.RegisterItem("minecraft:dirt", Dirt{})
	world.RegisterItem("minecraft:dirt", Dirt{Coarse: true})
	world.RegisterItem("minecraft:bedrock", Bedrock{})
	world.RegisterItem("minecraft:log", Log{Wood: material.OakWood()})
	world.RegisterItem("minecraft:log", Log{Wood: material.SpruceWood()})
	world.RegisterItem("minecraft:log", Log{Wood: material.BirchWood()})
	world.RegisterItem("minecraft:log", Log{Wood: material.JungleWood()})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: material.OakWood()})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: material.SpruceWood()})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: material.BirchWood()})
	world.RegisterItem("minecraft:leaves", Leaves{Wood: material.JungleWood()})
	world.RegisterItem("minecraft:chest", Chest{})
	world.RegisterItem("minecraft:leaves2", Leaves{Wood: material.AcaciaWood()})
	world.RegisterItem("minecraft:leaves2", Leaves{Wood: material.DarkOakWood()})
	world.RegisterItem("minecraft:log2", Log{Wood: material.AcaciaWood()})
	world.RegisterItem("minecraft:log2", Log{Wood: material.DarkOakWood()})
	world.RegisterItem("minecraft:stripped_spruce_log", Log{Wood: material.SpruceWood(), Stripped: true})
	world.RegisterItem("minecraft:stripped_birch_log", Log{Wood: material.BirchWood(), Stripped: true})
	world.RegisterItem("minecraft:stripped_jungle_log", Log{Wood: material.JungleWood(), Stripped: true})
	world.RegisterItem("minecraft:stripped_acacia_log", Log{Wood: material.AcaciaWood(), Stripped: true})
	world.RegisterItem("minecraft:stripped_dark_oak_log", Log{Wood: material.DarkOakWood(), Stripped: true})
	world.RegisterItem("minecraft:stripped_oak_log", Log{Wood: material.OakWood(), Stripped: true})
	for _, c := range allConcrete() {
		world.RegisterItem("minecraft:concrete", c.(Concrete))
	}
}

//go:linkname world_itemByName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.itemByName
//noinspection ALL
func world_itemByName(name string, meta int16) (world.Item, bool)

//go:linkname world_itemToName git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world.itemToName
//noinspection ALL
func world_itemToName(it world.Item) (name string, meta int16)

// itemFromNBT decodes the data of an item into an item stack.
func itemFromNBT(data map[string]interface{}) item.Stack {
	it, ok := world_itemByName(readString(data, "Name"), readInt16(data, "Damage"))
	if !ok {
		it = Air{}
	}
	if nbt, ok := it.(world.NBTer); ok {
		it = nbt.DecodeNBT(data).(world.Item)
	}
	stack := item.NewStack(it, int(readByte(data, "Count")))
	return stack
}

// itemToNBT encodes an item stack to its NBT representation.
func itemToNBT(s item.Stack) map[string]interface{} {
	m := make(map[string]interface{})
	if nbt, ok := s.Item().(world.NBTer); ok {
		m = nbt.EncodeNBT()
	}
	m["Name"], m["Damage"] = world_itemToName(s.Item())
	m["Count"] = byte(s.Count())
	return m
}

// invFromNBT decodes the data of an NBT slice into the inventory passed.
func invFromNBT(inv *inventory.Inventory, items []interface{}) {
	for _, itemData := range items {
		data, _ := itemData.(map[string]interface{})
		it := itemFromNBT(data)
		if it.Empty() {
			continue
		}
		_ = inv.SetItem(int(readByte(data, "Slot")), it)
	}
}

// invToNBT encodes an inventory to a data slice which may be encoded as NBT.
func invToNBT(inv *inventory.Inventory) []map[string]interface{} {
	var items []map[string]interface{}
	for index, i := range inv.All() {
		if i.Empty() {
			continue
		}
		data := itemToNBT(i)
		data["Slot"] = byte(index)
		items = append(items, data)
	}
	return items
}

// readByte reads a byte from a map at the key passed.
func readByte(m map[string]interface{}, key string) byte {
	v, _ := m[key]
	b, _ := v.(byte)
	return b
}

// readInt16 reads an int16 from a map at the key passed.
func readInt16(m map[string]interface{}, key string) int16 {
	v, _ := m[key]
	b, _ := v.(int16)
	return b
}

// readString reads a string from a map at the key passed.
func readString(m map[string]interface{}, key string) string {
	v, _ := m[key]
	b, _ := v.(string)
	return b
}

// readSlice reads an interface slice from a map at the key passed.
func readSlice(m map[string]interface{}, key string) []interface{} {
	v, _ := m[key]
	b, _ := v.([]interface{})
	return b
}
