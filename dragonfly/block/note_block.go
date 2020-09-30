package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/instrument"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/player/chat"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
)

// NoteBlock is a musical block that emits sounds when powered with redstone.
type NoteBlock struct {
	solid
	bass

	// Pitch is the current pitch the note block is set to. Value ranges from 0-24.
	Pitch int32
}

// playNote ...
func (n NoteBlock) playNote(pos world.BlockPos, w *world.World) {
	w.PlaySound(pos.Vec3(), sound.Note{Instrument: n.instrument(pos, w), Pitch: int(n.Pitch)})
	//TODO: Particles
}

// updateInstrument ...
func (n NoteBlock) instrument(pos world.BlockPos, w *world.World) instrument.Instrument {
	if instrumentBlock, ok := w.Block(pos.Side(world.FaceDown)).(InstrumentBlock); ok {
		return instrumentBlock.Instrument()
	}
	return instrument.Piano()
}

// DecodeNBT ...
func (n NoteBlock) DecodeNBT(data map[string]interface{}) interface{} {
	n.Pitch = readInt32(data, "note")
	return n
}

// EncodeNBT ...
func (n NoteBlock) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"note": n.Pitch}
}

// Punch ...
func (n NoteBlock) Punch(pos world.BlockPos, _ world.Face, w *world.World, u item.User) {
	if _, ok := w.Block(pos.Side(world.FaceUp)).(Air); !ok {
		return
	}
	n.playNote(pos, w)
}

// Activate ...
func (n NoteBlock) Activate(pos world.BlockPos, _ world.Face, w *world.World, _ item.User) {
	if _, ok := w.Block(pos.Side(world.FaceUp)).(Air); !ok {
		return
	}
	n.Pitch = (n.Pitch + 1) % 25
	chat.Global.Printf("pitch %v instrument %v", n.Pitch, n.instrument(pos, w).MagicNumber)
	n.playNote(pos, w)
	w.SetBlock(pos, n)
}

// EncodeItem ...
func (n NoteBlock) EncodeItem() (id int32, meta int16) {
	return 25, 0
}

// EncodeBlock ...
func (n NoteBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:noteblock", nil
}

// HasNBT...
func (n NoteBlock) HasNBT() bool {
	return true
}

// Hash ...
func (n NoteBlock) Hash() uint64 {
	return hashNoteblock
}
