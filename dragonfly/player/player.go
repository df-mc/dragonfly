package player

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/dragonfly-tech/dragonfly/dragonfly/session"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"strings"
	"sync"
)

// Player is an implementation of a player entity. It has methods that implement the behaviour that players
// need to play in the world.
type Player struct {
	// s holds the session of the player. This field should not be used directly, but instead,
	// Player.session() should be called.
	s            *session.Session
	sessionMutex sync.RWMutex
}

// New returns a new initialised player.
func New() *Player {
	return &Player{}
}

// NewWithSession returns a new player for a network session, so that the network session can control the
// player.
func NewWithSession(s *session.Session) *Player {
	p := New()
	p.s = s
	chat.Global.Subscribe(p)
	return p
}

// Message sends a formatted message to the player. The message is formatted following the rules of
// fmt.Sprintln, however the newline at the end is not written.
func (p *Player) Message(a ...interface{}) {
	// Remove at most two trailing newlines from the string.
	s := strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
	p.session().SendMessage(s)
}

// SendCommandOutput sends the output of a command to the player.
func (p *Player) SendCommandOutput(output *cmd.Output) {
	p.session().SendCommandOutput(output)
}

// Close closes the player, removing any references that would otherwise keep the player from being garbage
// collected, and removes the player from the world.
func (p *Player) Close() error {
	chat.Global.Unsubscribe(p)

	p.sessionMutex.Lock()
	p.s = nil
	p.sessionMutex.Unlock()

	return nil
}

// session returns the network session of the player. If it has one, it is returned. If not, a no-op session
// is returned.
func (p *Player) session() *session.Session {
	p.sessionMutex.RLock()
	defer p.sessionMutex.RUnlock()

	if p.s != nil {
		return p.s
	}
	return session.Nop
}
