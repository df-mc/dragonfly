package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Lodestone is a block that compasses may be linked to in any dimension.
type Lodestone struct {
	solid

	trackingHandle int32
}

// BreakInfo ...
func (l Lodestone) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(Lodestone{}))
}

// Activate links or relinks a compass to the lodestone.
func (l Lodestone) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	compass, ok := held.Item().(item.Compass)
	if !ok {
		return false
	}
	relink := compass.TrackingHandle != 0
	l.trackingHandle = tx.World().TrackPosition(pos, l.trackingHandle)
	tx.SetBlock(pos, l, nil)
	// Delayed a tick so the slot holding the linked compass reaches the client first: otherwise the in-hand
	// and inventory renderers cache different angles.
	tx.ScheduleBlockUpdate(pos, l, time.Second/20)
	linked := held.WithItem(item.Compass{TrackingHandle: l.trackingHandle})
	if relink {
		ctx.NewItem = linked
		ctx.ReplaceHeldItem = true
	} else {
		// One compass of the stack is converted, leaving the rest unlinked.
		ctx.NewItem = linked.Grow(1 - held.Count())
		ctx.SubtractFromCount(1)
	}
	tx.PlaySound(pos.Vec3Centre(), sound.LodestoneCompassLink{})
	return true
}

// ScheduledTick sends the tracking update delayed by Activate.
func (l Lodestone) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if l.trackingHandle == 0 {
		// The linked lodestone was replaced by an unlinked one before the tick ran.
		return
	}
	viewers := tx.Viewers(pos.Vec3Centre())
	if len(viewers) == 0 {
		return
	}
	dim, _ := world.DimensionID(tx.World().Dimension())
	for _, viewer := range viewers {
		viewer.ViewBlockAction(pos, world.PositionTrackingUpdateAction{
			Handle: l.trackingHandle, Position: pos, Dimension: dim,
		})
	}
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
