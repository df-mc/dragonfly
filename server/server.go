package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/internal"
	_ "github.com/df-mc/dragonfly/server/item" // Imported for compiler directives.
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/generator"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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

	joinMessage, quitMessage atomic.String
	playerProvider           player.Provider

	c        Config
	log      internal.Logger
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
// (A call to server.DefaultConfig().)
// The Logger passed will be used to log errors and information to. If nil is passed, a default Logger is
// used by calling logrus.New().
// Note that no two servers should be active at the same time. Doing so anyway will result in unexpected
// behaviour.
func New(c *Config, log internal.Logger) *Server {
	if log == nil {
		log = logrus.New()
	}
	if c == nil {
		conf := DefaultConfig()
		c = &conf
	}
	s := &Server{
		c:              *c,
		log:            log,
		players:        make(chan *player.Player),
		world:          world.New(log, c.World.SimulationDistance),
		p:              make(map[uuid.UUID]*player.Player),
		name:           *atomic.NewString(c.Server.Name),
		playerProvider: player.NopProvider{},
	}
	s.JoinMessage(c.Server.JoinMessage)
	s.QuitMessage(c.Server.QuitMessage)

	s.checkNetIsolation()
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

	server.log.Infof("Starting Minecraft Bedrock Edition server for v%v...", protocol.CurrentVersion)
	server.loadWorld()
	server.registerTargetFunc()

	if err := server.startListening(); err != nil {
		return err
	}
	server.run()
	return nil
}

// Start runs the server but does not block, unlike Run, but instead accepts connections on a different
// goroutine. Connections will be accepted until the listener is closed using a call to Close.
// Once started, players may be accepted using Server.Accept().
func (server *Server) Start() error {
	if !server.started.CAS(false, true) {
		panic("server already running")
	}

	server.log.Infof("Starting Minecraft Bedrock Edition server for v%v...", protocol.CurrentVersion)
	server.loadWorld()
	server.registerTargetFunc()

	if err := server.startListening(); err != nil {
		return err
	}
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

// PlayerProvider changes the data provider of a player to the provider passed. The provider will dictate
// the behaviour of player saving and loading. If nil is passed, the NopProvider will be used
// which does not read or write any data.
func (server *Server) PlayerProvider(provider player.Provider) {
	if provider == nil {
		provider = player.NopProvider{}
	}
	server.playerProvider = provider
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

// JoinMessage changes the join message for all players on the server. Leave this empty to disable it.
// %v is the placeholder for the username of the player
func (server *Server) JoinMessage(message string) {
	server.joinMessage.Store(message)
}

// QuitMessage changes the leave message for all players on the server. Leave this empty to disable it.
// %v is the placeholder for the username of the player
func (server *Server) QuitMessage(message string) {
	server.quitMessage.Store(message)
}

// Close closes the server, making any call to Run/Accept cancel immediately.
func (server *Server) Close() error {
	if !server.running() {
		panic("server not yet running")
	}

	server.log.Infof("Server shutting down...")
	defer server.log.Infof("Server stopped.")

	server.log.Debugf("Disconnecting players...")
	server.playerMutex.RLock()
	for _, p := range server.p {
		p.Disconnect(text.Colourf("<yellow>%v</yellow>", server.c.Server.ShutdownMessage))
	}
	server.playerMutex.RUnlock()

	server.playerProvider.Close()

	server.log.Debugf("Closing world...")
	if err := server.world.Close(); err != nil {
		return err
	}

	server.log.Debugf("Closing listener...")
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

	cfg := minecraft.ListenConfig{
		MaximumPlayers:         server.c.Server.MaximumPlayers,
		StatusProvider:         statusProvider{s: server},
		AuthenticationDisabled: !server.c.Server.AuthEnabled,
	}

	var err error
	//noinspection SpellCheckingInspection
	server.listener, err = cfg.Listen("raknet", server.c.Network.Address)
	if err != nil {
		return fmt.Errorf("listening on address failed: %w", err)
	}

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
		PlayerPosition: vec64To32(server.world.Spawn().Vec3Centre().Add(mgl64.Vec3{0, 1.62})),
		PlayerGameMode: 1,
		// We set these IDs to 1, because that's how the session will treat them.
		EntityUniqueID:               1,
		EntityRuntimeID:              1,
		Time:                         int64(server.world.Time()),
		GameRules:                    []protocol.GameRule{{Name: "naturalregeneration", Value: false}},
		Difficulty:                   2,
		Items:                        server.itemEntries(),
		PlayerMovementSettings:       protocol.PlayerMovementSettings{MovementType: protocol.PlayerMovementModeServer, ServerAuthoritativeBlockBreaking: true},
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
		server.log.Debugf("connection %v has a malformed UUID ('%v')\n", conn.RemoteAddr(), id)
		return
	}
	if p, ok := server.Player(id); ok {
		p.Disconnect("Logged in from another location.")
	}
	server.players <- server.createPlayer(id, conn)
}

// checkNetIsolation checks if a loopback exempt is in place to allow the hosting device to join the server. This is
// only relevant on Windows. It will never log anything for anything but Windows.
func (server *Server) checkNetIsolation() {
	if runtime.GOOS != "windows" {
		// Only an issue on Windows.
		return
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	if bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe")) {
		return
	}
	const loopbackExemptCmd = `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`
	server.log.Infof("You are currently unable to join the server on this machine. Run %v in an admin PowerShell session to be able to.\n", loopbackExemptCmd)
}

// handleSessionClose handles the closing of a session. It removes the player of the session from the server.
func (server *Server) handleSessionClose(controllable session.Controllable) {
	server.playerMutex.Lock()
	delete(server.p, controllable.UUID())
	server.playerMutex.Unlock()
}

// createPlayer creates a new player instance using the UUID and connection passed.
func (server *Server) createPlayer(id uuid.UUID, conn *minecraft.Conn) *player.Player {
	s := session.New(conn, server.c.World.MaximumChunkRadius, server.log, &server.joinMessage, &server.quitMessage)
	p := player.NewWithSession(conn.IdentityData().DisplayName, conn.IdentityData().XUID, id, server.createSkin(conn.ClientData()), s, server.world.Spawn().Vec3Middle(), server.playerProvider)
	s.Start(p, server.world, server.handleSessionClose)

	return p
}

// loadWorld loads the world of the server, ending the program if the world could not be loaded.
func (server *Server) loadWorld() {
	server.log.Debugf("Loading world...")

	p, err := mcdb.New(server.c.World.Folder)
	if err != nil {
		server.log.Fatalf("error loading world: %v", err)
	}
	server.world.Provider(p)
	server.world.Generator(generator.Flat{})

	server.log.Debugf("Loaded world '%v'.", server.world.Name())
}

// createSkin creates a new skin using the skin data found in the client data in the login, and returns it.
func (server *Server) createSkin(data login.ClientData) skin.Skin {
	// gopher tunnel guarantees the following values are valid data and are of the correct size.
	skinData, _ := base64.StdEncoding.DecodeString(data.SkinData)
	capeData, _ := base64.StdEncoding.DecodeString(data.CapeData)
	modelData, _ := base64.StdEncoding.DecodeString(data.SkinGeometry)
	skinResourcePatch, _ := base64.StdEncoding.DecodeString(data.SkinResourcePatch)
	modelConfig, _ := skin.DecodeModelConfig(skinResourcePatch)

	playerSkin := skin.New(data.SkinImageWidth, data.SkinImageHeight)
	playerSkin.Persona = data.PersonaSkin
	playerSkin.Pix = skinData
	playerSkin.Model = modelData
	playerSkin.ModelConfig = modelConfig
	playerSkin.PlayFabID = data.PlayFabID

	playerSkin.Cape = skin.NewCape(data.CapeImageWidth, data.CapeImageHeight)
	playerSkin.Cape.Pix = capeData

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

		anim := skin.NewAnimation(animation.ImageWidth, animation.ImageHeight, animation.AnimationExpression, t)
		anim.FrameCount = int(animation.Frames)
		anim.Pix, _ = base64.StdEncoding.DecodeString(animation.Image)

		playerSkin.Animations = append(playerSkin.Animations, anim)
	}

	return playerSkin
}

// registerTargetFunc registers a cmd.TargetFunc to be able to get all players connected and all entities in
// the server's world.
func (server *Server) registerTargetFunc() {
	cmd.AddTargetFunc(func(src cmd.Source) ([]cmd.Target, []cmd.Target) {
		entities, players := src.World().Entities(), server.Players()
		eTargets, pTargets := make([]cmd.Target, len(entities)), make([]cmd.Target, len(players))

		entities = src.World().Entities()
		for i, e := range entities {
			eTargets[i] = e
		}
		for i, p := range players {
			pTargets[i] = p
		}
		return eTargets, pTargets
	})
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

// itemEntries loads a list of all custom item entries of the server, ready to be sent in the StartGame
// packet.
func (server *Server) itemEntries() (entries []protocol.ItemEntry) {
	for _, it := range world.Items() {
		name, _ := it.EncodeItem()
		rid, _, _ := world.ItemRuntimeID(it)

		entries = append(entries, protocol.ItemEntry{
			Name:      name,
			RuntimeID: int16(rid),
		})
	}
	return
}
