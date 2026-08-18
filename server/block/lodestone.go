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
	// Send the tracking update on the next world tick. The inventory slot
	// containing the linked compass must reach the client in an earlier network
	// batch; otherwise the in-hand and inventory renderers may cache different
	// angles.
	tx.ScheduleBlockUpdate(pos, l, time.Second/20)
	linked := held.WithItem(item.Compass{TrackingHandle: l.trackingHandle})
	if relink {
		// Relinking a lodestone compass updates the complete stack in-place.
		ctx.NewItem = linked
		ctx.ReplaceHeldItem = true
	} else {
		// Linking regular compasses consumes one and produces one separate
		// lodestone compass, leaving the rest of the regular stack untouched.
		ctx.NewItem = linked.Grow(1 - held.Count())
		ctx.SubtractFromCount(1)
	}
	tx.PlaySound(pos.Vec3Centre(), sound.LodestoneCompassLink{})
	return true
}

// ScheduledTick sends the delayed position tracking update for newly linked
// lodestone compasses.
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
