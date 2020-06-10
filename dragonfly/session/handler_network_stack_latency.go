package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NetworkStackLatencyHandler handles the NetworkStackLatency packet.
type NetworkStackLatencyHandler struct{}

// Handle ...
func (*NetworkStackLatencyHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.NetworkStackLatency)

	s.pingMu.Lock()
	defer s.pingMu.Unlock()
	c, ok := s.pings[pk.Timestamp]
	if !ok {
		return fmt.Errorf("invalid Timestamp: did not send a request with this timestamp")
	}
	delete(s.pings, pk.Timestamp)
	c <- struct{}{}
	return nil
}
