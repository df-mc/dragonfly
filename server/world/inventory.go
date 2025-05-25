package world

import (
	"io"

	handles "github.com/royalmcpe/golang-handles-map"
)

type InventoryHandle handles.Handle

type InventoryHolder interface {
	Inventory() *handles.Handle
}

type Container interface {
	io.Closer
	Item(slot int) (ItemStack, error)
	SetItem(slot int, item ItemStack) error
	Slots() []ItemStack
	Items() []ItemStack
	First(ItemStack) (int, bool)
	FirstEmpty() (int, bool)
	Swap(slotA, slotB int) error
	AddItem(ItemStack) (n int, err error)
	RemoveItem(ItemStack) error
	ContainsItem(ItemStack) bool
	Empty() bool
	Clear() []ItemStack
	Size() int
}

// Possibly add all container methods that forward them but unsure
type Inventory struct {
	handle *InventoryHandle

	size       int
	customName string
	container  Container
	holder     InventoryHolder
}

func (inv Inventory) Size() int {
	return inv.size
}

func (inv Inventory) Holder() InventoryHolder {
	return inv.holder
}

func (inv *Inventory) SetContents(container Container) {
	inv.container = container
}

func (inv *Inventory) Container() Container {
	return inv.container
}

func (h Inventory) Handle() *handles.Handle {
	return (*handles.Handle)(h.handle)
}

func (h *Inventory) SetHandle(handle *handles.Handle) {
	h.handle = (*InventoryHandle)(handle)
}
