package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

type (
	// Stone is a block found underground in the world or on mountains.
	Stone struct {
		noNBT
		solid
	}

	// Granite is a type of igneous rock.
	Granite polishable
	// Diorite is a type of igneous rock.
	Diorite polishable
	// Andesite is a type of igneous rock.
	Andesite polishable

	// polishable forms the base of blocks that may be polished.
	polishable struct {
		noNBT
		solid
		// Polished specifies if the block is polished or not. When set to true, the block will represent its
		// polished variant, for example polished andesite.
		Polished bool
	}
)

var stoneBreakInfo = BreakInfo{
	Hardness:    1.5,
	Harvestable: pickaxeHarvestable,
	Effective:   pickaxeEffective,
	Drops:       simpleDrops(item.NewStack(Cobblestone{}, 1)),
}

// BreakInfo ...
func (s Stone) BreakInfo() BreakInfo {
	return stoneBreakInfo
}

// BreakInfo ...
func (g Granite) BreakInfo() BreakInfo {
	i := stoneBreakInfo
	i.Drops = simpleDrops(item.NewStack(g, 1))
	return i
}

// BreakInfo ...
func (d Diorite) BreakInfo() BreakInfo {
	i := stoneBreakInfo
	i.Drops = simpleDrops(item.NewStack(d, 1))
	return i
}

// BreakInfo ...
func (a Andesite) BreakInfo() BreakInfo {
	i := stoneBreakInfo
	i.Drops = simpleDrops(item.NewStack(a, 1))
	return i
}

// EncodeItem ...
func (s Stone) EncodeItem() (id int32, meta int16) {
	return 1, 0
}

// EncodeItem ...
func (a Andesite) EncodeItem() (id int32, meta int16) {
	if a.Polished {
		return 1, 6
	}
	return 1, 5
}

// EncodeItem ...
func (d Diorite) EncodeItem() (id int32, meta int16) {
	if d.Polished {
		return 1, 4
	}
	return 1, 3
}

// EncodeItem ...
func (g Granite) EncodeItem() (id int32, meta int16) {
	if g.Polished {
		return 1, 2
	}
	return 1, 1
}

// EncodeBlock ...
func (Stone) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:stone", map[string]interface{}{"stone_type": "stone"}
}

// Hash ...
func (Stone) Hash() uint64 {
	return hashStone
}

// EncodeBlock ...
func (a Andesite) EncodeBlock() (name string, properties map[string]interface{}) {
	if a.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "andesite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "andesite"}
}

// Hash ...
func (a Andesite) Hash() uint64 {
	return hashAndesite | (uint64(boolByte(a.Polished)) << 32)
}

// EncodeBlock ...
func (d Diorite) EncodeBlock() (name string, properties map[string]interface{}) {
	if d.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "diorite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "diorite"}
}

// Hash ...
func (d Diorite) Hash() uint64 {
	return hashDiorite | (uint64(boolByte(d.Polished)) << 32)
}

// EncodeBlock ...
func (g Granite) EncodeBlock() (name string, properties map[string]interface{}) {
	if g.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "granite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "granite"}
}

// Hash ...
func (g Granite) Hash() uint64 {
	return hashGranite | (uint64(boolByte(g.Polished)) << 32)
}
