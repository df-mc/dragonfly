package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Jukebox is a block used to play music discs.
type Jukebox struct {
	solid
	bass

	// item is the music disc played by the jukebox
	item item.Stack
}

// FuelInfo ...
func (j Jukebox) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// BreakInfo ...
func (j Jukebox) BreakInfo() BreakInfo {
	d := []item.Stack{item.NewStack(Jukebox{}, 1)}
	if !j.item.Empty() {
		d = append(d, j.item)
	}
	return newBreakInfo(2, alwaysHarvestable, axeEffective, simpleDrops(d...))
}

// jukeboxUser represents an item.User that can use a jukebox.
type jukeboxUser interface {
	item.User
	// SendJukeboxPopup sends a jukebox popup to the item.User
	SendJukeboxPopup(a ...any)
}

// Activate ...
func (j Jukebox) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User) bool {
	if !j.item.Empty() {
		ent := entity.NewItem(j.item, pos.Side(cube.FaceUp).Vec3Middle())
		ent.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
		w.AddEntity(ent)

		j.ClearMusicDisc(pos, w)
	} else if held, other := u.HeldItems(); !held.Empty() {
		if m, ok := held.Item().(item.MusicDisc); ok {
			j.InsertMusicDisc(m, pos, w)
			u.SetHeldItems(held.Grow(-1), other)

			if u, ok := u.(jukeboxUser); ok {
				u.SendJukeboxPopup(fmt.Sprintf("Now playing: %v - %v", m.DiscType.Author(), m.DiscType.DisplayName()))
			}
		}
	}

	return true
}

func (j Jukebox) InsertMusicDisc(m item.MusicDisc, pos cube.Pos, w *world.World) {
	if !j.item.Empty() {
		j.ClearMusicDisc(pos, w)
	}
	j.item = item.NewStack(m, 1)
	w.PlaySound(pos.Vec3(), sound.MusicDiscPlay{DiscType: m.DiscType})
	w.SetBlock(pos, j, nil)
}

func (j Jukebox) ClearMusicDisc(pos cube.Pos, w *world.World) {
	j.item = item.Stack{}
	w.PlaySound(pos.Vec3(), sound.MusicDiscEnd{})
	w.SetBlock(pos, j, nil)
}

// PostBreak ...
func (j Jukebox) PostBreak(pos cube.Pos, w *world.World, _ item.User) {
	if !j.item.Empty() {
		w.PlaySound(pos.Vec3(), sound.MusicDiscEnd{})
	}
}

// EncodeNBT ...
func (j Jukebox) EncodeNBT() map[string]any {
	m := map[string]any{"id": "Jukebox"}
	if !j.item.Empty() {
		m["RecordItem"] = nbtconv.WriteItem(j.item, true)
	}
	return m
}

// DecodeNBT ...
func (j Jukebox) DecodeNBT(data map[string]any) any {
	j.item = nbtconv.MapItem(data, "RecordItem")
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
