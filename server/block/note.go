package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Note is a musical block that emits sounds when powered with redstone.
type Note struct {
	solid
	bass

	// Pitch is the current pitch the note block is set to. Value ranges from 0-24.
	Pitch int
}

// playNote plays the note block's sound and shows a particle.
func (n Note) playNote(pos cube.Pos, tx *world.Tx) {
	tx.PlaySound(pos.Vec3(), sound.Note{Instrument: n.instrument(pos, tx), Pitch: n.Pitch})
	tx.AddParticle(pos.Vec3(), particle.Note{Instrument: n.Instrument(), Pitch: n.Pitch})
}

// instrument determines which instrument sound to use based on the block below.
func (n Note) instrument(pos cube.Pos, tx *world.Tx) sound.Instrument {
	if instrumentBlock, ok := tx.Block(pos.Side(cube.FaceDown)).(interface {
		Instrument() sound.Instrument
	}); ok {
		return instrumentBlock.Instrument()
	}
	return sound.Piano()
}

// Instrument returns the instrument for particle display.
func (n Note) Instrument() sound.Instrument {
	return sound.Piano()
}

// DecodeNBT decodes the note block's pitch from NBT data.
func (n Note) DecodeNBT(data map[string]any) any {
	n.Pitch = int(nbtconv.Uint8(data, "note"))
	return n
}

// EncodeNBT encodes the note block's pitch to NBT data.
func (n Note) EncodeNBT() map[string]any {
	return map[string]any{"note": byte(n.Pitch)}
}

// Activate handles manual interaction with the note block (tuning).
func (n Note) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	n.trigger(pos, tx)
	return true
}

// RedstoneUpdate is called when redstone power changes nearby.
func (n Note) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	if n.powered(pos, tx) {
		n.trigger(pos, tx)
	}
}

// NeighbourUpdateTick is called when a neighboring block updates.
func (n Note) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if n.powered(pos, tx) {
		n.trigger(pos, tx)
	}
}

// trigger plays the note if there's space above the note block.
func (n Note) trigger(pos cube.Pos, tx *world.Tx) {
	// Can only tune if there's air above
	if _, ok := tx.Block(pos.Side(cube.FaceUp)).(Air); !ok {
		return
	}
	n.Pitch = (n.Pitch + 1) % 25
	n.playNote(pos, tx)
	tx.SetBlock(pos, n, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
}

// powered checks if the note block is receiving redstone power.
func (n Note) powered(pos cube.Pos, tx *world.Tx) bool {
	for _, face := range cube.Faces() {
		adjacentPos := pos.Side(face)
		if power := tx.RedstonePower(adjacentPos, face.Opposite(), true); power > 0 {
			return true
		}
		if power := tx.RedstonePower(adjacentPos, face.Opposite(), false); power > 0 {
			return true
		}
	}
	return false
}

// BreakInfo returns information about breaking the note block.
func (n Note) BreakInfo() BreakInfo {
	return newBreakInfo(0.8, alwaysHarvestable, axeEffective, oneOf(Note{}))
}

// FuelInfo returns fuel information for the note block.
func (Note) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EncodeItem encodes the note block as an item.
func (n Note) EncodeItem() (name string, meta int16) {
	return "minecraft:noteblock", 0
}

// EncodeBlock encodes the note block as a block.
func (n Note) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:noteblock", nil
}
