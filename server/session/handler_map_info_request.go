package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MapInfoRequestHandler handles the MapInfoRequest packet.
type MapInfoRequestHandler struct {
}

// Handle ...
func (h *MapInfoRequestHandler) Handle(p packet.Packet, s *Session) error {
	var (
		pk      = p.(*packet.MapInfoRequest)
		mapItem item.MapItem
		ok      bool
	)

	for _, inv := range []*inventory.Inventory{
		s.inv,
		s.offHand,
		s.ui,
		s.armour.Inventory(),
	} {
		if inv.ContainsItemFunc(1, func(stack item.Stack) bool {
			if mapItem, ok = stack.Item().(item.MapItem); ok {
				return mapItem.BaseMap().MapIDEquals(pk.MapID)
			}

			return false // Item is not map.
		}) {
			break
		}

		return fmt.Errorf("client requests info of map %v while he does not have the corresponding map item in inventory, off hand inventory, UI inventory or armour inventory", pk.MapID)
	}

	mapItem.BaseMap().AddViewer(s)
	s.ViewMapDataChange(packet.MapUpdateFlagInitialisation, pk.MapID, world.MapPixelsChunk{}, mapItem.BaseMap().ViewableMapData)

	return nil
}
