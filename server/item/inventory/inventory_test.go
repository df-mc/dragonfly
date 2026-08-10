package inventory

import (
	"sync"
	"testing"

	"github.com/df-mc/dragonfly/server/item"
)

// TestSlotFuncRace verifies that changing the function called on a slot change does not race with a slot being
// written. setItem returns a closure that is called after the inventory is unlocked, so the function it calls has to
// be read while the lock is still held.
//
// This only fails under the race detector: run it with go test -race.
func TestSlotFuncRace(t *testing.T) {
	tests := []struct {
		name string
		set  func(inv *Inventory)
	}{
		{name: "slot function replaced", set: func(inv *Inventory) { inv.SlotFunc(func(int, item.Stack, item.Stack) {}) }},
		{name: "inventory closed", set: func(inv *Inventory) { _ = inv.Close() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := New(1, func(int, item.Stack, item.Stack) {})

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for range 200 {
					_ = inv.SetItem(0, item.NewStack(item.Diamond{}, 1))
				}
			}()
			go func() {
				defer wg.Done()
				for range 200 {
					tt.set(inv)
				}
			}()
			wg.Wait()
		})
	}
}
