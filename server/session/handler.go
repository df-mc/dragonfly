package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// packetHandler represents a type that is able to handle a specific type of incoming packets from the client.
type packetHandler interface {
	// Handle handles an incoming packet from the client. The session of the client is passed to the packetHandler.
	// Handle returns an error if the packet was in any way invalid.
	Handle(p packet.Packet, s *Session) error
}

// Handler handles events that are called by a session. Implementations of Handler may be used to listen to
// specific events such as when a session sends a packet
type Handler interface {
	// HandleClientPacket handles packets sent to Client
	HandleClientPacket(ctx *event.Context, packet packet.Packet)
	// HandleServerPacket handles packets sent to Server
	HandleServerPacket(ctx *event.Context, packet packet.Packet)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default Handler of sessions is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = NopHandler{}

func (NopHandler) HandleClientPacket(*event.Context, packet.Packet) {}
func (NopHandler) HandleServerPacket(*event.Context, packet.Packet) {}
