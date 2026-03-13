package item

// LockMode describes the Bedrock item_lock mode applied to an item stack.
type LockMode uint8

const (
	// NotLocked indicates no item_lock is applied to the stack.
	NotLocked LockMode = iota
	// LockInInventory prevents the stack from being removed from the player's inventory, dropped, crafted with,
	// placed in bundles, or renamed in an anvil.
	LockInInventory
	// LockInSlot prevents the stack from being moved or removed from its current slot in the player's inventory.
	LockInSlot
)

// String returns the Bedrock component mode name of the lock mode.
func (m LockMode) String() string {
	switch m {
	case LockInInventory:
		return "lock_in_inventory"
	case LockInSlot:
		return "lock_in_slot"
	default:
		return ""
	}
}

// ParseLockMode parses a Bedrock item_lock mode string.
func ParseLockMode(mode string) (LockMode, bool) {
	switch mode {
	case "lock_in_inventory":
		return LockInInventory, true
	case "lock_in_slot":
		return LockInSlot, true
	default:
		return NotLocked, false
	}
}

// LockModeFromLegacyValue parses Bedrock's legacy item_lock byte value.
func LockModeFromLegacyValue(v uint8) (LockMode, bool) {
	switch v {
	case 1:
		return LockInSlot, true
	case 2:
		return LockInInventory, true
	default:
		return NotLocked, false
	}
}

// LegacyValue returns the Bedrock item NBT byte used to represent the lock mode.
func (m LockMode) LegacyValue() uint8 {
	switch m {
	case LockInSlot:
		return 1
	case LockInInventory:
		return 2
	default:
		return 0
	}
}
