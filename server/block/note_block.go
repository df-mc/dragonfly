package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// NoteBlock is a musical block that emits sounds when powered with redstone.
type NoteBlock struct {
	solid
	bass

	// Pitch is the current pitch the note block is set to. Value ranges from 0-24.
	Pitch int
}

// playNote ...
func (n NoteBlock) playNote(pos cube.Pos, w *world.World) {
	w.PlaySound(pos.Vec3(), sound.Note{Instrument: n.instrument(pos, w), Pitch: n.Pitch})
	w.AddParticle(pos.Vec3(), particle.Note{Instrument: n.Instrument(), Pitch: n.Pitch})
}

// updateInstrument ...
func (n NoteBlock) instrument(pos cube.Pos, w *world.World) instrument.Instrument {
	if instrumentBlock, ok := w.Block(pos.Side(cube.FaceDown)).(InstrumentBlock); ok {
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
func (n NoteBlock) Punch(pos cube.Pos, _ cube.Face, w *world.World, u item.User) {
	if _, ok := w.Block(pos.Side(cube.FaceUp)).(Air); !ok {
		return
	}
	n.playNote(pos, w)
}

// Activate ...
func (n NoteBlock) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
	if _, ok := w.Block(pos.Side(cube.FaceUp)).(Air); !ok {
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
func (n NoteBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:noteblock", 0
}

// EncodeBlock ...
func (n NoteBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:noteblock", nil
}
