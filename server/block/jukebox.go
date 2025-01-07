package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
)

// Jukebox is a block used to play music discs.
type Jukebox struct {
	solid
	bass

	// Item is the music disc played by the jukebox.
	Item item.Stack
}

// InsertItem ...
func (j Jukebox) InsertItem(h Hopper, pos cube.Pos, tx *world.Tx) bool {
	if !j.Item.Empty() {
		return false
	}

	for sourceSlot, sourceStack := range h.inventory.Slots() {
		if sourceStack.Empty() {
			continue
		}

		if m, ok := sourceStack.Item().(item.MusicDisc); ok {
			j.Item = sourceStack
			tx.SetBlock(pos, j, nil)
			_ = h.inventory.SetItem(sourceSlot, sourceStack.Grow(-1))
			tx.PlaySound(pos.Vec3Centre(), sound.MusicDiscPlay{DiscType: m.DiscType})
			return true
		}
	}

	return false
}

// ExtractItem ...
func (j Jukebox) ExtractItem(_ Hopper, _ cube.Pos, _ *world.Tx) bool {
	// TODO: This functionality requires redstone to be implemented.
	return false
}

// FuelInfo ...
func (j Jukebox) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// BreakInfo ...
func (j Jukebox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(Jukebox{})).withBlastResistance(30).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if _, hasDisc := j.Disc(); hasDisc {
			dropItem(tx, j.Item, pos.Vec3())
			tx.PlaySound(pos.Vec3Centre(), sound.MusicDiscEnd{})
		}
	})
}

// jukeboxUser represents an item.User that can use a jukebox.
type jukeboxUser interface {
	item.User
	// SendJukeboxPopup sends a jukebox popup to the item.User.
	SendJukeboxPopup(a ...any)
}

// Activate ...
func (j Jukebox) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	if _, hasDisc := j.Disc(); hasDisc {
		dropItem(tx, j.Item, pos.Side(cube.FaceUp).Vec3Middle())

		j.Item = item.Stack{}
		tx.SetBlock(pos, j, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.MusicDiscEnd{})
	} else if held, _ := u.HeldItems(); !held.Empty() {
		if m, ok := held.Item().(item.MusicDisc); ok {
			j.Item = held

			tx.SetBlock(pos, j, nil)
			tx.PlaySound(pos.Vec3Centre(), sound.MusicDiscEnd{})
			ctx.SubtractFromCount(1)

			tx.PlaySound(pos.Vec3Centre(), sound.MusicDiscPlay{DiscType: m.DiscType})
			if u, ok := u.(jukeboxUser); ok {
				u.SendJukeboxPopup(fmt.Sprintf("Now playing: %v - %v", m.DiscType.Author(), m.DiscType.DisplayName()))
			}
		}
	}
	return true
}

// Disc returns the currently playing music disc
func (j Jukebox) Disc() (sound.DiscType, bool) {
	if !j.Item.Empty() {
		if m, ok := j.Item.Item().(item.MusicDisc); ok {
			return m.DiscType, true
		}
	}
	return sound.DiscType{}, false
}

// EncodeNBT ...
func (j Jukebox) EncodeNBT() map[string]any {
	m := map[string]any{"id": "Jukebox"}
	if _, hasDisc := j.Disc(); hasDisc {
		m["RecordItem"] = nbtconv.WriteItem(j.Item, true)
	}
	return m
}

// DecodeNBT ...
func (j Jukebox) DecodeNBT(data map[string]any) any {
	s := nbtconv.MapItem(data, "RecordItem")
	if _, ok := s.Item().(item.MusicDisc); ok {
		j.Item = s
	}
	return j
}

// EncodeItem ...
func (Jukebox) EncodeItem() (name string, meta int16) {
	return "minecraft:jukebox", 0
}

// EncodeBlock ...
func (Jukebox) EncodeBlock() (string, map[string]any) {
	return "minecraft:jukebox", nil
}
