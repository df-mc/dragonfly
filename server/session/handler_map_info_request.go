package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/item"
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
		ok      bool
		mapItem item.MapItem
	)
	if mapItem, ok = s.canAccessMapData(pk.MapID); !ok {
		return fmt.Errorf("client requests info of map %v while he does not have the corresponding map item in inventory, off hand inventory, UI inventory or armour inventory", pk.MapID)
	}

	mapItem.BaseMap().AddViewer(s)
	s.SendMapData(packet.MapUpdateFlagInitialisation, pk.MapID, world.MapPixelsChunk{}, mapItem.BaseMap().ViewableMapData)

	return nil
}
