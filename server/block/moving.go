package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
)

// Moving ...
type Moving struct {
	empty
	transparent

	// Moving represents the block that is moving.
	Moving world.Block
	// Extra represents an extra block that is moving with the main block.
	Extra world.Block
	// Piston is the position of the piston that is moving the block.
	Piston cube.Pos
	// Expanding is true if the moving block is expanding, false if it is contracting.
	Expanding bool
}

// EncodeBlock ...
func (Moving) EncodeBlock() (string, map[string]any) {
	return "minecraft:moving_block", nil
}

// PistonImmovable ...
func (Moving) PistonImmovable() bool {
	return true
}

// EncodeNBT ...
func (b Moving) EncodeNBT() map[string]any {
	if b.Moving == nil {
		b.Moving = Air{}
	}
	if b.Extra == nil {
		b.Extra = Air{}
	}
	data := map[string]any{
		"id":               "MovingBlock",
		"expanding":        b.Expanding,
		"movingBlock":      nbtconv.WriteBlock(b.Moving),
		"movingBlockExtra": nbtconv.WriteBlock(b.Extra),
		"pistonPosX":       int32(b.Piston.X()),
		"pistonPosY":       int32(b.Piston.Y()),
		"pistonPosZ":       int32(b.Piston.Z()),
	}
	if nbt, ok := b.Moving.(world.NBTer); ok {
		data["movingEntity"] = nbt.EncodeNBT()
	}
	return data
}

// DecodeNBT ...
func (b Moving) DecodeNBT(m map[string]any) any {
	b.Expanding = nbtconv.Bool(m, "expanding")
	b.Moving = nbtconv.Block(m, "movingBlock")
	b.Extra = nbtconv.Block(m, "movingBlockExtra")
	b.Piston = cube.Pos{
		int(nbtconv.Int32(m, "pistonPosX")),
		int(nbtconv.Int32(m, "pistonPosY")),
		int(nbtconv.Int32(m, "pistonPosZ")),
	}
	if nbt, ok := b.Moving.(world.NBTer); ok {
		b.Moving = nbt.DecodeNBT(m["movingEntity"].(map[string]any)).(world.Block)
	}
	return b
}
