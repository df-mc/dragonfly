package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// ItemFrame is a block entity that displays the item or block that is inside it.
type ItemFrame struct {
	empty
	transparent

	// Facing is the direction from the frame to the block.
	Facing cube.Face
	// Item is the item that is displayed inside the frame.
	Item item.Stack
	// Rotations is the number of rotations for the item in the frame. Each rotation is 45 degrees, with the exception
	// being maps having 90 degree rotations.
	Rotations int
	// DropChance is the chance of the item dropping when the frame is broken. In vanilla, this is always 1.0.
	DropChance float64
	// Glowing makes the frame the glowing variant.
	Glowing bool
}

// Activate ...
func (i ItemFrame) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User) bool {
	if !i.Item.Empty() {
		// TODO: Item frames with maps can only be rotated four times.
		i.Rotations = (i.Rotations + 1) % 8
		w.PlaySound(pos.Vec3Centre(), sound.ItemFrameRotate{})
	} else if held, other := u.HeldItems(); !held.Empty() {
		i.Item = held.Grow(-held.Count() + 1)
		// TODO: When maps are implemented, check the item is a map, and if so, display the large version of the frame.
		u.SetHeldItems(held.Grow(-1), other)
		w.PlaySound(pos.Vec3Centre(), sound.ItemFrameAdd{})
	} else {
		return true
	}

	w.SetBlock(pos, i, nil)
	return true
}

// Punch ...
func (i ItemFrame) Punch(pos cube.Pos, _ cube.Face, w *world.World, u item.User) {
	if i.Item.Empty() {
		return
	}

	if g, ok := u.(interface {
		GameMode() world.GameMode
	}); ok {
		if rand.Float64() <= i.DropChance && !g.GameMode().CreativeInventory() {
			it := entity.NewItem(i.Item, pos.Vec3Centre())
			it.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
			w.AddEntity(it)
		}
	}
	i.Item, i.Rotations = item.Stack{}, 0
	w.PlaySound(pos.Vec3Centre(), sound.ItemFrameRemove{})
	w.SetBlock(pos, i, nil)
}

// UseOnBlock ...
func (i ItemFrame) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, i)
	if !used {
		return false
	}
	if (w.Block(pos.Side(face.Opposite())).Model() == model.Empty{}) {
		// TODO: Allow exceptions for pressure plates.
		return false
	}
	i.Facing = face.Opposite()
	i.DropChance = 1.0

	place(w, pos, i, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (i ItemFrame) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(i))
}

// EncodeItem ...
func (i ItemFrame) EncodeItem() (name string, meta int16) {
	if i.Glowing {
		return "minecraft:glow_frame", 0
	}
	return "minecraft:frame", 0
}

// EncodeBlock ...
func (i ItemFrame) EncodeBlock() (name string, properties map[string]any) {
	name = "minecraft:frame"
	if i.Glowing {
		name = "minecraft:glow_frame"
	}
	return name, map[string]any{
		"facing_direction":     int32(i.Facing.Opposite()),
		"item_frame_map_bit":   uint8(0), // TODO: When maps are added, set this to true if the item is a map.
		"item_frame_photo_bit": uint8(0), // Only implemented in Education Edition.
	}
}

// DecodeNBT ...
func (i ItemFrame) DecodeNBT(data map[string]any) any {
	i.DropChance = float64(nbtconv.Map[float32](data, "ItemDropChance"))
	i.Rotations = int(nbtconv.Map[byte](data, "ItemRotation"))
	i.Item = nbtconv.MapItem(data, "Item")
	return i
}

// EncodeNBT ...
func (i ItemFrame) EncodeNBT() map[string]any {
	m := map[string]any{
		"ItemDropChance": float32(i.DropChance),
		"ItemRotation":   uint8(i.Rotations),
		"id":             "ItemFrame",
	}
	if i.Glowing {
		m["id"] = "GlowItemFrame"
	}
	if !i.Item.Empty() {
		m["Item"] = nbtconv.WriteItem(i.Item, true)
	}
	return m
}

// Pick returns the item that is picked when the block is picked.
func (i ItemFrame) Pick() item.Stack {
	if i.Item.Empty() {
		return item.NewStack(ItemFrame{Glowing: i.Glowing}, 1)
	}
	return item.NewStack(i.Item.Item(), 1)
}

// CanDisplace ...
func (ItemFrame) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (ItemFrame) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// NeighbourUpdateTick ...
func (i ItemFrame) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if (w.Block(pos.Side(i.Facing)).Model() == model.Empty{}) {
		// TODO: Allow exceptions for pressure plates.
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: i})
	}
}

// allItemFrames ...
func allItemFrames() (frames []world.Block) {
	for _, f := range cube.Faces() {
		frames = append(frames, ItemFrame{Facing: f, Glowing: true})
		frames = append(frames, ItemFrame{Facing: f, Glowing: false})
	}
	return
}
