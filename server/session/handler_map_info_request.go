package session

import (
	"fmt"
	"sync"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MapInfoRequestHandler handles the MapInfoRequest packet.
type MapInfoRequestHandler struct {
	mu sync.Mutex
}

// Handle ...
func (h *MapInfoRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.MapInfoRequest)

	for _, inv := range []*inventory.Inventory{
		s.inv,
		s.offHand,
		s.ui,
		s.armour.Inventory(),
	} {
		if inv.ContainsItemFunc(1, func(stack item.Stack) bool {
			if mapItem, ok := stack.Item().(item.MapInterface); ok {
				return mapItem.GetMapID() == pk.MapID
			}

			return false // Item is not map.
		}) {
			break
		}

		return fmt.Errorf("client requests info of map %v while he does not have the corresponding map item in inventory, off hand inventory, UI inventory or armour inventory", pk.MapID)
	}

	return nil
}
