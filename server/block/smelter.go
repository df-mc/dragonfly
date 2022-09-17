package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"math/rand"
	"sync"
	"time"
)

// smelter is a struct that may be embedded by blocks that can smelt blocks and items, such as blast furnaces, furnaces,
// and smokers.
type smelter struct {
	mu sync.Mutex

	viewers   map[ContainerViewer]struct{}
	inventory *inventory.Inventory

	remainingDuration time.Duration
	cookDuration      time.Duration
	maxDuration       time.Duration
	experience        int
}

// newSmelter initializes a new smelter with the given remaining, maximum, and cook durations and XP, and returns it.
func newSmelter() *smelter {
	s := &smelter{viewers: make(map[ContainerViewer]struct{})}
	s.inventory = inventory.New(3, func(slot int, _, item item.Stack) {
		s.mu.Lock()
		defer s.mu.Unlock()
		for viewer := range s.viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})
	return s
}

// Durations returns the remaining, maximum, and cook durations of the smelter.
func (s *smelter) Durations() (remaining time.Duration, max time.Duration, cook time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.remainingDuration, s.maxDuration, s.cookDuration
}

// Experience returns the collected experience of the smelter.
func (s *smelter) Experience() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.experience
}

// ResetExperience resets the collected experience of the smelter, and returns the amount of experience that was reset.
func (s *smelter) ResetExperience() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	xp := s.experience
	s.experience = 0
	return xp
}

// Inventory returns the inventory of the furnace.
func (s *smelter) Inventory() *inventory.Inventory {
	return s.inventory
}

// AddViewer adds a viewer to the furnace, so that it is updated whenever the inventory of the furnace is changed.
func (s *smelter) AddViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the furnace, so that slot updates in the inventory are no longer sent to
// it.
func (s *smelter) RemoveViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.viewers) == 0 {
		// No viewers.
		return
	}
	delete(s.viewers, v)
}

// setExperience sets the collected experience of the smelter to the given value.
func (s *smelter) setExperience(xp int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.experience = xp
}

// setDurations sets the remaining, maximum, and cook durations of the smelter to the given values.
func (s *smelter) setDurations(remaining, max, cook time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remainingDuration, s.maxDuration, s.cookDuration = remaining, max, cook
}

// tickSmelting ticks the smelter, ensuring the necessary items exist in the furnace, and then processing all inputted
// items for the necessary duration.
func (s *smelter) tickSmelting(requirement, decrement time.Duration, lit bool, supported func(item.SmeltInfo) bool) bool {
	s.mu.Lock()

	// First keep track of our past durations, since if any of them change, we need to be able to tell they did and then
	// update the viewers on the change.
	prevCookDuration := s.cookDuration
	prevRemainingDuration := s.remainingDuration
	prevMaxDuration := s.maxDuration

	// Now get each item in the smelter. We don't need to validate errors here since we know the bounds of the smelter.
	input, _ := s.inventory.Item(0)
	fuel, _ := s.inventory.Item(1)
	product, _ := s.inventory.Item(2)

	// Initialize some default smelt info, and update it if we can smelt the item.
	var inputInfo item.SmeltInfo
	if i, ok := input.Item().(item.Smeltable); ok && supported(i.SmeltInfo()) {
		inputInfo = i.SmeltInfo()
	}

	// Initialize some default fuel info, and update it if it can be used as fuel.
	var fuelInfo item.FuelInfo
	if f, ok := fuel.Item().(item.Fuel); ok {
		fuelInfo = f.FuelInfo()
		if fuelInfo.Residue.Empty() {
			// If we don't have a custom residue set, then we just decrement the fuel by one.
			fuelInfo.Residue = fuel.Grow(-1)
		}
	}

	// Now we need to ensure that we can actually smelt the item. We need to ensure that we have at least one input,
	// the input's product is compatible with the product already in the product slot, the product slot is not full,
	// and that we have enough fuel to smelt the item. If all of these conditions are met, then we update the remaining
	// duration and cook duration and create residue.
	canSmelt := input.Count() > 0 && (inputInfo.Product.Comparable(product)) && !inputInfo.Product.Empty() && product.Count() < product.MaxCount()
	if s.remainingDuration <= 0 && canSmelt && fuelInfo.Duration > 0 && fuel.Count() > 0 {
		s.remainingDuration, s.maxDuration, lit = fuelInfo.Duration, fuelInfo.Duration, true
		defer s.inventory.SetItem(1, fuelInfo.Residue)
	}

	// Now we need to process a single stage of fuel loss. First, ensure that we have enough remaining duration.
	if s.remainingDuration > 0 {
		// Decrement a tick from the remaining fuel duration.
		s.remainingDuration -= time.Millisecond * 50

		// If we have a valid smeltable item, process a single stage of smelting.
		if canSmelt {
			// Increase the cook duration by a tick.
			s.cookDuration += time.Millisecond * 50

			// Check if we've cooked enough to match the requirement.
			if s.cookDuration >= requirement {
				// We can now create the product and reduce the input by one.
				defer s.inventory.SetItem(0, input.Grow(-1))
				defer s.inventory.SetItem(2, item.NewStack(inputInfo.Product.Item(), product.Count()+inputInfo.Product.Count()))

				// Calculate the amount of experience to grant. Round the experience down to the nearest integer.
				// The remaining XP is a chance to be granted an additional experience point.
				xp := inputInfo.Experience * float64(inputInfo.Product.Count())
				earned := math.Floor(inputInfo.Experience)
				if chance := xp - earned; chance > 0 && rand.Float64() < chance {
					earned++
				}

				// Decrease the cook duration by the requirement, and update the smelter's stored experience.
				s.cookDuration -= requirement
				s.experience += int(earned)
			}
		} else if s.remainingDuration == 0 {
			// We've run out of fuel, so we need to reset the max duration too.
			s.maxDuration = 0
		} else {
			// We still have some remaining fuel, but the input isn't smeltable, so we reset the cook duration.
			s.cookDuration = 0
		}
	} else {
		// We don't have any more remaining duration, so we need to reset the max duration and put out the furnace.
		s.maxDuration, lit = 0, false
	}

	// We've run out of fuel, but we have some remaining cook duration, so instead of stopping entirely, we reduce the
	// cook duration by the decrement.
	if s.cookDuration > 0 && !lit {
		s.cookDuration -= decrement
	}

	// Update the viewers on the new durations.
	for v := range s.viewers {
		v.ViewFurnaceUpdate(prevCookDuration, s.cookDuration, prevRemainingDuration, s.remainingDuration, prevMaxDuration, s.maxDuration)
	}

	s.mu.Unlock()
	return lit
}
