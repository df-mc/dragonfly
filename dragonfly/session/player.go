package session

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/entity"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
)

// handleMovePlayer ...
func (s *Session) handleMovePlayer(pk *packet.MovePlayer) error {
	if pk.EntityRuntimeID != s.conn.GameData().EntityRuntimeID {
		return fmt.Errorf("incorrect entity runtime ID %v: runtime ID must be equal to %v", pk.EntityRuntimeID, s.conn.GameData().EntityRuntimeID)
	}
	entity.Move(s.c, pk.Position.Sub(s.c.Position()))
	entity.Rotate(s.c, pk.Yaw-s.c.Yaw(), pk.Pitch-s.c.Pitch())

	s.chunkLoader.Load().(*world.Loader).Move(pk.Position)
	s.writePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pk.Position[0]), int32(pk.Position[1]), int32(pk.Position[2])},
		Radius:   uint32(s.chunkRadius * 16),
	})
	return nil
}

// Disconnect disconnects the client and ultimately closes the session. If the message passed is non-empty,
// it will be shown to the client.
func (s *Session) Disconnect(message string) {
	s.writePacket(&packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
	if s != Nop {
		_ = s.conn.Flush()
		_ = s.conn.Close()
	}
}

// Transfer transfers the player to a server with the IP and port passed.
func (s *Session) Transfer(ip net.IP, port int) {
	s.writePacket(&packet.Transfer{
		Address: ip.String(),
		Port:    uint16(port),
	})
}
