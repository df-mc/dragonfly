package inventory

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"golang.org/x/exp/slices"
	"math"
	"strings"
	"sync"
)

// Inventory represents an inventory containing items. These inventories may be carried by entities or may be
// held by blocks such as chests.
// The size of an inventory may be specified upon construction, but cannot be changed after. The zero value of
// an inventory is invalid. Use New() to obtain a new inventory.
// Inventory is safe for concurrent usage: Its values are protected by a mutex.
type Inventory struct {
	mu    sync.RWMutex
	h     Handler
	slots []item.Stack

	f      func(slot int, before, after item.Stack)
	canAdd func(s item.Stack, slot int) bool
}

// ErrSlotOutOfRange is returned by any methods on inventory when a slot is passed which is not within the
// range of valid values for the inventory.
var ErrSlotOutOfRange = errors.New("slot is out of range: must be in range 0 <= slot < inventory.Size()")

// New creates a new inventory with the size passed. The inventory size cannot be changed after it has been
// constructed.
// A function may be passed which is called every time a slot is changed. The function may also be nil, if
// nothing needs to be done.
func New(size int, f func(slot int, before, after item.Stack)) *Inventory {
	if size <= 0 {
		panic("inventory size must be at least 1")
	}
	if f == nil {
		f = func(slot int, before, after item.Stack) {}
	}
	return &Inventory{h: NopHandler{}, slots: make([]item.Stack, size), f: f, canAdd: func(s item.Stack, slot int) bool { return true }}
}

// Item attempts to obtain an item from a specific slot in the inventory. If an item was present in that slot,
// the item is returned and the error is nil. If no item was present in the slot, a Stack with air as its item
// and a count of 0 is returned. Stack.Empty() may be called to check if this is the case.
// Item only returns an error if the slot passed is out of range. (0 <= slot < inventory.Size())
func (inv *Inventory) Item(slot int) (item.Stack, error) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	inv.check()
	if !inv.validSlot(slot) {
		return item.Stack{}, ErrSlotOutOfRange
	}
	return inv.slots[slot], nil
}

// SetItem sets a stack of items to a specific slot in the inventory. If an item is already present in the
// slot, that item will be overwritten.
// SetItem will return an error if the slot passed is out of range. (0 <= slot < inventory.Size())
func (inv *Inventory) SetItem(slot int, item item.Stack) error {
	inv.mu.Lock()

	inv.check()
	if !inv.validSlot(slot) {
		inv.mu.Unlock()
		return ErrSlotOutOfRange
	}
	f := inv.setItem(slot, item)

	inv.mu.Unlock()

	f()
	return nil
}

// Slots returns the all slots in the inventory as a slice. The index in the slice is the slot of the inventory that a
// specific item.Stack is in. Note that this item.Stack might be empty.
func (inv *Inventory) Slots() []item.Stack {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return slices.Clone(inv.slots)
}

// Items returns a list of all contents of the inventory. This method excludes air items, so the method
// only ever returns item stacks which actually represent an item.
func (inv *Inventory) Items() []item.Stack {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]item.Stack, 0, len(inv.slots))
	for _, it := range inv.slots {
		if !it.Empty() {
			items = append(items, it)
		}
	}
	return items
}

// First returns the first slot with an item if found. Second return value describes whether the item was found.
func (inv *Inventory) First(item item.Stack) (int, bool) {
	return inv.FirstFunc(item.Comparable)
}

// FirstFunc finds the first slot with an item.Stack that results in the comparable function passed returning true. The
// function returns false if no such item was found.
func (inv *Inventory) FirstFunc(comparable func(stack item.Stack) bool) (int, bool) {
	for slot, it := range inv.Slots() {
		if !it.Empty() && comparable(it) {
			return slot, true
		}
	}
	return -1, false
}

// FirstEmpty returns the first empty slot if found. Second return value describes whether an empty slot was found.
func (inv *Inventory) FirstEmpty() (int, bool) {
	for slot, it := range inv.Slots() {
		if it.Empty() {
			return slot, true
		}
	}
	return -1, false
}

// Swap swaps the items between two slots. Returns an error if either slot A or B are invalid.
func (inv *Inventory) Swap(slotA, slotB int) error {
	inv.mu.Lock()

	inv.check()
	if !inv.validSlot(slotA) || !inv.validSlot(slotB) {
		inv.mu.Unlock()
		return ErrSlotOutOfRange
	}
	a, b := inv.slots[slotA], inv.slots[slotB]
	fa, fb := inv.setItem(slotA, b), inv.setItem(slotB, a)

	inv.mu.Unlock()

	fa()
	fb()
	return nil
}

// AddItem attempts to add an item to the inventory. It does so in a couple of steps: It first iterates over
// the inventory to make sure no existing stacks of the same type exist. If these stacks do exist, the item
// added is first added on top of those stacks to make sure they are fully filled.
// If no existing stacks with leftover space are left, empty slots will be filled up with the remainder of the
// item added.
// If the item could not be fully added to the inventory, an error is returned along with the count that was
// added to the inventory.
func (inv *Inventory) AddItem(it item.Stack) (n int, err error) {
	if it.Empty() {
		return 0, nil
	}
	first := it.Count()
	emptySlots := make([]int, 0, 16)

	inv.mu.Lock()

	inv.check()
	for slot, invIt := range inv.slots {
		if invIt.Empty() {
			// This slot was empty, and we should first try to add the item stack to existing stacks.
			emptySlots = append(emptySlots, slot)
			continue
		}
		a, b := invIt.AddStack(it)
		if it.Count() == b.Count() {
			// Count stayed the same, meaning this slot either wasn't equal to this stack or was max size.
			continue
		}
		f := inv.setItem(slot, a)
		//noinspection GoDeferInLoop
		defer f()

		if it = b; it.Empty() {
			inv.mu.Unlock()
			// We were able to add the entire stack to existing stacks in the inventory.
			return first, nil
		}
	}
	for _, slot := range emptySlots {
		a, b := it.Grow(-math.MaxInt32).AddStack(it)

		f := inv.setItem(slot, a)
		//noinspection GoDeferInLoop
		defer f()

		if it = b; it.Empty() {
			inv.mu.Unlock()
			// We were able to add the entire stack to empty slots.
			return first, nil
		}
	}
	inv.mu.Unlock()
	// We were unable to clear out the entire stack to be added to the inventory: There wasn't enough space.
	return first - it.Count(), fmt.Errorf("could not add full item stack to inventory")
}

// RemoveItem attempts to remove an item from the inventory. It will visit all slots in the inventory and
// empties them until it.Count() items have been removed from the inventory.
// If less than it.Count() items were removed from the inventory, an error is returned.
func (inv *Inventory) RemoveItem(it item.Stack) error {
	return inv.RemoveItemFunc(it.Count(), it.Comparable)
}

// RemoveItemFunc removes up to n items from the Inventory. It will visit all slots in the inventory and empties them
// until n items have been removed from the inventory, assuming the comparable function returns true for the slots
// visited. No items will be deducted from slots if the comparable function returns false.
// If less than n items were removed, an error is returned.
func (inv *Inventory) RemoveItemFunc(n int, comparable func(stack item.Stack) bool) error {
	inv.mu.Lock()
	inv.check()
	for slot, slotIt := range inv.slots {
		if slotIt.Empty() || !comparable(slotIt) {
			continue
		}
		f := inv.setItem(slot, slotIt.Grow(-n))
		//noinspection GoDeferInLoop
		defer f()

		if n -= slotIt.Count(); n <= 0 {
			break
		}
	}
	inv.mu.Unlock()

	if n > 0 {
		return fmt.Errorf("could not remove all items from the inventory")
	}
	return nil
}

// ContainsItem checks if the Inventory contains an item.Stack. It will visit all slots in the Inventory until it finds
// at enough items. If enough were found, true is returned.
func (inv *Inventory) ContainsItem(it item.Stack) bool {
	return inv.ContainsItemFunc(it.Count(), it.Comparable)
}

// ContainsItemFunc checks if the Inventory contains at least n items. It will visit all slots in the Inventory until it
// finds n items on which the comparable function returns true. ContainsItemFunc returns true if this is the case.
func (inv *Inventory) ContainsItemFunc(n int, comparable func(stack item.Stack) bool) bool {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.check()
	for _, slotIt := range inv.slots {
		if !slotIt.Empty() && comparable(slotIt) {
			if n -= slotIt.Count(); n <= 0 {
				break
			}
		}
	}
	return n <= 0
}

// Empty checks if the inventory is fully empty: It iterates over the inventory and makes sure every stack in
// it is empty.
func (inv *Inventory) Empty() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	inv.check()
	for _, it := range inv.slots {
		if !it.Empty() {
			return false
		}
	}
	return true
}

// Clear clears the entire inventory. All non-zero items are returned.
func (inv *Inventory) Clear() []item.Stack {
	inv.mu.Lock()

	inv.check()

	items := make([]item.Stack, 0, inv.size())
	for slot, i := range inv.slots {
		if !i.Empty() {
			items = append(items, i)
			f := inv.setItem(slot, item.Stack{})
			//noinspection GoDeferInLoop
			defer f()
		}
	}
	inv.mu.Unlock()

	return items
}

// Handle assigns a Handler to an Inventory so that its methods are called for the respective events. Nil may be passed
// to set the default NopHandler.
func (inv *Inventory) Handle(h Handler) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.check()
	if h == nil {
		h = NopHandler{}
	}
	inv.h = h
}

// Handler returns the Handler currently assigned to the Inventory. This is the NopHandler by default.
func (inv *Inventory) Handler() Handler {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	inv.check()
	return inv.h
}

// setItem sets an item to a specific slot and overwrites the existing item. It calls the function which is
// called for every item change and does so without locking the inventory.
func (inv *Inventory) setItem(slot int, it item.Stack) func() {
	if !inv.canAdd(it, slot) {
		return func() {}
	}
	if it.Count() > it.MaxCount() {
		it = it.Grow(it.MaxCount() - it.Count())
	}
	before := inv.slots[slot]
	inv.slots[slot] = it
	return func() {
		inv.f(slot, before, it)
	}
}

// Size returns the size of the inventory. It is always the same value as that passed in the call to New() and
// is always at least 1.
func (inv *Inventory) Size() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.size()
}

// size returns the size of the inventory without locking.
func (inv *Inventory) size() int {
	return len(inv.slots)
}

// Close closes the inventory, freeing the function called for every slot change. It also clears any items
// that may currently be in the inventory.
// The returned error is always nil.
func (inv *Inventory) Close() error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.check()
	inv.f = func(int, item.Stack, item.Stack) {}
	return nil
}

// String implements the fmt.Stringer interface.
func (inv *Inventory) String() string {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	s := make([]string, 0, inv.size())
	for _, it := range inv.slots {
		s = append(s, it.String())
	}
	return "(" + strings.Join(s, ", ") + ")"
}

// validSlot checks if the slot passed is valid for the inventory. It returns false if the slot is either
// smaller than 0 or bigger/equal to the size of the inventory's size.
func (inv *Inventory) validSlot(slot int) bool {
	return slot >= 0 && slot < inv.size()
}

// check panics if the inventory is valid, and panics if it is not. This typically happens if the inventory
// was not created using New().
func (inv *Inventory) check() {
	if inv.size() == 0 {
		panic("uninitialised inventory: inventory must be constructed using inventory.New()")
	}
}
