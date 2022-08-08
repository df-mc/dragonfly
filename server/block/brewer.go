package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
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
	b.inventory = inventory.New(5, func(slot int, item item.Stack) {
		b.mu.Lock()
		defer b.mu.Unlock()
		for viewer := range b.viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})
	return b
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
func (b *brewer) Inventory() *inventory.Inventory {
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

	// Get each item in the brewer. We don't need to validate errors here since we know the bounds of the brewer.
	left, _ := b.inventory.Item(1)
	middle, _ := b.inventory.Item(2)
	right, _ := b.inventory.Item(3)

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

	// Now get the ingredient item.
	ingredient, _ := b.inventory.Item(0)

	// Check each input and see if it is affected by the ingredient.
	leftOutput, leftAffected := recipe.Perform(block, left.Item(), ingredient.Item())
	middleOutput, middleAffected := recipe.Perform(block, middle.Item(), ingredient.Item())
	rightOutput, rightAffected := recipe.Perform(block, right.Item(), ingredient.Item())

	// Ensure that we have enough fuel to continue.
	if b.fuelAmount > 0 {
		// Now make sure that we have at least one potion that is affected by the ingredient.
		if leftAffected || middleAffected || rightAffected {
			// Tick our duration. If we have no brew duration, set it to the default of twenty seconds.
			if b.duration == 0 {
				b.duration = time.Second * 20
			}
			b.duration -= time.Millisecond * 50

			// If we have no duration, we are done.
			if b.duration <= 0 {
				// Create the output items.
				if leftAffected {
					defer b.inventory.SetItem(1, leftOutput[0])
				}
				if middleAffected {
					defer b.inventory.SetItem(2, middleOutput[0])
				}
				if rightAffected {
					defer b.inventory.SetItem(3, rightOutput[0])
				}

				// Reduce the ingredient by one.
				defer b.inventory.SetItem(0, ingredient.Grow(-1))
				w.PlaySound(pos.Vec3Centre(), sound.PotionBrewed{})

				// Decrement the fuel, and reset the duration.
				b.fuelAmount--
				b.duration = 0
			}
		} else {
			// None of the potions are affected by the ingredient, so reset the duration.
			b.duration = 0
		}
	} else {
		// We don't have enough fuel, so reset our progress.
		b.duration, b.fuelAmount, b.fuelTotal = 0, 0, 0
	}

	// Update the viewers on the new durations.
	for v := range b.viewers {
		v.ViewBrewingUpdate(prevDuration, b.duration, prevFuelAmount, b.fuelAmount, prevFuelTotal, b.fuelTotal)
	}

	b.mu.Unlock()
}
