package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerBoundDiagnosticsHandler handles diagnostic updates from the client.
type ServerBoundDiagnosticsHandler struct{}

// Handle ...
func (h *ServerBoundDiagnosticsHandler) Handle(p packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.ServerBoundDiagnostics)
	c.UpdateDiagnostics(Diagnostics{
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
