package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
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
}

// newSmelter initializes a new smelter and returns it.
func newSmelter() *smelter {
	s := &smelter{}
	s.inventory = inventory.New(3, func(slot int, item item.Stack) {
		s.mu.Lock()
		defer s.mu.Unlock()
		for viewer := range s.viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})
	return s
}

// RemainingDuration returns the remaining time the fuel in the smelter will provide.
func (s *smelter) RemainingDuration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.remainingDuration
}

// MaxDuration is the maximum time the fuel can last.
func (s *smelter) MaxDuration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.maxDuration
}

// CookDuration returns the remaining time the input can be cooked for before consuming more fuel.
func (s *smelter) CookDuration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cookDuration
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

// tickSmelting ticks the smelter, ensuring the necessary items exist in the furnace, and then processing all inputted
// items for the necessary duration.
func (s *smelter) tickSmelting(w *world.World, pos cube.Pos, speed int) (update bool, schedule bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prevCookDuration := s.cookDuration
	prevRemainingDuration := s.remainingDuration
	prevMaxDuration := s.maxDuration

	input, _ := s.inventory.Item(0)
	fuel, _ := s.inventory.Item(1)
	product, _ := s.inventory.Item(2)

	var info FuelInfo
	if f, ok := fuel.Item().(Fuel); ok {
		info = f.FuelInfo()
	}

	results, _ := recipe.Perform(w.Block(pos), []world.Item{input.Item()})
	if len(results) == 0 {
		results = append(results, item.Stack{})
	}

	result := results[0] // We only need the first result - this is a smelting recipe.

	canSmelt := input.Count() > 0 && (result.Comparable(product)) && !result.Empty() && product.Count() < product.MaxCount()
	if s.remainingDuration <= 0 && canSmelt && info.Duration > 0 && fuel.Count() > 0 {
		s.remainingDuration, s.maxDuration, update = info.Duration, info.Duration, true
		_ = s.Inventory().SetItem(1, info.Residue)
	}

	if s.remainingDuration > 0 {
		s.remainingDuration -= time.Millisecond * 50
		update, schedule = true, true

		if canSmelt {
			s.cookDuration += time.Millisecond * 50
			if requirement := time.Second * 4 / time.Duration(speed); s.cookDuration >= requirement {
				product = item.NewStack(result.Item(), product.Count()+1)

				_ = s.inventory.SetItem(2, product)
				_ = s.inventory.SetItem(0, input.Grow(-1))
				s.cookDuration -= requirement
			}
		} else if s.remainingDuration <= 0 {
			s.cookDuration, s.maxDuration, s.remainingDuration = 0, 0, 0
		} else {
			s.cookDuration = 0
		}
	} else {
		s.cookDuration, s.maxDuration, s.remainingDuration, update = 0, 0, 0, true
	}

	for v := range s.viewers {
		v.ViewFurnaceUpdate(prevCookDuration, s.cookDuration, prevRemainingDuration, s.remainingDuration, prevMaxDuration, s.maxDuration)
	}
	return update, schedule
}
