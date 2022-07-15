package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// StainedTerracotta is a block formed from clay, with a hardness and blast resistance comparable to stone. In contrast
// to Terracotta, t can be coloured in the same 16 colours that wool can be dyed, but more dulled and earthen.
type StainedTerracotta struct {
	solid
	bassDrum

	// Colour specifies the colour of the block.
	Colour item.Colour
}

// SoilFor ...
func (t StainedTerracotta) SoilFor(block world.Block) bool {
	_, ok := block.(DeadBush)
	return ok
}

// BreakInfo ...
func (t StainedTerracotta) BreakInfo() BreakInfo {
	return newBreakInfo(1.25, pickaxeHarvestable, pickaxeEffective, oneOf(t))
}

// EncodeItem ...
func (t StainedTerracotta) EncodeItem() (name string, meta int16) {
	return "minecraft:stained_hardened_clay", int16(t.Colour.Uint8())
}

// EncodeBlock ...
func (t StainedTerracotta) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:stained_hardened_clay", map[string]any{"color": t.Colour.String()}
}

// allStainedTerracotta returns stained terracotta blocks with all possible colours.
func allStainedTerracotta() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range item.Colours() {
		b = append(b, StainedTerracotta{Colour: c})
	}
	return b
}
