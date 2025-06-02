package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// packetHandler represents a type that is able to handle a specific type of incoming packets from the client.
type packetHandler interface {
	// Handle handles an incoming packet from the client. The session of the client is passed to the packetHandler.
	// Handle returns an error if the packet was in any way invalid.
	Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error
}

// Context ...
type Context = event.Context[*Session]

// UserPacketHandler ...
type UserPacketHandler interface {
	// HandleClientPacket ...
	HandleClientPacket(ctx *Context, pk packet.Packet)
	// HandleServerPacket ...
	HandleServerPacket(ctx *Context, pk packet.Packet)
}
