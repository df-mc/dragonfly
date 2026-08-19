package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// TestHoldsTrackingCompass verifies that a query is answered for every inventory the client draws, and
// refused for a handle the player holds no compass for.
func TestHoldsTrackingCompass(t *testing.T) {
	const handle int32 = 7
	compass := item.NewStack(item.Compass{TrackingHandle: handle}, 1)

	newSession := func() *Session {
		s := &Session{inv: inventory.New(36, nil), offHand: inventory.New(1, nil), ui: inventory.New(54, nil)}
		s.openedWindow.Store(inventory.New(27, nil))
		return s
	}

	for _, c := range []struct {
		name string
		put  func(s *Session)
	}{
		{"main inventory", func(s *Session) { _ = s.inv.SetItem(0, compass) }},
		{"off hand", func(s *Session) { _ = s.offHand.SetItem(0, compass) }},
		{"cursor or crafting grid", func(s *Session) { _ = s.ui.SetItem(0, compass) }},
		{"opened container", func(s *Session) { _ = s.openedWindow.Load().SetItem(0, compass) }},
		{"opened ender chest", func(s *Session) {
			// An ender chest is stored as the opened window while it is open.
			s.enderChest = inventory.New(27, nil)
			s.openedWindow.Store(s.enderChest)
			_ = s.enderChest.SetItem(0, compass)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := newSession()
			c.put(s)
			if !s.holdsTrackingCompass(handle) {
				t.Errorf("a compass in the %v should allow the query", c.name)
			}
		})
	}

	t.Run("no compass", func(t *testing.T) {
		s := newSession()
		_ = s.inv.SetItem(0, item.NewStack(item.Compass{TrackingHandle: handle + 1}, 1))
		if s.holdsTrackingCompass(handle) {
			t.Error("a handle the player holds no compass for must be refused")
		}
	})

	t.Run("closed ender chest", func(t *testing.T) {
		s := newSession()
		s.enderChest = inventory.New(27, nil)
		_ = s.enderChest.SetItem(0, compass)
		if s.holdsTrackingCompass(handle) {
			t.Error("a compass in a closed ender chest is not drawn and must not authorise the query")
		}
	})
}
