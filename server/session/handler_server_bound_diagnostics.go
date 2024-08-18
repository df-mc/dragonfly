package session

import (
	"github.com/df-mc/dragonfly/server/player/diagnostics"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerBoundDiagnosticsHandler handles diagnostic updates from the client.
type ServerBoundDiagnosticsHandler struct{}

// Handle ...
func (h *ServerBoundDiagnosticsHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ServerBoundDiagnostics)
	s.c.UpdateDiagnostics(diagnostics.Diagnostics{
		AverageFramesPerSecond:        float64(pk.AverageFramesPerSecond),
		AverageServerSimTickTime:      float64(pk.AverageServerSimTickTime),
		AverageClientSimTickTime:      float64(pk.AverageClientSimTickTime),
		AverageBeginFrameTime:         float64(pk.AverageBeginFrameTime),
		AverageInputTime:              float64(pk.AverageInputTime),
		AverageRenderTime:             float64(pk.AverageRenderTime),
		AverageEndFrameTime:           float64(pk.AverageEndFrameTime),
		AverageRemainderTimePercent:   float64(pk.AverageRemainderTimePercent),
		AverageUnaccountedTimePercent: float64(pk.AverageUnaccountedTimePercent),
	})
	return nil
}
