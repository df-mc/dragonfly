package nbtconv

import "github.com/df-mc/dragonfly/server/item"

const (
	itemLockKey       = "minecraft:item_lock"
	legacyItemLockKey = "item_lock"
	itemLockModeKey   = "mode"
)

// readItemLock reads the Bedrock item_lock component from the item NBT and saves it to the stack.
func readItemLock(m map[string]any, s *item.Stack) {
	mode, ok := lockModeFromMap(m)
	if !ok {
		return
	}
	*s = s.WithLockMode(mode)
}

// writeItemLock writes the Bedrock item_lock component to the item NBT if one is present on the stack.
func writeItemLock(m map[string]any, s item.Stack) {
	if !s.Locked() {
		return
	}
	v := s.LockMode().LegacyValue()
	m[itemLockKey] = v
}

func lockModeFromMap(m map[string]any) (item.LockMode, bool) {
	if mode, ok := lockModeFromValue(m[itemLockKey]); ok {
		return mode, true
	}
	return lockModeFromValue(m[legacyItemLockKey])
}

func lockModeFromValue(v any) (item.LockMode, bool) {
	switch v := v.(type) {
	case map[string]any:
		mode, _ := v[itemLockModeKey].(string)
		return item.ParseLockMode(mode)
	case uint8:
		return item.LockModeFromLegacyValue(v)
	case int8:
		if v < 0 {
			return item.NotLocked, false
		}
		return item.LockModeFromLegacyValue(uint8(v))
	case int16:
		if v < 0 || v > 255 {
			return item.NotLocked, false
		}
		return item.LockModeFromLegacyValue(uint8(v))
	case int32:
		if v < 0 || v > 255 {
			return item.NotLocked, false
		}
		return item.LockModeFromLegacyValue(uint8(v))
	case int64:
		if v < 0 || v > 255 {
			return item.NotLocked, false
		}
		return item.LockModeFromLegacyValue(uint8(v))
	case string:
		return item.ParseLockMode(v)
	default:
		return item.NotLocked, false
	}
}
