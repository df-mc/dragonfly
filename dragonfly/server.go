package dragonfly

import (
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/df-mc/dragonfly/dragonfly/item" // Imported for compiler directives.
	"github.com/df-mc/dragonfly/dragonfly/player"
	"github.com/df-mc/dragonfly/dragonfly/player/skin"
	"github.com/df-mc/dragonfly/dragonfly/session"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/generator"
	"github.com/df-mc/dragonfly/dragonfly/world/mcdb"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	_ "unsafe" // Imported for compiler directives.
)

// Server implements a Dragonfly server. It runs the main server loop and handles the connections of players
// trying to join the server.
type Server struct {
	started atomic.Bool
	name    atomic.String

	c        Config
	log      *logrus.Logger
	listener *minecraft.Listener
	world    *world.World
	players  chan *player.Player

	startTime time.Time

	playerMutex sync.RWMutex
	// p holds a map of all players currently connected to the server. When they leave, they are removed from
	// the map.
	p map[uuid.UUID]*player.Player
}

// New returns a new server using the Config passed. If nil is passed, a default configuration is returned.
// (A call to dragonfly.DefaultConfig().)
// The Logger passed will be used to log errors and information to. If nil is passed, a default Logger is
// used by calling logrus.New().
// Note that no two servers should be active at the same time. Doing so anyway will result in unexpected
// behaviour.
func New(c *Config, log *logrus.Logger) *Server {
	if log == nil {
		log = logrus.New()
	}
	if c == nil {
		conf := DefaultConfig()
		c = &conf
	}
	s := &Server{
		c:       *c,
		log:     log,
		players: make(chan *player.Player),
		world:   world.New(log, c.World.SimulationDistance),
		p:       make(map[uuid.UUID]*player.Player),
		name:    *atomic.NewString(c.Server.Name),
	}
	return s
}

// Accept accepts an incoming player into the server. It blocks until a player connects to the server.
// Accept returns an error if the Server is closed using a call to Close.
func (server *Server) Accept() (*player.Player, error) {
	p, ok := <-server.players
	if !ok {
		return nil, errors.New("server closed")
	}
	server.playerMutex.Lock()
	server.p[p.UUID()] = p
	server.playerMutex.Unlock()

	return p, nil
}

// World returns the world of the server. Players will be spawned in this world and this world will be read
// from and written to when the world is edited.
func (server *Server) World() *world.World {
	return server.world
}

// Run runs the server and blocks until it is closed using a call to Close(). When called, the server will
// accept incoming connections. Run will block the current goroutine until the server is stopped. To start
// the server on a different goroutine, use (*Server).Start() instead.
// After a call to Run, calls to Server.Accept() may be made to accept players into the server.
func (server *Server) Run() error {
	if !server.started.CAS(false, true) {
		panic("server already running")
	}

	server.log.Info("Starting server...")
	server.loadWorld()
	server.World().Generator(generator.Flat{})
	if err := server.startListening(); err != nil {
		return err
	}
	item_registerVanillaCreativeItems()
	world_registerAllStates()
	server.run()
	return nil
}

// Start runs the server but does not block, unlike Run, but instead accepts connections on a different
// goroutine. Connections will be accepted until the listener is closed using a call to Close.
// One started, players may be accepted using Server.Accept().
func (server *Server) Start() error {
	if !server.started.CAS(false, true) {
		panic("server already running")
	}

	server.log.Info("Starting server...")
	server.loadWorld()
	server.World().Generator(generator.Flat{})
	if err := server.startListening(); err != nil {
		return err
	}
	item_registerVanillaCreativeItems()
	world_registerAllStates()
	go server.run()
	return nil
}

// Uptime returns the duration that the server has been running for. Measurement starts the moment a call to
// Server.Start or Server.Run is made.
func (server *Server) Uptime() time.Duration {
	if !server.running() {
		return 0
	}
	return time.Since(server.startTime)
}

// PlayerCount returns the current player count of the server. It is equivalent to calling
// len(server.Players()).
func (server *Server) PlayerCount() int {
	server.playerMutex.RLock()
	defer server.playerMutex.RUnlock()

	return len(server.p)
}

// MaxPlayerCount returns the maximum amount of players that are allowed to play on the server at the same
// time. Players trying to join when the server is full will be refused to enter.
// If the config has a maximum player count set to 0, MaxPlayerCount will return Server.PlayerCount + 1.
func (server *Server) MaxPlayerCount() int {
	if server.c.Server.MaximumPlayers == 0 {
		return server.PlayerCount() + 1
	}
	return server.c.Server.MaximumPlayers
}

// Players returns a list of all players currently connected to the server. Note that the slice returned is
// not updated when new players join or leave, so it is only valid for as long as no new players join or
// players leave.
func (server *Server) Players() []*player.Player {
	server.playerMutex.RLock()
	defer server.playerMutex.RUnlock()

	players := make([]*player.Player, 0, len(server.p))
	for _, p := range server.p {
		players = append(players, p)
	}
	return players
}

// Player looks for a player on the server with the UUID passed. If found, the player is returned and the bool
// returns holds a true value. If not, the bool returned is false and the player is nil.
func (server *Server) Player(uuid uuid.UUID) (*player.Player, bool) {
	server.playerMutex.RLock()
	defer server.playerMutex.RUnlock()

	if p, ok := server.p[uuid]; ok {
		return p, true
	}
	return nil, false
}

// PlayerByName looks for a player on the server with the name passed. If found, the player is returned and the bool
// returns holds a true value. If not, the bool is false and the player is nil
func (server *Server) PlayerByName(name string) (*player.Player, bool) {
	for _, p := range server.Players() {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

// SetNamef sets the name of the Server, also known as the MOTD. This name is displayed in the server list.
// The formatting of the name passed follows the rules of fmt.Sprintf.
func (server *Server) SetNamef(format string, a ...interface{}) {
	server.name.Store(fmt.Sprintf(format, a...))
}

// SetName sets the name of the Server, also known as the MOTD. This name is displayed in the server list.
// The formatting of the name passed follows the rules of fmt.Sprint.
func (server *Server) SetName(a ...interface{}) {
	server.name.Store(fmt.Sprint(a...))
}

// Close closes the server, making any call to Run/Accept cancel immediately.
func (server *Server) Close() error {
	if !server.running() {
		panic("server not yet running")
	}

	server.log.Info("Server shutting down...")
	defer server.log.Info("Server stopped.")

	server.log.Debug("Disconnecting players...")
	server.playerMutex.RLock()
	for _, p := range server.p {
		p.Disconnect(text.Yellow()(server.c.Server.ShutdownMessage))
	}
	server.playerMutex.RUnlock()

	server.log.Debug("Closing world...")
	if err := server.world.Close(); err != nil {
		return err
	}

	server.log.Debug("Closing listener...")
	return server.listener.Close()
}

// CloseOnProgramEnd closes the server right before the program ends, so that all data of the server are
// saved properly.
func (server *Server) CloseOnProgramEnd() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		if err := server.Close(); err != nil {
			server.log.Errorf("error shutting down server: %v", err)
		}
	}()
}

// running checks if the server is currently running.
func (server *Server) running() bool {
	return server.started.Load()
}

// startListening starts making the EncodeBlock listener listen, accepting new connections from players.
func (server *Server) startListening() error {
	server.startTime = time.Now()

	w := server.log.Writer()
	defer func() {
		_ = w.Close()
	}()

	server.listener = &minecraft.Listener{
		// We wrap a log.Logger around our Logrus logger so that it will print in the same format as the
		// normal Logrus logger would.
		ErrorLog:               log.New(w, "", 0),
		ServerName:             server.c.Server.Name,
		MaximumPlayers:         server.c.Server.MaximumPlayers,
		AuthenticationDisabled: !server.c.Server.AuthEnabled,
	}
	//noinspection SpellCheckingInspection
	if err := server.listener.Listen("raknet", server.c.Network.Address); err != nil {
		return fmt.Errorf("listening on address failed: %w", err)
	}
	server.listener.StatusProvider(statusProvider{s: server})

	server.log.Infof("Server running on %v.\n", server.listener.Addr())
	return nil
}

// run runs the server, continuously accepting new connections from players. It returns when the server is
// closed by a call to Close.
func (server *Server) run() {
	for {
		c, err := server.listener.Accept()
		if err != nil {
			// Accept will only return an error if the Listener was closed, meaning trying to continue
			// listening is futile.
			close(server.players)
			return
		}
		go server.handleConn(c.(*minecraft.Conn))
	}
}

// handleConn handles an incoming connection accepted from the Listener.
func (server *Server) handleConn(conn *minecraft.Conn) {
	//noinspection SpellCheckingInspection
	data := minecraft.GameData{
		Yaw:            90,
		WorldName:      server.c.World.Name,
		Blocks:         server.blockEntries(),
		PlayerPosition: vec64To32(server.world.Spawn().Vec3Centre()),
		PlayerGameMode: 1,
		// We set these IDs to 1, because that's how the session will treat them.
		EntityUniqueID:               1,
		EntityRuntimeID:              1,
		Time:                         int64(server.world.Time()),
		GameRules:                    map[string]interface{}{"naturalregeneration": false},
		Difficulty:                   2,
		ServerAuthoritativeMovement:  true,
		ServerAuthoritativeInventory: true,
	}
	if err := conn.StartGame(data); err != nil {
		_ = server.listener.Disconnect(conn, "Connection timeout.")
		server.log.Debugf("connection %v failed spawning: %v\n", conn.RemoteAddr(), err)
		return
	}
	id, err := uuid.Parse(conn.IdentityData().Identity)
	if err != nil {
		_ = conn.Close()
		server.log.Warnf("connection %v has a malformed UUID ('%v')\n", conn.RemoteAddr(), id)
		return
	}
	if p, ok := server.Player(id); ok {
		p.Disconnect("Logged in from another location.")
	}
	server.players <- server.createPlayer(id, conn)
}

// handleSessionClose handles the closing of a session. It removes the player of the session from the server.
func (server *Server) handleSessionClose(controllable session.Controllable) {
	server.playerMutex.Lock()
	delete(server.p, controllable.UUID())
	server.playerMutex.Unlock()
}

// createPlayer creates a new player instance using the UUID and connection passed.
func (server *Server) createPlayer(id uuid.UUID, conn *minecraft.Conn) *player.Player {
	s := session.New(conn, server.c.World.MaximumChunkRadius, server.log)
	p := player.NewWithSession(conn.IdentityData().DisplayName, conn.IdentityData().XUID, id, server.createSkin(conn.ClientData()), s, server.world.Spawn().Vec3Middle())
	s.Start(p, server.world, server.handleSessionClose)

	return p
}

// loadWorld loads the world of the server, ending the program if the world could not be loaded.
func (server *Server) loadWorld() {
	server.log.Debug("Loading world...")
	p, err := mcdb.New(server.c.World.Folder)
	if err != nil {
		server.log.Fatalf("error loading world: %v", err)
	}
	server.world.Provider(p)
	server.log.Debugf("Loaded world '%v'.", server.world.Name())
}

// createSkin creates a new skin using the skin data found in the client data in the login, and returns it.
func (server *Server) createSkin(data login.ClientData) skin.Skin {
	// gopher tunnel guarantees the following values are valid data and are of the correct size.
	skinData, _ := base64.StdEncoding.DecodeString(data.SkinData)
	modelData, _ := base64.StdEncoding.DecodeString(data.SkinGeometry)
	skinResourcePatch, _ := base64.StdEncoding.DecodeString(data.SkinResourcePatch)
	modelConfig, _ := skin.DecodeModelConfig(skinResourcePatch)

	playerSkin := skin.New(data.SkinImageWidth, data.SkinImageHeight)
	playerSkin.Persona = data.PersonaSkin
	playerSkin.Pix = skinData
	playerSkin.Model = modelData
	playerSkin.ModelConfig = modelConfig

	for _, animation := range data.AnimatedImageData {
		var t skin.AnimationType
		switch animation.Type {
		case protocol.SkinAnimationHead:
			t = skin.AnimationHead
		case protocol.SkinAnimationBody32x32:
			t = skin.AnimationBody32x32
		case protocol.SkinAnimationBody128x128:
			t = skin.AnimationBody128x128
		}

		anim := skin.NewAnimation(animation.ImageWidth, animation.ImageHeight, t)
		anim.FrameCount = int(animation.Frames)
		anim.Pix, _ = base64.StdEncoding.DecodeString(animation.Image)

		playerSkin.Animations = append(playerSkin.Animations, anim)
	}

	return playerSkin
}

// blockEntries loads a list of all block state entries of the server, ready to be sent in the StartGame
// packet.
func (server *Server) blockEntries() (entries []interface{}) {
	for _, b := range world_allBlocks() {
		name, properties := b.EncodeBlock()
		entries = append(entries, map[string]interface{}{
			"block": map[string]interface{}{
				"version": protocol.CurrentBlockVersion,
				"name":    name,
				"states":  properties,
			},
		})
	}
	return
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

//go:linkname world_registerAllStates github.com/df-mc/dragonfly/dragonfly/world.registerAllStates
//noinspection ALL
func world_registerAllStates()

//go:linkname item_registerVanillaCreativeItems github.com/df-mc/dragonfly/dragonfly/item.registerVanillaCreativeItems
//noinspection ALL
func item_registerVanillaCreativeItems()

//go:linkname world_allBlocks github.com/df-mc/dragonfly/dragonfly/world.allBlocks
//noinspection ALL
func world_allBlocks() []world.Block
