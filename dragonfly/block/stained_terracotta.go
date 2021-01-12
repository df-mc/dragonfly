package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// StainedTerracotta is a block formed from clay, with a hardness and blast resistance comparable to stone. In contrast
// to Terracotta, t can be coloured in the same 16 colours that wool can be dyed, but more dulled and earthen.
type StainedTerracotta struct {
	noNBT
	solid
	bassDrum

	// Colour specifies the colour of the block.
	Colour colour.Colour
}

// BreakInfo ...
func (t StainedTerracotta) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.25,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(t, 1)),
	}
}

// EncodeItem ...
func (t StainedTerracotta) EncodeItem() (id int32, meta int16) {
	return 159, int16(t.Colour.Uint8())
}

// EncodeBlock ...
func (t StainedTerracotta) EncodeBlock() (name string, properties map[string]interface{}) {
	colourName := t.Colour.String()
	if t.Colour == colour.LightGrey() {
		// Light grey is actually called "silver" in the block state. Mojang pls.
		colourName = "silver"
	}
	return "minecraft:stained_hardened_clay", map[string]interface{}{"color": colourName}
}

// Hash ...
func (t StainedTerracotta) Hash() uint64 {
	return hashStainedTerracotta | (uint64(t.Colour.Uint8()) << 32)
}

// allStainedTerracotta returns stained terracotta blocks with all possible colours.
func allStainedTerracotta() []StainedTerracotta {
	b := make([]StainedTerracotta, 0, 16)
	for _, c := range colour.All() {
		b = append(b, StainedTerracotta{Colour: c})
	}
	return b
}
