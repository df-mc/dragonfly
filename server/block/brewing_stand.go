package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// BrewingStand is a block used for brewing potions, splash potions, and lingering potions. It also serves as a cleric's
// job site block generated in village churches.
type BrewingStand struct {
	sourceWaterDisplacer
	transparent
	*brewer

	// LeftSlot is true if the left slot is filled.
	LeftSlot bool
	// MiddleSlot is true if the middle slot is filled.
	MiddleSlot bool
	// RightSlot is true if the right slot is filled.
	RightSlot bool
}

// NewBrewingStand creates a new initialised brewing stand. The inventory is properly initialised.
func NewBrewingStand() BrewingStand {
	return BrewingStand{brewer: newBrewer()}
}

// Model ...
func (b BrewingStand) Model() world.BlockModel {
	return model.BrewingStand{}
}

// SideClosed ...
func (b BrewingStand) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// Tick is called to check if the brewing stand should update and start or stop brewing.
func (b BrewingStand) Tick(_ int64, pos cube.Pos, tx *world.Tx) {
	// Get each item in the brewing stand. We don't need to validate errors here since we know the bounds of the stand.
	left, _ := b.inventory.Item(1)
	middle, _ := b.inventory.Item(2)
	right, _ := b.inventory.Item(3)

	// If any of the slots in the inventory got updated, update the appearance of the brewing stand.
	displayLeft, displayMiddle, displayRight := b.LeftSlot, b.MiddleSlot, b.RightSlot
	b.LeftSlot, b.MiddleSlot, b.RightSlot = !left.Empty(), !middle.Empty(), !right.Empty()
	if b.LeftSlot != displayLeft || b.MiddleSlot != displayMiddle || b.RightSlot != displayRight {
		tx.SetBlock(pos, b, nil)
	}

	// Tick brewing.
	b.tickBrewing("brewing_stand", pos, tx)
}

// Activate ...
func (b BrewingStand) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// UseOnBlock ...
func (b BrewingStand) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, b)
	if !used {
		return
	}

	//noinspection GoAssignmentToReceiver
	b = NewBrewingStand()
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// EncodeNBT ...
func (b BrewingStand) EncodeNBT() map[string]any {
	if b.brewer == nil {
		//noinspection GoAssignmentToReceiver
		b = NewBrewingStand()
	}
	duration := b.Duration()
	fuel, totalFuel := b.Fuel()
	return map[string]any{
		"id":         "BrewingStand",
		"Items":      nbtconv.InvToNBT(b.inventory),
		"CookTime":   int16(duration.Milliseconds() / 50),
		"FuelTotal":  int16(totalFuel),
		"FuelAmount": int16(fuel),
	}
}

// DecodeNBT ...
func (b BrewingStand) DecodeNBT(data map[string]any) any {
	brew := time.Duration(nbtconv.Int16(data, "CookTime")) * time.Millisecond * 50

	fuel := int32(nbtconv.Int16(data, "FuelAmount"))
	maxFuel := int32(nbtconv.Int16(data, "FuelTotal"))

	//noinspection GoAssignmentToReceiver
	b = NewBrewingStand()
	b.setDuration(brew)
	b.setFuel(fuel, maxFuel)
	nbtconv.InvFromNBT(b.inventory, nbtconv.Slice(data, "Items"))
	return b
}

// BreakInfo ...
func (b BrewingStand) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, oneOf(BrewingStand{})).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		for _, i := range b.Inventory(tx, pos).Clear() {
			dropItem(tx, i, pos.Vec3Centre())
		}
	})
}

// EncodeBlock ...
func (b BrewingStand) EncodeBlock() (string, map[string]any) {
	return "minecraft:brewing_stand", map[string]any{
		"brewing_stand_slot_a_bit": b.LeftSlot,
		"brewing_stand_slot_b_bit": b.MiddleSlot,
		"brewing_stand_slot_c_bit": b.RightSlot,
	}
}

// EncodeItem ...
func (b BrewingStand) EncodeItem() (name string, meta int16) {
	return "minecraft:brewing_stand", 0
}

// allBrewingStands ...
func allBrewingStands() (stands []world.Block) {
	for _, left := range []bool{false, true} {
		for _, middle := range []bool{false, true} {
			for _, right := range []bool{false, true} {
				stands = append(stands, BrewingStand{LeftSlot: left, MiddleSlot: middle, RightSlot: right})
			}
		}
	}
	return
}
