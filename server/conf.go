package server

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/playerdb"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	"github.com/df-mc/dragonfly/server/world/generator"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type Config struct {
	Log Logger

	Name string

	PlayerProvider player.Provider

	Listeners []Listener

	Resources []*resource.Pack

	ResourcesRequired bool

	DisableResourceBuilding bool

	Allower Allower

	AuthDisabled bool

	MaxPlayers int

	MaxChunkRadius int

	JoinMessage, QuitMessage, ShutdownMessage string

	WorldProvider world.Provider

	Generator func(dim world.Dimension) world.Generator

	RandomTickSpeed int
}

// Logger is used to report information and errors from a dragonfly Server. Any
// Logger implementation may be used by passing it to the Log field in Config.
type Logger interface {
	world.Logger
	session.Logger
	Infof(format string, v ...any)
	Fatalf(format string, v ...any)
}

func (conf Config) New() *Server {
	if len(conf.Listeners) == 0 {
		panic("config: at least 1 listener expected, 0 found")
	}
	if conf.Log == nil {
		conf.Log = logrus.New()
	}
	if conf.Name == "" {
		conf.Name = "Dragonfly Server"
	}
	if conf.PlayerProvider == nil {
		conf.PlayerProvider = player.NopProvider{}
	}
	if conf.Allower == nil {
		conf.Allower = allower{}
	}
	if conf.WorldProvider == nil {
		conf.WorldProvider = world.NopProvider{}
	}
	if conf.Generator == nil {
		conf.Generator = loadGenerator
	}
	if conf.MaxChunkRadius == 0 {
		conf.MaxChunkRadius = 12
	}

	srv := &Server{
		conf:     conf,
		incoming: make(chan *session.Session),
		p:        make(map[uuid.UUID]*player.Player),
		world:    &world.World{}, nether: &world.World{}, end: &world.World{},
	}
	srv.world = srv.createWorld(world.Overworld, &srv.nether, &srv.end)
	srv.nether = srv.createWorld(world.Nether, &srv.world, &srv.end)
	srv.end = srv.createWorld(world.End, &srv.nether, &srv.world)

	srv.registerTargetFunc()
	srv.checkNetIsolation()

	return srv
}

// UserConfig is the configuration of a Dragonfly server. It holds settings
// that affect different aspects of the server, such as its name and maximum
// players.
type UserConfig struct {
	// Network holds settings related to network aspects of the server.
	Network struct {
		// Address is the address on which the server should listen. Players may
		// connect to this address in order to join.
		Address string
	}
	Server struct {
		// Name is the name of the server as it shows up in the server list.
		Name string
		// ShutdownMessage is the message shown to players when the server shuts
		// down. If empty, players will be directed to the menu screen right
		// away.
		ShutdownMessage string
		// AuthEnabled controls whether players must be connected to Xbox Live
		// in order to join the server.
		AuthEnabled bool
		// JoinMessage is the message that appears when a player joins the
		// server. Leave this empty to disable it. %v is the placeholder for the
		// username of the player
		JoinMessage string
		// QuitMessage is the message that appears when a player leaves the
		// server. Leave this empty to disable it. %v is the placeholder for the
		// username of the player
		QuitMessage string
	}
	World struct {
		// Name is the name of the world that the server holds. A world with
		// this name will be loaded and the name will be displayed at the top of
		// the player list in the in-game pause menu.
		Name string
		// Folder is the folder that the data of the world resides in.
		Folder string
	}
	Players struct {
		// MaxCount is the maximum amount of players allowed to join the server
		// at the same time. If set to 0, the amount of maximum players will
		// grow every time a player joins.
		MaxCount int
		// MaximumChunkRadius is the maximum chunk radius that players may set
		// in their settings. If they try to set it above this number, it will
		// be capped and set to the max.
		MaximumChunkRadius int
		// SaveData controls whether a player's data will be saved and loaded.
		// If true, the server will use the default LevelDB data provider and if
		// false, an empty provider will be used. To use your own provider, turn
		// this value to false as you will still be able to pass your own
		// provider.
		SaveData bool
		// Folder controls where the player data will be stored by the default
		// LevelDB player provider if it is enabled.
		Folder string
	}
	Resources struct {
		// AutoBuildPack is if the server should automatically generate a
		// resource pack for custom features.
		AutoBuildPack bool
		// Folder controls the location where resource packs will be loaded
		// from.
		Folder string
		// Required is a boolean to force the client to load the resource pack
		// on join. If they do not accept, they'll have to leave the server.
		Required bool
	}
}

// Config converts a UserConfig to a Config, so that it may be used for creating
// a Server. An error is returned if creating data providers or loading
// resources failed.
func (uc UserConfig) Config(log Logger) (Config, error) {
	var err error
	conf := Config{
		Log:                     log,
		Name:                    uc.Server.Name,
		ResourcesRequired:       uc.Resources.Required,
		AuthDisabled:            !uc.Server.AuthEnabled,
		MaxPlayers:              uc.Players.MaxCount,
		MaxChunkRadius:          uc.Players.MaximumChunkRadius,
		JoinMessage:             uc.Server.JoinMessage,
		QuitMessage:             uc.Server.QuitMessage,
		ShutdownMessage:         uc.Server.ShutdownMessage,
		DisableResourceBuilding: !uc.Resources.AutoBuildPack,
	}
	conf.WorldProvider, err = mcdb.New(log, uc.World.Folder, opt.FlateCompression)
	if err != nil {
		return conf, fmt.Errorf("create world provider: %w", err)
	}
	conf.Resources, err = loadResources(uc.Resources.Folder)
	if err != nil {
		return conf, fmt.Errorf("load resources: %w", err)
	}
	if uc.Players.SaveData {
		conf.PlayerProvider, err = playerdb.NewProvider(uc.Players.Folder)
		if err != nil {
			return conf, fmt.Errorf("create player provider: %w", err)
		}
	}
	cfg := minecraft.ListenConfig{
		MaximumPlayers:         conf.MaxPlayers,
		StatusProvider:         statusProvider{name: conf.Name},
		AuthenticationDisabled: conf.AuthDisabled,
		ResourcePacks:          conf.Resources,
		Biomes:                 biomes(),
		TexturePacksRequired:   conf.ResourcesRequired,
	}
	var l *minecraft.Listener
	l, err = cfg.Listen("raknet", uc.Network.Address)
	if err != nil {
		return conf, fmt.Errorf("create minecraft listener: %w", err)
	}
	conf.Listeners = []Listener{listener{l}}
	conf.Log.Infof("Server running on %v.\n", l.Addr())

	return conf, nil
}

// loadResources loads all resource packs found in a directory passed.
func loadResources(dir string) ([]*resource.Pack, error) {
	_ = os.MkdirAll(dir, 0777)

	resources, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	packs := make([]*resource.Pack, len(resources))
	for i, entry := range resources {
		packs[i], err = resource.Compile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("compile resource (%v): %w", entry.Name(), err)
		}
	}
	return packs, nil
}

// loadGenerator loads a standard world.Generator for a world.Dimension.
func loadGenerator(dim world.Dimension) world.Generator {
	switch dim {
	case world.Overworld:
		return generator.NewFlat(biome.Plains{}, []world.Block{block.Grass{}, block.Dirt{}, block.Dirt{}, block.Bedrock{}})
	case world.Nether:
		return generator.NewFlat(biome.NetherWastes{}, []world.Block{block.Netherrack{}, block.Netherrack{}, block.Netherrack{}, block.Bedrock{}})
	case world.End:
		return generator.NewFlat(biome.End{}, []world.Block{block.EndStone{}, block.EndStone{}, block.EndStone{}, block.Bedrock{}})
	}
	panic("should never happen")
}

// DefaultConfig returns a configuration with the default values filled out.
func DefaultConfig() UserConfig {
	c := UserConfig{}
	c.Network.Address = ":19132"
	c.Server.Name = "Dragonfly Server"
	c.Server.ShutdownMessage = "Server closed."
	c.Server.AuthEnabled = true
	c.Server.JoinMessage = "%v has joined the game"
	c.Server.QuitMessage = "%v has left the game"
	c.World.Name = "World"
	c.World.Folder = "world"
	c.Players.MaximumChunkRadius = 32
	c.Players.SaveData = true
	c.Players.Folder = "players"
	c.Resources.AutoBuildPack = true
	c.Resources.Folder = "resources"
	c.Resources.Required = false
	return c
}
