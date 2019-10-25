package player

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/event"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/bossbar"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/chat"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/scoreboard"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/skin"
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/title"
	"github.com/dragonfly-tech/dragonfly/dragonfly/session"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/cmd"
	"net"
	"strings"
	"sync"
)

// Player is an implementation of a player entity. It has methods that implement the behaviour that players
// need to play in the world.
type Player struct {
	world.Pos
	name string
	uuid uuid.UUID
	xuid string

	world *world.World

	skin skin.Skin

	sMutex sync.RWMutex
	// s holds the session of the player. This field should not be used directly, but instead,
	// Player.session() should be called.
	s *session.Session

	hMutex sync.RWMutex
	// h holds the current handler of the player. It may be changed at any time by calling the Start method.
	h Handler
}

// New returns a new initialised player. A random UUID is generated for the player, so that it may be
// identified over network.
func New(name string, skin skin.Skin) *Player {
	// TODO: Add a way to change the world of a player once created without session. Might need a function like entity.Move for re-usability.
	return &Player{name: name, h: NopHandler{}, uuid: uuid.New(), skin: skin}
}

// NewWithSession returns a new player for a network session, so that the network session can control the
// player.
// A set of additional fields must be provided to initialise the player with the client's data, such as the
// name and the skin of the player.
func NewWithSession(name, xuid string, uuid uuid.UUID, skin skin.Skin, s *session.Session, w *world.World) *Player {
	p := New(name, skin)
	p.s = s
	p.uuid = uuid
	p.xuid = xuid
	p.skin = skin
	p.world = w

	chat.Global.Subscribe(p)
	return p
}

// Name returns the username of the player. If the player is controlled by a client, it is the username of
// the client. (Typically the XBOX Live name)
func (p *Player) Name() string {
	return p.name
}

// UUID returns the UUID of the player. This UUID will remain consistent with an XBOX Live account, and will,
// unlike the name of the player, never change.
// It is therefore recommended to use the UUID over the name of the player. Additionally, it is recommended to
// use the UUID over the XUID because of its standard format.
func (p *Player) UUID() uuid.UUID {
	return p.uuid
}

// XUID returns the XBOX Live user ID of the player. It will remain consistent with the XBOX Live account,
// and will not change in the lifetime of an account.
// The XUID is a number that can be parsed as an int64. No more information on what it represents is
// available, and the UUID should be preferred.
// The XUID returned is empty if the Player is not connected to a network session.
func (p *Player) XUID() string {
	return p.xuid
}

// Skin returns the skin that a player joined with. This skin will be visible to other players that the player
// is shown to.
// If the player was not connected to a network session, a default skin will be set.
func (p *Player) Skin() skin.Skin {
	return p.skin
}

// World returns the world that the player is currently in.
func (p *Player) World() *world.World {
	return p.world
}

// Handle changes the current handler of the player. As a result, events called by the player will call
// handlers of the Handler passed.
// Handle sets the player's handler to NopHandler if nil is passed.
func (p *Player) Handle(h Handler) {
	p.hMutex.Lock()
	defer p.hMutex.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	p.h = h
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

// SendTitle sends a title to the player. The title may be configured to change the duration it is displayed
// and the text it shows.
// If non-empty, the subtitle is shown in a smaller font below the title. The same counts for the action text
// of the title, which is shown in a font similar to that of a tip/popup.
func (p *Player) SendTitle(t *title.Title) {
	p.session().SetTitleDurations(t.FadeInDuration(), t.Duration(), t.FadeOutDuration())
	p.session().SendTitle(t.Text())
	if t.Subtitle() != "" {
		p.session().SendSubtitle(t.Subtitle())
	}
	if t.ActionText() != "" {
		p.session().SendActionBarMessage(t.ActionText())
	}
}

// SendScoreboard sends a scoreboard to the player. The scoreboard will be present indefinitely until removed
// by the caller.
// SendScoreboard may be called at any time to change the scoreboard of the player.
func (p *Player) SendScoreboard(scoreboard *scoreboard.Scoreboard) {
	p.session().SendScoreboard(scoreboard.Name())
	p.session().SendScoreboardLines(scoreboard.Lines())
}

// RemoveScoreboard removes any scoreboard currently present on the screen of the player. Nothing happens if
// the player has no scoreboard currently active.
func (p *Player) RemoveScoreboard() {
	p.session().RemoveScoreboard()
}

// SendBossBar sends a boss bar to the player, so that it will be shown indefinitely at the top of the
// player's screen.
// The boss bar may be removed by calling Player.RemoveBossBar().
func (p *Player) SendBossBar(bar *bossbar.BossBar) {
	p.session().SendBossBar(bar.Text(), bar.HealthPercentage())
}

// RemoveBossBar removes any boss bar currently active on the player's screen. If no boss bar is currently
// present, nothing happens.
func (p *Player) RemoveBossBar() {
	p.session().RemoveBossBar()
}

// Chat writes a message in the global chat (chat.Global). The message is prefixed with the name of the
// player.
func (p *Player) Chat(message string) {
	ctx := event.C()
	p.handler().HandleChat(ctx, message)

	ctx.Continue(func() {
		chat.Global.Printf("<%v> %v\n", p.name, message)
	})
}

// ExecuteCommand executes a command passed as the player. If the command could not be found, or if the usage
// was incorrect, an error message is sent to the player.
func (p *Player) ExecuteCommand(commandLine string) {
	args := strings.Split(commandLine, " ")
	commandName := strings.TrimPrefix(args[0], "/")

	command, ok := cmd.CommandByAlias(commandName)
	if !ok {
		output := &cmd.Output{}
		output.Errorf("Unknown command '%v'", commandName)
		p.SendCommandOutput(output)
		return
	}

	ctx := event.C()
	p.handler().HandleCommandExecution(ctx, command, args[1:])
	ctx.Continue(func() {
		command.Execute(strings.TrimPrefix(commandLine, "/"+commandName+" "), p)
	})
}

// Disconnect closes the player and removes it from the world.
// Disconnect, unlike Close, allows a custom message to be passed to show to the player when it is
// disconnected. The message is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) Disconnect(a ...interface{}) {
	p.session().Disconnect(format(a))
	p.close()
}

// Transfer transfers the player to a server at the address passed. If the address could not be resolved, an
// error is returned. If it is returned, the player is closed and transferred to the server.
func (p *Player) Transfer(address string) (err error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	ctx := event.C()
	p.handler().HandleTransfer(ctx, addr)

	ctx.Continue(func() {
		p.session().Transfer(addr.IP, addr.Port)
		err = p.Close()
	})
	return
}

// SendCommandOutput sends the output of a command to the player.
func (p *Player) SendCommandOutput(output *cmd.Output) {
	p.session().SendCommandOutput(output)
}

// Close closes the player and removes it from the world.
// Close disconnects the player with a 'Player closed.' message. Disconnect should be used to disconnect a
// player with a custom message.
func (p *Player) Close() error {
	p.session().Disconnect("Connection closed.")
	p.close()
	return nil
}

// close closed the player without disconnecting it. It executes code shared by both the closing and the
// disconnecting of players.
func (p *Player) close() {
	p.handler().HandleClose()
	p.Handle(NopHandler{})
	chat.Global.Unsubscribe(p)

	p.sMutex.Lock()
	p.s = nil
	p.sMutex.Unlock()
}

// session returns the network session of the player. If it has one, it is returned. If not, a no-op session
// is returned.
func (p *Player) session() *session.Session {
	p.sMutex.RLock()
	s := p.s
	p.sMutex.RUnlock()

	if s == nil {
		return session.Nop
	}
	return s
}

// handler returns the handler of the player.
func (p *Player) handler() Handler {
	p.hMutex.RLock()
	handler := p.h
	p.hMutex.RUnlock()
	return handler
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []interface{}) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}
