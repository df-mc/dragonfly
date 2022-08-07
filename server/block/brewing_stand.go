package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
	"time"
)

// BrewingStand is a block used for brewing potions, splash potions, and lingering potions. It also serves as a cleric's
// job site block generated in village churches.
type BrewingStand struct {
	// LeftSlot is true if the left slot is filled.
	LeftSlot bool
	// MiddleSlot is true if the middle slot is filled.
	MiddleSlot bool
	// RightSlot is true if the right slot is filled.
	RightSlot bool

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}

	brewDuration    time.Duration
	fuelDuration    time.Duration
	maxFuelDuration time.Duration
}

// NewBrewingStand creates a new initialised brewing stand. The inventory is properly initialised.
func NewBrewingStand() BrewingStand {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return BrewingStand{
		inventory: inventory.New(5, func(slot int, item item.Stack) {
			m.RLock()
			defer m.RUnlock()
			for viewer := range v {
				viewer.ViewSlotChange(slot, item)
			}
		}),
		viewerMu: m,
		viewers:  v,
	}
}

// Model ...
func (b BrewingStand) Model() world.BlockModel {
	return model.BrewingStand{}
}

// AddViewer ...
func (b BrewingStand) AddViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	b.viewerMu.Lock()
	defer b.viewerMu.Unlock()
	b.viewers[v] = struct{}{}
}

// RemoveViewer ...
func (b BrewingStand) RemoveViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	b.viewerMu.Lock()
	defer b.viewerMu.Unlock()
	delete(b.viewers, v)
}

// Inventory ...
func (b BrewingStand) Inventory() *inventory.Inventory {
	return b.inventory
}

// Tick is called to check if the furnace should update and start or stop smelting.
func (b BrewingStand) Tick(_ int64, pos cube.Pos, w *world.World) {
	left, _ := b.inventory.Item(1)
	middle, _ := b.inventory.Item(2)
	right, _ := b.inventory.Item(3)

	displayLeft, displayMiddle, displayRight := b.LeftSlot, b.MiddleSlot, b.RightSlot
	b.LeftSlot, b.MiddleSlot, b.RightSlot = !left.Empty(), !middle.Empty(), !right.Empty()
	if b.LeftSlot != displayLeft || b.MiddleSlot != displayMiddle || b.RightSlot != displayRight {
		w.SetBlock(pos, b, nil)
	}
}

// Activate ...
func (b BrewingStand) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// UseOnBlock ...
func (b BrewingStand) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, b)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	b = NewBrewingStand()

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// EncodeNBT ...
func (b BrewingStand) EncodeNBT() map[string]any {
	if b.inventory == nil {
		//noinspection GoAssignmentToReceiver
		b = NewBrewingStand()
	}
	return map[string]any{
		"id":         "BrewingStand",
		"CookTime":   int16(b.brewDuration.Milliseconds() / 50),
		"FuelAmount": int16(b.fuelDuration.Milliseconds() / 50),
		"FuelTotal":  int16(b.maxFuelDuration.Milliseconds() / 50),
		"Items":      nbtconv.InvToNBT(b.inventory),
	}
}

// DecodeNBT ...
func (b BrewingStand) DecodeNBT(data map[string]any) any {
	//noinspection GoAssignmentToReceiver
	b = NewBrewingStand()
	b.brewDuration = time.Duration(nbtconv.Map[int16](data, "CookTime")) * time.Millisecond * 50
	b.fuelDuration = time.Duration(nbtconv.Map[int16](data, "FuelAmount")) * time.Millisecond * 50
	b.maxFuelDuration = time.Duration(nbtconv.Map[int16](data, "FuelTotal")) * time.Millisecond * 50
	nbtconv.InvFromNBT(b.inventory, nbtconv.Map[[]any](data, "Items"))
	return b
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
