package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Sandstone is a solid block commonly found in deserts and beaches underneath sand.
type Sandstone struct {
	noNBT
	solid
	Red  bool
	Data int16
}

// BreakInfo ...
func (s Sandstone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sandstone) EncodeItem() (id int32, meta int16) {
	if s.Red {
		return 179, s.Data
	} else {
		return 24, s.Data
	}
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (name string, properties map[string]interface{}) {
	name = "minecraft:sandstone"
	if s.Red {
		name = "minecraft:red_sandstone"
	}
	switch s.Data {
	case 0:
		properties = map[string]interface{}{"sand_stone_type": "default"}
	case 1:
		properties = map[string]interface{}{"sand_stone_type": "heiroglyphs"}
	case 2:
		properties = map[string]interface{}{"sand_stone_type": "cut"}
	case 3:
		properties = map[string]interface{}{"sand_stone_type": "smooth"}
	}
	return
}

// Hash ...
func (s Sandstone) Hash() uint64 {
	switch s.Data {
	case 1:
		return hashChiseledSandstone | (uint64(boolByte(s.Red)) << 32)
	case 2:
		return hashCutSandstone | (uint64(boolByte(s.Red)) << 32)
	case 3:
		return hashSmoothSandstone | (uint64(boolByte(s.Red)) << 32)
	}
	return hashSandstone | (uint64(boolByte(s.Red)) << 32)
}

// allSandstone returns all possible sandstone states.
func allSandstone() []world.Block {
	return []world.Block{
		Sandstone{
			Red:  false,
			Data: 0,
		},
		Sandstone{
			Red:  false,
			Data: 1,
		},
		Sandstone{
			Red:  false,
			Data: 2,
		},
		Sandstone{
			Red:  false,
			Data: 3,
		},
		Sandstone{
			Red:  true,
			Data: 0,
		},
		Sandstone{
			Red:  true,
			Data: 1,
		},
		Sandstone{
			Red:  true,
			Data: 2,
		},
		Sandstone{
			Red:  true,
			Data: 3,
		},
	}
}
