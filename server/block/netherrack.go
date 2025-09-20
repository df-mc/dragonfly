package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Netherrack is a block found in The Nether.
type Netherrack struct {
	solid
	bassDrum
}

func (n Netherrack) SoilFor(block world.Block) bool {
	flower, ok := block.(Flower)
	return ok && flower.Type == WitherRose()
}

func (n Netherrack) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, pickaxeHarvestable, pickaxeEffective, oneOf(n))
}

func (Netherrack) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(item.NetherBrick{}, 1), 0.1)
}

func (Netherrack) EncodeItem() (name string, meta int16) {
	return "minecraft:netherrack", 0
}

func (Netherrack) EncodeBlock() (string, map[string]any) {
	return "minecraft:netherrack", nil
}
