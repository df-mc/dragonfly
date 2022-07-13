package inventory

import "github.com/df-mc/dragonfly/server/world"

// FakeInventoryType represents a type of fake inventory, such as a HopperFakeInventory or a ChestFakeInventory.
type FakeInventoryType struct {
	fakeInventoryType
}

// HopperFakeInventory returns a FakeInventoryType that represents a hopper.
func HopperFakeInventory() FakeInventoryType {
	return FakeInventoryType{0}
}

// DispenserFakeInventory returns a FakeInventoryType that represents a dispenser.
func DispenserFakeInventory() FakeInventoryType {
	return FakeInventoryType{1}
}

// ChestFakeInventory returns a FakeInventoryType that represents a chest.
func ChestFakeInventory() FakeInventoryType {
	return FakeInventoryType{2}
}

// DoubleChestFakeInventory returns a FakeInventoryType that represents a double chest.
func DoubleChestFakeInventory() FakeInventoryType {
	return FakeInventoryType{3}
}

type fakeInventoryType uint8

// Uint8 ...
func (f fakeInventoryType) Uint8() uint8 {
	return uint8(f)
}

// Block returns the world.Block that represents the fake inventory.
func (f fakeInventoryType) Block() world.Block {
	switch f {
	case 0:
		b, _ := world.BlockByName("minecraft:hopper", map[string]any{"facing_direction": int32(0), "toggle_bit": uint8(0)})
		return b
	case 1:
		b, _ := world.BlockByName("minecraft:dispenser", map[string]any{"facing_direction": int32(0), "triggered_bit": uint8(0)})
		return b
	case 2, 3:
		b, _ := world.BlockByName("minecraft:chest", map[string]any{"facing_direction": int32(0)})
		return b
	}
	panic("should never happen")
}

// Size returns the size of the fake inventory.
func (f fakeInventoryType) Size() int {
	switch f {
	case 0:
		return 5
	case 1:
		return 9
	case 2:
		return 27
	case 3:
		return 54
	}
	panic("should never happen")
}
