package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// EnchantingTable is a block that allows players to spend their experience point levels to enchant tools, weapons,
// books, armor, and certain other items.
type EnchantingTable struct {
	transparent
	bassDrum
}

// Model ...
func (e EnchantingTable) Model() world.BlockModel {
	return model.EnchantingTable{}
}

// BreakInfo ...
func (e EnchantingTable) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(e))
}

// CanDisplace ...
func (EnchantingTable) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (EnchantingTable) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// LightEmissionLevel ...
func (EnchantingTable) LightEmissionLevel() uint8 {
	return 7
}

// Activate ...
func (EnchantingTable) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// EncodeItem ...
func (EnchantingTable) EncodeItem() (name string, meta int16) {
	return "minecraft:enchanting_table", 0
}

// EncodeBlock ...
func (EnchantingTable) EncodeBlock() (string, map[string]any) {
	return "minecraft:enchanting_table", nil
}

// EncodeNBT is used to encode the block to NBT, so that the enchanting table book will be rendered properly client-side.
// The actual rotation value doesn't need to be set in the NBT, we just need to write the default NBT for the block.
func (e EnchantingTable) EncodeNBT() map[string]any {
	return map[string]any{"id": "EnchantTable"}
}

// DecodeNBT is used to implement world.NBTer.
func (e EnchantingTable) DecodeNBT(map[string]any) any {
	return e
}
