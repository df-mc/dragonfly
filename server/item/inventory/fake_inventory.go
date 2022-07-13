package inventory

import (
	"github.com/df-mc/dragonfly/server/item"
	"sync"
)

// FakeInventory is a wrapper around an Inventory that contains various information used for fake inventories, such as
// viewers, or a custom name.
type FakeInventory struct {
	*Inventory

	customName    string
	inventoryType FakeInventoryType

	viewerMu *sync.RWMutex
	viewers  map[Viewer]struct{}
}

// NewFakeInventory creates a new FakeInventory with the given size and name.
func NewFakeInventory(name string, inventoryType FakeInventoryType) *FakeInventory {
	m := new(sync.RWMutex)
	v := make(map[Viewer]struct{}, 1)
	return &FakeInventory{
		Inventory: New(inventoryType.Size(), func(slot int, item item.Stack) {
			m.RLock()
			defer m.RUnlock()
			for viewer := range v {
				viewer.ViewSlotChange(slot, item)
			}
		}),

		customName:    name,
		inventoryType: inventoryType,

		viewerMu: m,
		viewers:  v,
	}
}

// Name returns the name of the fake inventory.
func (f *FakeInventory) Name() string {
	return f.customName
}

// Type returns the type of the fake inventory.
func (f *FakeInventory) Type() FakeInventoryType {
	return f.inventoryType
}

// AddViewer adds a viewer to the fake inventory, so that it is updated whenever the inventory is changed.
func (f *FakeInventory) AddViewer(v Viewer) {
	f.viewerMu.Lock()
	defer f.viewerMu.Unlock()
	f.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the fake inventory, so that slot updates in the inventory are no longer sent to
// it.
func (f *FakeInventory) RemoveViewer(v Viewer) {
	f.viewerMu.Lock()
	defer f.viewerMu.Unlock()
	delete(f.viewers, v)
}
