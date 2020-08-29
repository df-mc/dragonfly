package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Shroomlights are light-emitting blocks that generate in huge fungi.
type Shroomlight struct {
	noNBT
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
func (Shroomlight) EncodeItem() (id int32, meta int16) {
	return -230, 0
}

// EncodeBlock ...
func (Shroomlight) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:shroomlight", nil
}

// Hash ...
func (Shroomlight) Hash() uint64 {
	return hashShroomlight
}
