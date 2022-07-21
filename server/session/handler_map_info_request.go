package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MapInfoRequestHandler handles the MapInfoRequest packet.
type MapInfoRequestHandler struct {
}

// Handle ...
func (h *MapInfoRequestHandler) Handle(p packet.Packet, s *Session) error {
	var (
		pk      = p.(*packet.MapInfoRequest)
		mapItem item.MapInterface
		ok      bool
	)

	for _, inv := range []*inventory.Inventory{
		s.inv,
		s.offHand,
		s.ui,
		s.armour.Inventory(),
	} {
		if inv.ContainsItemFunc(1, func(stack item.Stack) bool {
			if mapItem, ok = stack.Item().(item.MapInterface); ok {
				return mapItem.GetMapID() == pk.MapID
			}

			return false // Item is not map.
		}) {
			break
		}

		return fmt.Errorf("client requests info of map %v while he does not have the corresponding map item in inventory, off hand inventory, UI inventory or armour inventory", pk.MapID)
	}

	var (
		data   = mapItem.GetData()
		pixels = data.Pixels
		height = int32(len(data.Pixels))
		width  int32

		trackeds []protocol.MapTrackedObject
	)
	for _, rows := range data.Pixels {
		if len(rows) > int(width) {
			width = int32(len(rows))
		}
	}
	for _, e := range data.TrackEntities {
		trackeds = append(trackeds, protocol.MapTrackedObject{
			Type:           protocol.MapObjectTypeEntity,
			EntityUniqueID: int64(s.entityRuntimeID(e)),
		})
	}
	for _, p := range data.TrackBlocks {
		trackeds = append(trackeds, protocol.MapTrackedObject{
			Type:          protocol.MapObjectTypeBlock,
			BlockPosition: protocol.BlockPos{int32(p[0]), int32(p[1]), int32(p[2])},
		})
	}

	s.writePacket(&packet.ClientBoundMapItemData{
		MapID:          pk.MapID,
		UpdateFlags:    packet.MapUpdateFlagInitialisation,
		Dimension:      byte(mapItem.GetDimension().EncodeDimension()),
		LockedMap:      false, // TODO: Locked map support
		Scale:          byte(mapItem.GetScale()),
		TrackedObjects: trackeds,
		// Decorations is a list of fixed decorations located on the map. The decorations will not change
		// client-side, unless the server updates them.
		Decorations: []protocol.MapDecoration{},

		Height: height,
		Width:  width,
		Pixels: pixels,
	})

	return nil
}
