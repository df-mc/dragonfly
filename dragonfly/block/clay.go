package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Clay is a block that can be found underwater.
type Clay struct {
	noNBT
	solid
}

// BreakInfo ...
func (c Clay) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(item.ClayBall{}, 4)), //TODO: Drops itself if mined with silk touch
	}
}

// EncodeItem ...
func (c Clay) EncodeItem() (id int32, meta int16) {
	return 82, 0
}

// EncodeBlock ...
func (c Clay) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:clay", nil
}

// Hash ...
func (c Clay) Hash() uint64 {
	return hashClay
}
