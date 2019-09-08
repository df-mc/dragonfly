package session

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sirupsen/logrus"
)

// Nop holds a Session that does not do anything when attempting to send packets to it. The connection is not
// initialised.
var Nop = &Session{
	log:  logrus.New(),
	conn: &minecraft.Conn{},
}

// Session handles incoming packets from connections and sends outgoing packets by providing a thin layer
// of abstraction over direct packets. A Session basically 'controls' an entity.
type Session struct {
	log *logrus.Logger

	c    Controllable
	conn *minecraft.Conn
}

// New returns a new session using a controllable entity. The session will control this entity using the
// packets that it receives.
// New takes the connection from which to accept packets. It will continue handling this connection until it
// is closed either by the client or by a server side disconnect.
func New(c Controllable, conn *minecraft.Conn, log *logrus.Logger) *Session {
	s := &Session{c: c, conn: conn, log: log}
	go s.handlePackets()
	return s
}

// handlePackets continuously handles incoming packets from the connection. It processes them accordingly.
// Once the connection is closed, handlePackets will return.
func (s *Session) handlePackets() {
	defer func() {
		_ = s.c.Close()
		_ = s.conn.Close()
		s.c = nil
	}()
	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			return
		}
		if err := s.handlePacket(pk); err != nil {
			// An error occurred during the handling of a packet. Print the error and stop handling any more
			// packets.
			s.log.Errorf("error processing packet from %v: %v\n", s.conn.RemoteAddr(), err)
			return
		}
	}
}

// handlePacket handles an incoming packet, processing it accordingly. If the packet had invalid data or was
// otherwise not valid in its context, an error is returned.
func (s *Session) handlePacket(pk packet.Packet) error {
	switch pk := pk.(type) {
	case *packet.Text:
		return s.handleText(pk)
	default:
		s.log.Debugf("unhandled packet %T%v from %v\n", pk, fmt.Sprintf("%+v", pk)[1:], s.conn.RemoteAddr())
	}
	return nil
}

// writePacket writes a packet to the connection.
func (s *Session) writePacket(pk packet.Packet) error {
	if s == Nop {
		return nil
	}
	return s.conn.WritePacket(pk)
}

// handleText ...
func (s *Session) handleText(pk *packet.Text) error {
	if pk.TextType != packet.TextTypeChat {
		return fmt.Errorf("text packet can only contain text type of type chat (%v) but got %v", packet.TextTypeChat, pk.TextType)
	}
	if pk.SourceName != s.conn.IdentityData().DisplayName {
		return fmt.Errorf("text packet source name must be equal to display name")
	}
	_, _ = chat.Global.Printf("<%v> %v\n", s.conn.IdentityData().DisplayName, pk.Message)
	return nil
}

// SendMessage ...
func (s *Session) SendMessage(message string) {
	_ = s.conn.WritePacket(&packet.Text{
		TextType: packet.TextTypeRaw,
		Message:  message,
	})
}
