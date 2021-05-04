package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Shroomlight are light-emitting blocks that generate in huge fungi.
type Shroomlight struct {
	solid
}

// LightEmissionLevel ...
func (Shroomlight) LightEmissionLevel() uint8 {
	return 15
}

// BreakInfo ...
func (s Shroomlight) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1,
		Harvestable: alwaysHarvestable,
		Effective:   hoeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (Shroomlight) EncodeItem() (name string, meta int16) {
	return "minecraft:shroomlight", 0
}

// EncodeBlock ...
func (Shroomlight) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:shroomlight", nil
}
