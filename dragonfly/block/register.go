package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/colour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
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
	world.RegisterBlock(Cobblestone{}, Cobblestone{Mossy: true})
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
	world.RegisterItem("minecraft:cobblestone", Cobblestone{})
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
	world.RegisterItem("minecraft:mossy_cobblestone", Cobblestone{Mossy: true})
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
	for _, c := range colour.All() {
		world.RegisterItem("minecraft:concrete", Concrete{Colour: c})
	}
}

// readSlice reads an interface slice from a map at the key passed.
func readSlice(m map[string]interface{}, key string) []interface{} {
	v, _ := m[key]
	b, _ := v.([]interface{})
	return b
}

// readMap reads an interface map from a map at the key passed.
func readMap(m map[string]interface{}, key string) map[string]interface{} {
	v, _ := m[key]
	b, _ := v.(map[string]interface{})
	if b == nil {
		b = map[string]interface{}{}
	}
	return b
}

// readString reads a string from a map at the key passed.
func readString(m map[string]interface{}, key string) string {
	v, _ := m[key]
	b, _ := v.(string)
	return b
}
