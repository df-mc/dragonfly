package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// RootedDirt is a variant of dirt that generates in lush caves underneath azalea trees.
type RootedDirt struct {
	solid
}

// SoilFor ...
func (r RootedDirt) SoilFor(block world.Block) bool {
	switch block.(type) {
	case ShortGrass, Fern, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts, PinkPetals, SugarCane, DeadBush, BambooSapling, Bamboo:
		return true
	}
	return false
}

// BoneMeal ...
func (r RootedDirt) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Air); !ok {
		return item.BoneMealResultNone
	}
	tx.SetBlock(below, HangingRoots{}, nil)
	return item.BoneMealResultSmall
}

// Till returns dirt, as using a hoe on rooted dirt strips its roots instead of tilling it into farmland.
func (RootedDirt) Till() (world.Block, bool) {
	return Dirt{}, true
}

// TillDrops ...
func (RootedDirt) TillDrops() []item.Stack {
	return []item.Stack{item.NewStack(HangingRoots{}, 1)}
}

// Shovel ...
func (RootedDirt) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}

// BreakInfo ...
func (r RootedDirt) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(r))
}

// EncodeItem ...
func (RootedDirt) EncodeItem() (name string, meta int16) {
	return "minecraft:dirt_with_roots", 0
}

// EncodeBlock ...
func (RootedDirt) EncodeBlock() (string, map[string]any) {
	return "minecraft:dirt_with_roots", nil
}
