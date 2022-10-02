package server

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/internal/iteminternal"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	_ "github.com/df-mc/dragonfly/server/item" // Imported for maintaining correct initialisation order.
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Server implements a Dragonfly server. It runs the main server loop and
// handles the connections of players trying to join the server.
type Server struct {
	conf Config

	once    sync.Once
	started atomic.Bool

	world, nether, end *world.World

	customItems []protocol.ItemComponentEntry

	listeners []Listener
	incoming  chan *session.Session

	pmu sync.RWMutex
	// p holds a map of all players currently connected to the server. When they
	// leave, they are removed from the map.
	p map[uuid.UUID]*player.Player
	// pwg is a sync.WaitGroup used to wait for all players to be disconnected
	// before server shutdown, so that their data is saved properly.
	pwg sync.WaitGroup
	// wg is used to wait for all Listeners to be closed and their respective
	// goroutines to be finished.
	wg sync.WaitGroup
}

// HandleFunc is a function that may be passed to Server.Accept(). It can be
// used to prepare the session of a player before it can do anything.
type HandleFunc func(p *player.Player)

// New creates a Server using a default Config. The Server's worlds are created
// and connections from the Server's listeners may be accepted by calling
// Server.Listen() and Server.Accept() afterwards.
func New() *Server {
	var conf Config
	return conf.New()
}

// Listen starts running the server's listeners but does not block, unlike Run.
// Connections will be accepted on a different goroutine until the listeners
// are closed using a call to Close. Once started, players may be accepted
// using Server.Accept().
func (srv *Server) Listen() {
	if !srv.started.CAS(false, true) {
		panic("start server: already started")
	}

	srv.conf.Log.Infof("Starting Dragonfly for Minecraft v%v...", protocol.CurrentVersion)
	srv.startListening()
	go srv.wait()
}

// Accept accepts an incoming player into the server. It blocks until a player
// connects to the server. A HandleFunc may be passed which is run immediately
// before a *player.Player is accepted to the Server. This function may be used
// to add a player.Handler to the player and prepare its session. The function
// may be nil if player joining does not need to be handled. Accept returns
// false if the Server is closed using a call to Close.
func (srv *Server) Accept(f HandleFunc) bool {
	s, ok := <-srv.incoming
	if !ok {
		return false
	}
	p := s.Controllable().(*player.Player)
	if f != nil {
		f(p)
	}

	srv.pmu.Lock()
	srv.p[p.UUID()] = p
	srv.pmu.Unlock()

	s.Start()
	return true
}

// World returns the overworld of the server. Players will be spawned in this
// world and this world will be read from and written to when the world is
// edited.
func (srv *Server) World() *world.World {
	return srv.world
}

// Nether returns the nether world of the server. Players are transported to it
// when entering a nether portal in the world returned by the World method.
func (srv *Server) Nether() *world.World {
	return srv.nether
}

// End returns the end world of the server. Players are transported to it when
// entering an end portal in the world returned by the World method.
func (srv *Server) End() *world.World {
	return srv.end
}

// MaxPlayerCount returns the maximum amount of players that are allowed to
// play on the server at the same time. Players trying to join when the server
// is full will be refused to enter. If the config has a maximum player count
// set to 0, MaxPlayerCount will return Server.PlayerCount + 1.
func (srv *Server) MaxPlayerCount() int {
	if srv.conf.MaxPlayers == 0 {
		return len(srv.Players()) + 1
	}
	return srv.conf.MaxPlayers
}

// Players returns a list of all players currently connected to the server.
// Note that the slice returned is not updated when new players join or leave,
// so it is only valid for as long as no new players join or players leave.
func (srv *Server) Players() []*player.Player {
	srv.pmu.RLock()
	defer srv.pmu.RUnlock()
	return maps.Values(srv.p)
}

// Player looks for a player on the server with the UUID passed. If found, the
// player is returned and the bool returns holds a true value. If not, the bool
// returned is false and the player is nil.
func (srv *Server) Player(uuid uuid.UUID) (*player.Player, bool) {
	srv.pmu.RLock()
	defer srv.pmu.RUnlock()
	p, ok := srv.p[uuid]
	return p, ok
}

// PlayerByName looks for a player on the server with the name passed. If
// found, the player is returned and the bool returns holds a true value. If
// not, the bool is false and the player is nil
func (srv *Server) PlayerByName(name string) (*player.Player, bool) {
	return sliceutil.SearchValue(srv.Players(), func(p *player.Player) bool {
		return p.Name() == name
	})
}

// PlayerByXUID looks for a player on the server with the XUID passed. If found,
// the player is returned and the bool returned is true. If no player with the
// XUID was found, nil and false are returned.
func (srv *Server) PlayerByXUID(xuid string) (*player.Player, bool) {
	return sliceutil.SearchValue(srv.Players(), func(p *player.Player) bool {
		return p.XUID() == xuid
	})
}

// CloseOnProgramEnd closes the server right before the program ends, so that
// all data of the server are saved properly.
func (srv *Server) CloseOnProgramEnd() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		if err := srv.Close(); err != nil {
			srv.conf.Log.Errorf("close server: %v", err)
		}
	}()
}

// Close closes the server, making any call to Run/Accept cancel immediately.
func (srv *Server) Close() error {
	if !srv.started.Load() {
		panic("server not yet running")
	}
	srv.once.Do(srv.close)
	return nil
}

// close stops the server, storing player and world data to disk when
// necessary.
func (srv *Server) close() {
	srv.conf.Log.Infof("Server shutting down...")
	defer srv.conf.Log.Infof("Server stopped.")

	srv.conf.Log.Debugf("Disconnecting players...")
	for _, p := range srv.Players() {
		p.Disconnect(text.Colourf("<yellow>%v</yellow>", srv.conf.ShutdownMessage))
	}
	srv.pwg.Wait()

	srv.conf.Log.Debugf("Closing player provider...")
	if err := srv.conf.PlayerProvider.Close(); err != nil {
		srv.conf.Log.Errorf("Error while closing player provider: %v", err)
	}

	srv.conf.Log.Debugf("Closing worlds...")
	if err := srv.nether.Close(); err != nil {
		srv.conf.Log.Errorf("Error closing nether %v", err)
	}
	if err := srv.end.Close(); err != nil {
		srv.conf.Log.Errorf("Error closing end: %v", err)
	}
	if err := srv.world.Close(); err != nil {
		srv.conf.Log.Errorf("Error closing overworld: %v", err)
	}

	srv.conf.Log.Debugf("Closing listeners...")
	for _, l := range srv.listeners {
		if err := l.Close(); err != nil {
			srv.conf.Log.Errorf("Error closing listener: %v", err)
		}
	}
}

// listen makes the Server listen for new connections from the Listener passed.
// This may be used to listen for players on different interfaces. Note that
// the maximum player count of additional Listeners added is not enforced
// automatically. The limit must be enforced by the Listener.
func (srv *Server) listen(l Listener) {
	srv.wg.Add(1)

	wg := new(sync.WaitGroup)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		for {
			c, err := l.Accept()
			if err != nil {
				// Cancel the context so that any call to StartGameContext is
				// cancelled rapidly.
				cancel()
				// First wait until all connections that are being handled are
				// done inserting the player into the channel. Afterwards, when
				// we're sure no more values will be inserted in the players
				// channel, we can return so the player channel can be closed.
				wg.Wait()
				srv.wg.Done()
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				if msg, ok := srv.conf.Allower.Allow(c.RemoteAddr(), c.IdentityData(), c.ClientData()); !ok {
					_ = c.WritePacket(&packet.Disconnect{HideDisconnectionScreen: msg == "", Message: msg})
					_ = c.Close()
					return
				}
				srv.finaliseConn(ctx, c, l)
			}()
		}
	}()
}

// startListening starts making the EncodeBlock listener listen, accepting new
// connections from players.
func (srv *Server) startListening() {
	srv.makeItemComponents()

	for _, lf := range srv.conf.Listeners {
		l, err := lf(srv.conf)
		if err != nil {
			srv.conf.Log.Fatalf("create listener: %v", err)
		}
		srv.listeners = append(srv.listeners, l)
		srv.listen(l)
	}
}

// makeItemComponents initializes the server's item components map using the
// registered custom items. It allows item components to be created only once
// at startup
func (srv *Server) makeItemComponents() {
	custom := world.CustomItems()
	srv.customItems = make([]protocol.ItemComponentEntry, len(custom))

	for _, it := range custom {
		name, _ := it.EncodeItem()
		srv.customItems = append(srv.customItems, protocol.ItemComponentEntry{
			Name: name,
			Data: iteminternal.Components(it),
		})
	}
}

// wait awaits the closing of all Listeners added to the Server through a call
// to listen and closed the players channel once that happens.
func (srv *Server) wait() {
	srv.wg.Wait()
	close(srv.incoming)
}

// finaliseConn finalises the session.Conn passed and subtracts from the
// sync.WaitGroup once done.
func (srv *Server) finaliseConn(ctx context.Context, conn session.Conn, l Listener) {
	id := uuid.MustParse(conn.IdentityData().Identity)
	data := srv.defaultGameData()

	var playerData *player.Data
	if d, err := srv.conf.PlayerProvider.Load(id); err == nil {
		data.PlayerPosition = vec64To32(d.Position).Add(mgl32.Vec3{0, 1.62})
		data.Yaw, data.Pitch = float32(d.Yaw), float32(d.Pitch)
		data.Dimension = int32(srv.dimension(d.Dimension).Dimension().EncodeDimension())

		playerData = &d
	}

	if err := conn.StartGameContext(ctx, data); err != nil {
		_ = l.Disconnect(conn, "Connection timeout.")
		srv.conf.Log.Debugf("connection %v failed spawning: %v\n", conn.RemoteAddr(), err)
		return
	}
	_ = conn.WritePacket(&packet.ItemComponent{Items: srv.customItems})
	if p, ok := srv.Player(id); ok {
		p.Disconnect("Logged in from another location.")
	}
	srv.incoming <- srv.createPlayer(id, conn, playerData)
}

// defaultGameData returns a minecraft.GameData as sent for a new player. It
// may later be modified if the player was saved in the player provide rof the
// server.
func (srv *Server) defaultGameData() minecraft.GameData {
	return minecraft.GameData{
		// We set these IDs to 1, because that's how the session will treat them.
		EntityUniqueID:               1,
		EntityRuntimeID:              1,
		Difficulty:                   2,
		Yaw:                          90,
		ServerAuthoritativeInventory: true,
		WorldName:                    srv.conf.Name,
		Items:                        srv.itemEntries(),
		Time:                         int64(srv.world.Time()),
		PlayerGameMode:               packet.GameTypeCreative,
		PlayerPermissions:            packet.PermissionLevelMember,
		GameRules:                    []protocol.GameRule{{Name: "naturalregeneration", Value: false}},
		PlayerPosition:               vec64To32(srv.world.Spawn().Vec3Centre().Add(mgl64.Vec3{0, 1.62})),
		PlayerMovementSettings:       protocol.PlayerMovementSettings{MovementType: protocol.PlayerMovementModeServer, ServerAuthoritativeBlockBreaking: true},
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

// checkNetIsolation checks if a loopback exempt is in place to allow the
// hosting device to join the server. This is only relevant on Windows. It will
// never log anything for anything but Windows.
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
	srv.conf.Log.Infof("You are currently unable to join the server on this machine. Run %v in an admin PowerShell session to resolve.\n", loopbackExemptCmd)
}

// handleSessionClose handles the closing of a session. It removes the player
// of the session from the server.
func (srv *Server) handleSessionClose(c session.Controllable) {
	srv.pmu.Lock()
	p, ok := srv.p[c.UUID()]
	delete(srv.p, c.UUID())
	srv.pmu.Unlock()
	if !ok {
		// When a player disconnects immediately after a session is started, it might not be added to the players map
		// yet. This is expected, but we need to be careful not to crash when this happens.
		return
	}

	if err := srv.conf.PlayerProvider.Save(p.UUID(), p.Data()); err != nil {
		srv.conf.Log.Errorf("Error while saving data: %v", err)
	}
	srv.pwg.Done()
}

// createPlayer creates a new player instance using the UUID and connection
// passed.
func (srv *Server) createPlayer(id uuid.UUID, conn session.Conn, data *player.Data) *session.Session {
	w, gm, pos := srv.world, srv.world.DefaultGameMode(), srv.world.Spawn().Vec3Middle()
	if data != nil {
		w, gm, pos = srv.dimension(data.Dimension), data.GameMode, data.Position
	}
	s := session.New(conn, srv.conf.MaxChunkRadius, srv.conf.Log, srv.conf.JoinMessage, srv.conf.QuitMessage)
	p := player.NewWithSession(conn.IdentityData().DisplayName, conn.IdentityData().XUID, id, srv.parseSkin(conn.ClientData()), s, pos, data)

	s.Spawn(p, pos, w, gm, srv.handleSessionClose)
	srv.pwg.Add(1)
	return s
}

// createWorld loads a world of the server with a specific dimension, ending
// the program if the world could not be loaded. The layers passed are used to
// create a generator.Flat that is used as generator for the world.
func (srv *Server) createWorld(dim world.Dimension, nether, end **world.World) *world.World {
	logger := srv.conf.Log
	if v, ok := logger.(interface {
		WithField(key string, field any) *logrus.Entry
	}); ok {
		// Add a dimension field to be able to distinguish between the different
		// dimensions in the log. Dimensions implement fmt.Stringer so we can
		// just fmt.Sprint them for a readable name.
		logger = v.WithField("dimension", strings.ToLower(fmt.Sprint(dim)))
	}
	logger.Debugf("Loading world...")

	conf := world.Config{
		Log:             logger,
		Dim:             dim,
		Provider:        srv.conf.WorldProvider,
		Generator:       srv.conf.Generator(dim),
		RandomTickSpeed: srv.conf.RandomTickSpeed,
		PortalDestination: func(dim world.Dimension) *world.World {
			if dim == world.Nether {
				return *nether
			} else if dim == world.End {
				return *end
			}
			return nil
		},
	}
	w := conf.New()
	logger.Infof(`Opened world "%v".`, w.Name())
	return w
}

// parseSkin parses a skin from the login.ClientData  and returns it.
func (srv *Server) parseSkin(data login.ClientData) skin.Skin {
	// Gophertunnel guarantees the following values are valid data and are of
	// the correct size.
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

// registerTargetFunc registers a cmd.TargetFunc to be able to get all players
// connected and all entities in the server's world.
func (srv *Server) registerTargetFunc() {
	cmd.AddTargetFunc(func(src cmd.Source) (entities, players []cmd.Target) {
		return sliceutil.Convert[cmd.Target](src.World().Entities()), sliceutil.Convert[cmd.Target](srv.Players())
	})
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

// itemEntries loads a list of all custom item entries of the server, ready to
// be sent in the StartGame packet.
func (srv *Server) itemEntries() []protocol.ItemEntry {
	entries := make([]protocol.ItemEntry, 0, len(itemRuntimeIDs))

	for name, rid := range itemRuntimeIDs {
		entries = append(entries, protocol.ItemEntry{
			Name:      name,
			RuntimeID: int16(rid),
		})
	}
	for _, it := range world.CustomItems() {
		name, _ := it.EncodeItem()
		rid, _, _ := world.ItemRuntimeID(it)
		entries = append(entries, protocol.ItemEntry{
			Name:           name,
			ComponentBased: true,
			RuntimeID:      int16(rid),
		})
	}
	return entries
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

// biomes builds a mapping of all biome definitions of the server, ready to be set in the biomes field of the server
// listener.
func biomes() map[string]any {
	definitions := make(map[string]any)
	for _, b := range world.Biomes() {
		definition := map[string]any{
			"name_hash":   b.String(), // This isn't actually a hash despite what the field name may suggest.
			"temperature": float32(b.Temperature()),
			"downfall":    float32(b.Rainfall()),
			"rain":        b.Rainfall() > 0,
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

var (
	//go:embed world/item_runtime_ids.nbt
	itemRuntimeIDData []byte
	itemRuntimeIDs    = map[string]int32{}
)

// init reads all item entries from the resource JSON, and sets the according
// values in the runtime ID maps. init also seeds the global `rand` with the
// current time.
func init() {
	// Seeding the random for things like lightning that need to use RNG.
	rand.Seed(time.Now().UnixNano())

	_ = nbt.Unmarshal(itemRuntimeIDData, &itemRuntimeIDs)
}
