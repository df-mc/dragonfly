package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/grass"
)

type TallGrass struct {
	noNBT
	solid
	Type grass.TallGrass
}

// EncodeBlock ...
func (g TallGrass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:tallgrass", map[string]interface{}{"tall_grass_type": g.Type.Name()}
}

// BreakInfo ...
func (g TallGrass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(),
	}
}
