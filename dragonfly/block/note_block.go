package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/instrument"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/particle"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
)

// NoteBlock is a musical block that emits sounds when powered with redstone.
type NoteBlock struct {
	nbt
	solid
	bass

	// Pitch is the current pitch the note block is set to. Value ranges from 0-24.
	Pitch int
}

// playNote ...
func (n NoteBlock) playNote(pos world.BlockPos, w *world.World) {
	w.PlaySound(pos.Vec3(), sound.Note{Instrument: n.instrument(pos, w), Pitch: n.Pitch})
	w.AddParticle(pos.Vec3(), particle.Note{Instrument: n.Instrument(), Pitch: n.Pitch})
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
	n.Pitch = int(readByte(data, "note"))
	return n
}

// EncodeNBT ...
func (n NoteBlock) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"note": byte(n.Pitch)}
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
	n.playNote(pos, w)
	w.SetBlock(pos, n)
}

// BreakInfo ...
func (n NoteBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(n, 1)),
	}
}

// EncodeItem ...
func (n NoteBlock) EncodeItem() (id int32, meta int16) {
	return 25, 0
}

// EncodeBlock ...
func (n NoteBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:noteblock", map[string]interface{}{"pitch": n.Pitch}
}

// allNoteBlocks ...
func allNoteBlocks() (noteBlocks []NoteBlock) {
	for i := 0; i < 25; i++ {
		noteBlocks = append(noteBlocks, NoteBlock{Pitch: i})
	}
	return
}
