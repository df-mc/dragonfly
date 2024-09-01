package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"sync"
	"time"
)

// brewer is a struct that may be embedded by blocks that can brew potions, such as brewing stands.
type brewer struct {
	mu sync.Mutex

	viewers   map[ContainerViewer]struct{}
	inventory *inventory.Inventory

	duration   time.Duration
	fuelAmount int32
	fuelTotal  int32
}

// newBrewer creates a new initialised brewer. The inventory is properly initialised.
func newBrewer() *brewer {
	b := &brewer{viewers: make(map[ContainerViewer]struct{})}
	b.inventory = inventory.New(5, func(slot int, _, item item.Stack) {
		b.mu.Lock()
		defer b.mu.Unlock()
		for viewer := range b.viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})
	return b
}

// InsertItem ...
func (b *brewer) InsertItem(h Hopper, pos cube.Pos, w *world.World) bool {
	for sourceSlot, sourceStack := range h.inventory.Slots() {
		var slot int

		if sourceStack.Empty() {
			continue
		}

		if h.Facing == cube.FaceDown {
			slot = 0
		} else {
			if _, ok := sourceStack.Item().(item.BlazePowder); ok {
				slot = 4
			}

			for brewingSlot, brewingStack := range b.inventory.Slots() {
				if brewingSlot == 0 || brewingSlot == 4 {
					continue
				}

				if !brewingStack.Empty() {
					continue
				}

				slot = brewingSlot
				break
			}
		}

		stack := sourceStack.Grow(-sourceStack.Count() + 1)
		it, _ := b.Inventory(w, pos).Item(slot)
		if slot == 0 {
			//TODO: check if the item is a brewing ingredient.
		}
		if !sourceStack.Comparable(it) {
			// The items are not the same.
			continue
		}
		if it.Count() == it.MaxCount() {
			// The item has the maximum count that the stack is able to hold.
			continue
		}
		if !it.Empty() {
			stack = it.Grow(1)
		}

		_ = b.Inventory(w, pos).SetItem(slot, stack)
		_ = h.inventory.SetItem(sourceSlot, sourceStack.Grow(-1))
		return true

	}
	return false
}

// ExtractItem ...
func (b *brewer) ExtractItem(h Hopper, pos cube.Pos, w *world.World) bool {
	for sourceSlot, sourceStack := range b.inventory.Slots() {
		if sourceStack.Empty() {
			continue
		}

		if sourceSlot == 0 || sourceSlot == 4 {
			continue
		}

		fmt.Println(sourceSlot)

		_, err := h.inventory.AddItem(sourceStack.Grow(-sourceStack.Count() + 1))
		if err != nil {
			// The hopper is full.
			continue
		}

		_ = b.Inventory(w, pos).SetItem(sourceSlot, sourceStack.Grow(-1))
		return true
	}
	return false
}

// Duration returns the remaining duration of the brewing process.
func (b *brewer) Duration() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.duration
}

// Fuel returns the fuel and maximum fuel of the brewer.
func (b *brewer) Fuel() (fuel, maxFuel int32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.fuelAmount, b.fuelTotal
}

// Inventory returns the inventory of the brewer.
func (b *brewer) Inventory(*world.World, cube.Pos) *inventory.Inventory {
	return b.inventory
}

// AddViewer adds a viewer to the brewer, so that it is updated whenever the inventory of the brewer is changed.
func (b *brewer) AddViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the brewer, so that slot updates in the inventory are no longer sent to
// it.
func (b *brewer) RemoveViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.viewers, v)
}

// setDuration sets the brew duration of the brewer to the given duration.
func (b *brewer) setDuration(duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.duration = duration
}

// setFuel sets the fuel of the brewer to the given fuel and maximum fuel.
func (b *brewer) setFuel(fuel, maxFuel int32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.fuelAmount, b.fuelTotal = fuel, maxFuel
}

// tickBrewing ticks the brewer, ensuring the necessary items exist in the brewer, and then processing all inputted
// items for the necessary duration.
func (b *brewer) tickBrewing(block string, pos cube.Pos, w *world.World) {
	b.mu.Lock()

	// Keep track of our past durations, since if any of them change, we need to be able to tell they did and then
	// update the viewers on the change.
	prevDuration := b.duration
	prevFuelAmount := b.fuelAmount
	prevFuelTotal := b.fuelTotal

	// If we need fuel, try and burn some.
	fuel, _ := b.inventory.Item(4)

	if _, ok := fuel.Item().(item.BlazePowder); ok && b.fuelAmount <= 0 {
		defer b.inventory.SetItem(4, fuel.Grow(-1))
		b.fuelAmount, b.fuelTotal = 20, 20
	}

	// Update the viewers on the new durations.
	for v := range b.viewers {
		v.ViewBrewingUpdate(prevDuration, b.duration, prevFuelAmount, b.fuelAmount, prevFuelTotal, b.fuelTotal)
	}

	b.mu.Unlock()
}
