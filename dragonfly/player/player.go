package player

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/dragonfly-tech/dragonfly/dragonfly/session"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"net"
	"strings"
	"sync"
)

// Player is an implementation of a player entity. It has methods that implement the behaviour that players
// need to play in the world.
type Player struct {
	name string

	// s holds the session of the player. This field should not be used directly, but instead,
	// Player.session() should be called.
	s            *session.Session
	sessionMutex sync.RWMutex
}

// New returns a new initialised player.
func New(name string) *Player {
	return &Player{name: name}
}

// NewWithSession returns a new player for a network session, so that the network session can control the
// player.
func NewWithSession(name string, s *session.Session) *Player {
	p := New(name)
	p.s = s
	chat.Global.Subscribe(p)
	return p
}

// Name returns the username of the player. If the player is controlled by a client, it is the username of
// the client. (Typically the XBOX Live name)
func (p *Player) Name() string {
	return p.name
}

// Message sends a formatted message to the player. The message is formatted following the rules of
// fmt.Sprintln, however the newline at the end is not written.
func (p *Player) Message(a ...interface{}) {
	p.session().SendMessage(format(a))
}

// SendPopup sends a formatted popup to the player. The popup is shown above the hotbar of the player and
// overwrites/is overwritten by the name of the item equipped.
// The popup is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) SendPopup(a ...interface{}) {
	p.session().SendPopup(format(a))
}

// SendTip sends a tip to the player. The tip is shown in the middle of the screen of the player.
// The tip is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) SendTip(a ...interface{}) {
	p.session().SendTip(format(a))
}

// Disconnect closes the player and removes it from the world.
// Disconnect, unlike Close, allows a custom message to be passed to show to the player when it is
// disconnected. The message is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) Disconnect(a ...interface{}) {
	p.close()
	p.session().Disconnect(format(a))
}

// Transfer transfers the player to a server at the address passed. If the address could not be resolved, an
// error is returned. If it is returned, the player is closed and transferred to the server.
func (p *Player) Transfer(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	p.session().Transfer(addr.IP, addr.Port)
	return p.session().Close()
}

// SendCommandOutput sends the output of a command to the player.
func (p *Player) SendCommandOutput(output *cmd.Output) {
	p.session().SendCommandOutput(output)
}

// Close closes the player and removes it from the world.
// Close disconnects the player with a 'Player closed.' message. Disconnect should be used to disconnect a
// player with a custom message.
func (p *Player) Close() error {
	p.close()
	p.session().Disconnect("Player closed.")
	return nil
}

// close closed the player without disconnecting it. It executes code shared by both the closing and the
// disconnecting of players.
func (p *Player) close() {
	chat.Global.Unsubscribe(p)
}

// session returns the network session of the player. If it has one, it is returned. If not, a no-op session
// is returned.
func (p *Player) session() *session.Session {
	return p.s
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []interface{}) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}
