package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Shroomlight are light-emitting blocks that generate in huge fungi.
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
