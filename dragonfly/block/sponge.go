package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Sponge is a block that can be used to remove water around itself when placed, turning into a wet sponge in the
// process.
type Sponge struct {
	// Wet specifies whether the dry or the wet variant of the block is used.
	Wet bool
}

// BreakInfo ...
func (s Sponge) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.6,
		Drops: simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sponge) EncodeItem() (id int32, meta int16) {
	if s.Wet {
		meta = 1
	}

	return 19, meta
}

// EncodeBlock ...
func (s Sponge) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Wet {
		return "minecraft:sponge", map[string]interface{}{"sponge_type": "wet"}
	}
	return "minecraft:sponge", map[string]interface{}{"sponge_type": "dry"}
}