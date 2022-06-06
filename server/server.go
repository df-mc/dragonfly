package server

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/internal/iteminternal"
	"github.com/df-mc/dragonfly/server/internal/packbuilder"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	_ "github.com/df-mc/dragonfly/server/item" // Imported for compiler directives.
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/playerdb"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	"github.com/df-mc/dragonfly/server/world/generator"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Server implements a Dragonfly server. It runs the main server loop and handles the connections of players
// trying to join the server.
type Server struct {
	c       Config
	log     Logger
	name    atomic.Value[string]
	started atomic.Bool

	playerProvider     atomic.Value[player.Provider]
	world, nether, end *world.World

	listenMu  sync.Mutex
	listeners []Listener
	a         atomic.Value[Allower]

	resources      []*resource.Pack
	itemComponents map[string]map[string]any

	joinMessage, quitMessage atomic.Value[string]

	incoming    chan *session.Session
	playerMutex sync.RWMutex
	// p holds a map of all players currently connected to the server. When they leave, they are removed from
	// the map.
	p map[uuid.UUID]*player.Player
	// pwg is a sync.WaitGroup used to wait for all players to be disconnected before server shutdown, so that their
	// data is saved properly.
	pwg sync.WaitGroup

	wg sync.WaitGroup
}

// Logger is used to report information and errors from a dragonfly Server. Any Logger implementation may be used by
// passing it to server.New.
type Logger interface {
	world.Logger
	session.Logger
	Infof(format string, v ...any)
	Fatalf(format string, v ...any)
}

func init() {
	// Seeding the random for things like lightning that need to use RNG.
	rand.Seed(time.Now().UnixNano())
}

// New returns a new server using the Config passed. If nil is passed, a default configuration is returned.
// (A call to server.DefaultConfig().)
// The Logger passed will be used to log errors and information to. If nil is passed, a default Logger is
// used by calling logrus.New().
// Note that no two servers should be active at the same time. Doing so anyway will result in unexpected
// behaviour.
func New(c *Config, log Logger) *Server {
	if log == nil {
		log = logrus.New()
	}
	log.Infof("Loading server...")
	if c == nil {
		conf := DefaultConfig()
		c = &conf
	}
	s := &Server{
		c:              *c,
		log:            log,
		incoming:       make(chan *session.Session),
		p:              make(map[uuid.UUID]*player.Player),
		name:           *atomic.NewValue(c.Server.Name),
		playerProvider: *atomic.NewValue[player.Provider](player.NopProvider{}),
		a:              *atomic.NewValue[Allower](allower{}),
		world:          &world.World{}, nether: &world.World{}, end: &world.World{},
	}
	p, err := mcdb.New(c.World.Folder, opt.FlateCompression)
	if err != nil {
		log.Fatalf("error loading world: %v", err)
	}
	s.world = s.createWorld(world.Overworld, &s.nether, &s.end, biome.Plains{}, []world.Block{block.Grass{}, block.Dirt{}, block.Dirt{}, block.Bedrock{}}, p)
	s.nether = s.createWorld(world.Nether, &s.world, &s.end, biome.NetherWastes{}, []world.Block{block.Netherrack{}, block.Netherrack{}, block.Netherrack{}, block.Bedrock{}}, p)
	s.end = s.createWorld(world.End, &s.nether, &s.world, biome.End{}, []world.Block{block.EndStone{}, block.EndStone{}, block.EndStone{}, block.Bedrock{}}, p)

	s.registerTargetFunc()

	s.JoinMessage(c.Server.JoinMessage)
	s.QuitMessage(c.Server.QuitMessage)

	s.loadResources(c.Resources.Folder, log)
	s.checkNetIsolation()

	if c.Players.SaveData {
		p, err := playerdb.NewProvider(c.Players.Folder)
		if err != nil {
			log.Fatalf("error loading player provider: %v", err)
		}
		s.PlayerProvider(p)
	}
	return s
}

// HandleFunc is a function that may be passed to Server.Accept(). It can be used to prepare the session of a
// player before it can do anything.
type HandleFunc func(p *player.Player)

// Accept accepts an incoming player into the server. It blocks until a player connects to the server. A HandleFunc may
// be passed which is run immediately before a *player.Player is accepted to the Server. This function may be used to
// add a player.Handler to the player and prepare its session. The function may be nil if player joining does not need
// to be handled.
// Accept returns false if the Server is closed using a call to Close.
func (srv *Server) Accept(f HandleFunc) bool {
	s, ok := <-srv.incoming
	if !ok {
		return false
	}
	p := s.Controllable().(*player.Player)
	if f != nil {
		f(p)
	}

	srv.playerMutex.Lock()
	srv.p[p.UUID()] = p
	srv.playerMutex.Unlock()

	s.Start()
	return true
}

// World returns the overworld of the server. Players will be spawned in this world and this world will be read
// from and written to when the world is edited.
func (srv *Server) World() *world.World {
	return srv.world
}

// Nether returns the nether world of the server. Players are transported to it when entering a nether portal in the
// world returned by the World method.
func (srv *Server) Nether() *world.World {
	return srv.nether
}

// End returns the end world of the server. Players are transported to it when entering an end portal in the world
// returned by the World method.
func (srv *Server) End() *world.World {
	return srv.end
}

// Start runs the server but does not block, unlike Run, but instead accepts connections on a different
// goroutine. Connections will be accepted until the listener is closed using a call to Close.
// Once started, players may be accepted using Server.Accept().
func (srv *Server) Start() error {
	if !srv.started.CAS(false, true) {
		panic("srv already running")
	}

	srv.log.Infof("Starting Dragonfly for Minecraft v%v...", protocol.CurrentVersion)
	if err := srv.startListening(); err != nil {
		return err
	}
	go srv.wait()
	return nil
}

// PlayerCount returns the current player count of the server. It is equivalent to calling
// len(server.Players()).
func (srv *Server) PlayerCount() int {
	srv.playerMutex.RLock()
	defer srv.playerMutex.RUnlock()

	return len(srv.p)
}

// MaxPlayerCount returns the maximum amount of players that are allowed to play on the server at the same
// time. Players trying to join when the server is full will be refused to enter.
// If the config has a maximum player count set to 0, MaxPlayerCount will return Server.PlayerCount + 1.
func (srv *Server) MaxPlayerCount() int {
	if srv.c.Players.MaxCount == 0 {
		return srv.PlayerCount() + 1
	}
	return srv.c.Players.MaxCount
}

// Players returns a list of all players currently connected to the server. Note that the slice returned is
// not updated when new players join or leave, so it is only valid for as long as no new players join or
// players leave.
func (srv *Server) Players() []*player.Player {
	srv.playerMutex.RLock()
	defer srv.playerMutex.RUnlock()
	return maps.Values(srv.p)
}

// Player looks for a player on the server with the UUID passed. If found, the player is returned and the bool
// returns holds a true value. If not, the bool returned is false and the player is nil.
func (srv *Server) Player(uuid uuid.UUID) (*player.Player, bool) {
	srv.playerMutex.RLock()
	defer srv.playerMutex.RUnlock()
	p, ok := srv.p[uuid]
	return p, ok
}

// PlayerByName looks for a player on the server with the name passed. If found, the player is returned and the bool
// returns holds a true value. If not, the bool is false and the player is nil
func (srv *Server) PlayerByName(name string) (*player.Player, bool) {
	for _, p := range srv.Players() {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

// PlayerProvider changes the data provider of a player to the provider passed. The provider will dictate
// the behaviour of player saving and loading. If nil is passed, the NopProvider will be used
// which does not read or write any data.
func (srv *Server) PlayerProvider(provider player.Provider) {
	if provider == nil {
		provider = player.NopProvider{}
	}
	srv.playerProvider.Store(provider)
}

// AddResourcePack loads a resource pack to the server. The pack will eventually be sent to clients who join the
// server when started.
func (srv *Server) AddResourcePack(pack *resource.Pack) {
	srv.resources = append(srv.resources, pack)
}

// Resources returns a list of all resource packs currently loaded on the server.
func (srv *Server) Resources() []*resource.Pack {
	return srv.resources
}

// SetName sets the name of the Server, also known as the MOTD. This name is displayed in the server list.
// The formatting of the name passed follows the rules of fmt.Sprint.
func (srv *Server) SetName(a ...any) {
	srv.name.Store(format(a))
}

// JoinMessage changes the join message for all players on the server. Leave this empty to disable it.
// %v is the placeholder for the username of the player
func (srv *Server) JoinMessage(a ...any) {
	srv.joinMessage.Store(format(a))
}

// QuitMessage changes the leave message for all players on the server. Leave this empty to disable it.
// %v is the placeholder for the username of the player
func (srv *Server) QuitMessage(a ...any) {
	srv.quitMessage.Store(format(a))
}

// Close closes the server, making any call to Run/Accept cancel immediately.
func (srv *Server) Close() error {
	if !srv.running() {
		panic("server not yet running")
	}

	srv.log.Infof("Server shutting down...")
	defer srv.log.Infof("Server stopped.")

	srv.log.Debugf("Disconnecting players...")
	srv.playerMutex.RLock()
	for _, p := range srv.p {
		p.Disconnect(text.Colourf("<yellow>%v</yellow>", srv.c.Server.ShutdownMessage))
	}
	srv.playerMutex.RUnlock()
	srv.pwg.Wait()

	srv.log.Debugf("Closing player provider...")
	if err := srv.playerProvider.Load().Close(); err != nil {
		srv.log.Errorf("Error while closing player provider: %v", err)
	}

	srv.log.Debugf("Closing worlds...")
	if err := srv.nether.Close(); err != nil {
		srv.log.Errorf("Error closing nether %v", err)
	}
	if err := srv.end.Close(); err != nil {
		srv.log.Errorf("Error closing end: %v", err)
	}
	if err := srv.world.Close(); err != nil {
		srv.log.Errorf("Error closing overworld: %v", err)
	}
	srv.log.Debugf("Closing listeners...")
	srv.listenMu.Lock()

	defer srv.listenMu.Unlock()
	for _, l := range srv.listeners {
		if err := l.Close(); err != nil {
			srv.log.Errorf("Error closing listener: %v", err)
		}
	}
	return nil
}

// Allow makes the Server filter which connections to the Server are accepted. Connections on which the Allower returns
// false are rejected immediately. If nil is passed, all connections are accepted.
func (srv *Server) Allow(a Allower) {
	if a == nil {
		a = allower{}
	}
	srv.a.Store(a)
}

// Listen makes the Server listen for new connections from the Listener passed. This may be used to listen for players
// on different interfaces. Note that the maximum player count of additional Listeners added is not enforced
// automatically. The limit must be enforced by the Listener.
func (srv *Server) Listen(l Listener) {
	srv.listenMu.Lock()
	srv.listeners = append(srv.listeners, l)
	srv.listenMu.Unlock()

	srv.wg.Add(1)

	wg := new(sync.WaitGroup)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		for {
			c, err := l.Accept()
			if err != nil {
				// Cancel the context so that any call to StartGameContext is cancelled rapidly.
				cancel()
				// First wait until all connections that are being handled are done inserting the player into the channel.
				// Afterwards, when we're sure no more values will be inserted in the players channel, we can return so the
				// player channel can be closed.
				wg.Wait()
				srv.wg.Done()
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				if msg, ok := srv.a.Load().Allow(c.RemoteAddr(), c.IdentityData(), c.ClientData()); !ok {
					_ = c.WritePacket(&packet.Disconnect{HideDisconnectionScreen: msg == "", Message: msg})
					_ = c.Close()
					return
				}
				srv.finaliseConn(ctx, c, l)
			}()
		}
	}()
}

// CloseOnProgramEnd closes the server right before the program ends, so that all data of the server are
// saved properly.
func (srv *Server) CloseOnProgramEnd() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		if err := srv.Close(); err != nil {
			srv.log.Errorf("error shutting down srv: %v", err)
		}
	}()
}

// running checks if the server is currently running.
func (srv *Server) running() bool {
	return srv.started.Load()
}

// startListening starts making the EncodeBlock listener listen, accepting new connections from players.
func (srv *Server) startListening() error {
	texturePacksRequired := srv.c.Resources.Required
	srv.makeItemComponents()
	if srv.c.Resources.AutoBuildPack {
		if pack, ok := packbuilder.BuildResourcePack(); ok {
			srv.resources = append(srv.resources, pack)
			texturePacksRequired = true
		}
	}

	cfg := minecraft.ListenConfig{
		MaximumPlayers:         srv.c.Players.MaxCount,
		StatusProvider:         statusProvider{s: srv},
		AuthenticationDisabled: !srv.c.Server.AuthEnabled,
		ResourcePacks:          srv.resources,
		Biomes:                 srv.biomes(),
		TexturePacksRequired:   texturePacksRequired,
	}

	l, err := cfg.Listen("raknet", srv.c.Network.Address)
	if err != nil {
		return fmt.Errorf("listening on address failed: %w", err)
	}
	srv.Listen(listener{Listener: l})

	srv.log.Infof("Server running on %v.\n", l.Addr())
	return nil
}

// makeItemComponents initializes the server's item components map using the registered custom items. It allows item
// components to be created only once at startup
func (srv *Server) makeItemComponents() {
	srv.itemComponents = make(map[string]map[string]any)
	for _, it := range world.CustomItems() {
		name, _ := it.EncodeItem()
		if data, ok := iteminternal.Components(it); ok {
			srv.itemComponents[name] = data
		}
	}
}

// wait awaits the closing of all Listeners added to the Server through a call to Listen and closed the players channel
// once that happens.
func (srv *Server) wait() {
	srv.wg.Wait()
	close(srv.incoming)
}

// finaliseConn finalises the session.Conn passed and subtracts from the sync.WaitGroup once done.
func (srv *Server) finaliseConn(ctx context.Context, conn session.Conn, l Listener) {
	id := uuid.MustParse(conn.IdentityData().Identity)
	data := srv.defaultGameData()

	var playerData *player.Data
	if d, err := srv.playerProvider.Load().Load(id); err == nil {
		data.PlayerPosition = vec64To32(d.Position).Add(mgl32.Vec3{0, 1.62})
		data.Yaw, data.Pitch = float32(d.Yaw), float32(d.Pitch)
		data.Dimension = int32(srv.dimension(d.Dimension).Dimension().EncodeDimension())

		playerData = &d
	}

	if err := conn.StartGameContext(ctx, data); err != nil {
		_ = l.Disconnect(conn, "Connection timeout.")
		srv.log.Debugf("connection %v failed spawning: %v\n", conn.RemoteAddr(), err)
		return
	}

	itemComponentEntries := make([]protocol.ItemComponentEntry, len(srv.itemComponents))
	for name, entry := range srv.itemComponents {
		itemComponentEntries = append(itemComponentEntries, protocol.ItemComponentEntry{
			Name: name,
			Data: entry,
		})
	}
	_ = conn.WritePacket(&packet.ItemComponent{Items: itemComponentEntries})
	if p, ok := srv.Player(id); ok {
		p.Disconnect("Logged in from another location.")
	}
	srv.incoming <- srv.createPlayer(id, conn, playerData)
}

// defaultGameData returns a minecraft.GameData as sent for a new player. It may later be modified if the player was
// saved in the player provide rof the server.
func (srv *Server) defaultGameData() minecraft.GameData {
	return minecraft.GameData{
		Yaw:            90,
		WorldName:      srv.world.Name(),
		PlayerPosition: vec64To32(srv.world.Spawn().Vec3Centre().Add(mgl64.Vec3{0, 1.62})),
		PlayerGameMode: 1,
		// We set these IDs to 1, because that's how the session will treat them.
		EntityUniqueID:               1,
		EntityRuntimeID:              1,
		Time:                         int64(srv.world.Time()),
		GameRules:                    []protocol.GameRule{{Name: "naturalregeneration", Value: false}},
		Difficulty:                   2,
		Items:                        srv.itemEntries(),
		PlayerMovementSettings:       protocol.PlayerMovementSettings{MovementType: protocol.PlayerMovementModeServer, ServerAuthoritativeBlockBreaking: true},
		ServerAuthoritativeInventory: true,
	}
}

// dimension returns a world by a dimension ID passed.
func (srv *Server) dimension(id int) *world.World {
	switch id {
	default:
		return srv.world
	case 1:
		return srv.nether
	case 2:
		return srv.end
	}
}

// checkNetIsolation checks if a loopback exempt is in place to allow the hosting device to join the server. This is
// only relevant on Windows. It will never log anything for anything but Windows.
func (srv *Server) checkNetIsolation() {
	if runtime.GOOS != "windows" {
		// Only an issue on Windows.
		return
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	if bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe")) {
		return
	}
	const loopbackExemptCmd = `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`
	srv.log.Infof("You are currently unable to join the server on this machine. Run %v in an admin PowerShell session to be able to.\n", loopbackExemptCmd)
}

// handleSessionClose handles the closing of a session. It removes the player of the session from the server.
func (srv *Server) handleSessionClose(c session.Controllable) {
	srv.playerMutex.Lock()
	p, ok := srv.p[c.UUID()]
	delete(srv.p, c.UUID())
	srv.playerMutex.Unlock()
	if !ok {
		// When a player disconnects immediately after a session is started, it might not be added to the players map
		// yet. This is expected, but we need to be careful not to crash when this happens.
		return
	}

	if err := srv.playerProvider.Load().Save(p.UUID(), p.Data()); err != nil {
		srv.log.Errorf("Error while saving data: %v", err)
	}
	srv.pwg.Done()
}

// createPlayer creates a new player instance using the UUID and connection passed.
func (srv *Server) createPlayer(id uuid.UUID, conn session.Conn, data *player.Data) *session.Session {
	w, gm, pos := srv.world, srv.world.DefaultGameMode(), srv.world.Spawn().Vec3Middle()
	if data != nil {
		w, gm, pos = srv.dimension(data.Dimension), data.GameMode, data.Position
	}
	s := session.New(conn, srv.c.Players.MaximumChunkRadius, srv.log, &srv.joinMessage, &srv.quitMessage)
	p := player.NewWithSession(conn.IdentityData().DisplayName, conn.IdentityData().XUID, id, srv.createSkin(conn.ClientData()), s, pos, data)

	s.Spawn(p, w, gm, srv.handleSessionClose)
	srv.pwg.Add(1)
	return s
}

// createWorld loads a world of the server with a specific dimension, ending the program if the world could not be loaded.
// The layers passed are used to create a generator.Flat that is used as generator for the world.
func (srv *Server) createWorld(d world.Dimension, nether, end **world.World, biome world.Biome, layers []world.Block, p world.Provider) *world.World {
	log := srv.log
	if v, ok := log.(interface {
		WithField(key string, field any) *logrus.Entry
	}); ok {
		// Add a dimension field to be able to distinguish between the different dimensions in the log. Dimensions
		// implement fmt.Stringer so we can just fmt.Sprint them for a readable name.
		log = v.WithField("dimension", strings.ToLower(fmt.Sprint(d)))
	}
	log.Debugf("Loading world...")

	conf := world.Config{
		Log:       log,
		Dim:       d,
		Provider:  p,
		Generator: generator.NewFlat(biome, layers),
		PortalDestination: func(dim world.Dimension) *world.World {
			if dim == world.Nether {
				return *nether
			} else if dim == world.End {
				return *end
			}
			return nil
		},
	}
	if f := srv.c.WorldConfig; f != nil {
		conf = f(conf)
	}
	w := conf.New()
	log.Infof(`Loaded world "%v".`, w.Name())
	return w
}

// createSkin creates a new skin using the skin data found in the client data in the login, and returns it.
func (srv *Server) createSkin(data login.ClientData) skin.Skin {
	// Gophertunnel guarantees the following values are valid data and are of the correct size.
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
func (srv *Server) registerTargetFunc() {
	cmd.AddTargetFunc(func(src cmd.Source) (entities, players []cmd.Target) {
		return sliceutil.Convert[cmd.Target](src.World().Entities()), sliceutil.Convert[cmd.Target](srv.Players())
	})
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

// itemEntries loads a list of all custom item entries of the server, ready to be sent in the StartGame
// packet.
func (srv *Server) itemEntries() (entries []protocol.ItemEntry) {
	for name, rid := range itemRuntimeIDs {
		entries = append(entries, protocol.ItemEntry{
			Name:      name,
			RuntimeID: int16(rid),
		})
	}
	for _, it := range world.CustomItems() {
		name, _ := it.EncodeItem()
		rid, _, _ := world.ItemRuntimeID(it)

		_, componentBased := srv.itemComponents[name]
		entries = append(entries, protocol.ItemEntry{
			Name:           name,
			RuntimeID:      int16(rid),
			ComponentBased: componentBased,
		})
	}
	return
}

// ashyBiome represents a biome that has any form of ash.
type ashyBiome interface {
	// Ash returns the ash and white ash of the biome.
	Ash() (ash float64, whiteAsh float64)
}

// sporingBiome represents a biome that has blue or red spores.
type sporingBiome interface {
	// Spores returns the blue and red spores of the biome.
	Spores() (blueSpores float64, redSpores float64)
}

// biomes builds a mapping of all biome definitions of the server, ready to be set in the biomes field of the
// server listener.
func (srv *Server) biomes() map[string]any {
	definitions := make(map[string]any)
	for _, b := range world.Biomes() {
		definition := map[string]any{
			"temperature": float32(b.Temperature()),
			"downfall":    float32(b.Rainfall()),
		}
		if a, ok := b.(ashyBiome); ok {
			ash, whiteAsh := a.Ash()
			definition["ash"], definition["white_ash"] = float32(ash), float32(whiteAsh)
		}
		if s, ok := b.(sporingBiome); ok {
			blueSpores, redSpores := s.Spores()
			definition["blue_spores"], definition["red_spores"] = float32(blueSpores), float32(redSpores)
		}
		definitions[b.String()] = definition
	}
	return definitions
}

// loadResources loads resource packs from path of specifed directory.
func (srv *Server) loadResources(p string, log Logger) {
	_ = os.Mkdir(p, 0777)
	resources, err := os.ReadDir(p)
	if err != nil {
		log.Fatalf("failed opening resource pack directory: %v\n", err)
	}
	for _, entry := range resources {
		pack, err := resource.Compile(filepath.Join(p, entry.Name()))
		if err != nil {
			log.Fatalf("Failed loading resource pack: %v", entry.Name())
		}
		srv.AddResourcePack(pack)
	}
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

var (
	//go:embed world/item_runtime_ids.nbt
	itemRuntimeIDData []byte
	itemRuntimeIDs    = map[string]int32{}
)

// init reads all item entries from the resource JSON, and sets the according values in the runtime ID maps.
func init() {
	_ = nbt.Unmarshal(itemRuntimeIDData, &itemRuntimeIDs)
}
