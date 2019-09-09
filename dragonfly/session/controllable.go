package session

import (
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"io"
)

// Controllable represents an entity that may be controlled by a Session. Generally, a Controllable is
// implemented in the form of a Player.
// Methods in Controllable will be added as Session needs them in order to handle packets.
type Controllable interface {
	io.Closer
	cmd.Source
}
