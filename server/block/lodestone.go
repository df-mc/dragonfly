package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Lodestone is a block that compasses may be linked to in any dimension.
type Lodestone struct {
	solid
	bassDrum

	trackingHandle int32
}

// BreakInfo ...
func (l Lodestone) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(Lodestone{}))
}

// LinkCompass links or relinks the compass held to the lodestone.
func (l Lodestone) LinkCompass(pos cube.Pos, tx *world.Tx, held item.Stack, ctx *item.UseContext) {
	compass, _ := held.Item().(item.Compass)
	issued := l.trackingHandle == 0 && tx.World().PositionTrackingHandleAt(pos) == 0
	l.trackingHandle = tx.World().TrackPosition(pos, l.trackingHandle)
	tx.SetBlock(pos, l, nil)
	if issued {
		tx.World().BroadcastPositionTracking(pos, l.trackingHandle)
	}
	if compass.TrackingHandle != 0 {
		ctx.NewItem = held.WithItem(item.Compass{TrackingHandle: l.trackingHandle})
		ctx.ReplaceHeldItem = true
	} else {
		// Only the name and curse of vanishing carry over to the new compass.
		linked := item.NewStack(item.Compass{TrackingHandle: l.trackingHandle}, 1)
		if name := held.CustomName(); name != "" {
			linked = linked.WithCustomName(name)
		}
		if e, ok := held.Enchantment(enchantment.CurseOfVanishing); ok {
			linked = linked.WithEnchantments(e)
		}
		ctx.NewItem = linked
		ctx.SubtractFromCount(1)
	}
	tx.PlaySound(pos.Vec3Centre(), sound.LodestoneCompassLink{})
}

// TrackingHandle ...
func (l Lodestone) TrackingHandle() int32 { return l.trackingHandle }

// WithTrackingHandle ...
func (l Lodestone) WithTrackingHandle(handle int32) world.Block {
	l.trackingHandle = handle
	return l
}

// EncodeNBT ...
func (l Lodestone) EncodeNBT() map[string]any {
	return map[string]any{"id": "Lodestone", "trackingHandle": l.trackingHandle}
}

// DecodeNBT ...
func (l Lodestone) DecodeNBT(data map[string]any) any {
	l.trackingHandle = nbtconv.Int32(data, "trackingHandle")
	return l
}

// EncodeItem ...
func (Lodestone) EncodeItem() (string, int16) { return "minecraft:lodestone", 0 }

// EncodeBlock ...
func (Lodestone) EncodeBlock() (string, map[string]any) { return "minecraft:lodestone", nil }
