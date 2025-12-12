package server

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/internal/packbuilder"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/playerdb"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	"github.com/df-mc/dragonfly/server/world/generator"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// Config contains options for starting a Minecraft server.
type Config struct {
	// Log is the Logger to use for logging information. If nil, Log is set to
	// slog.Default(). Errors reported by the underlying network are only logged
	// if Log has at least debug level.
	Log *slog.Logger
	// Listeners is a list of functions to create a Listener using a Config, one
	// for each Listener to be added to the Server. If left empty, no players
	// will be able to connect to the Server.
	Listeners []func(conf Config) (Listener, error)
	// Name is the name of the server. By default, it is shown to users in the
	// server list before joining the server and when opening the in-game menu.
	Name string
	// Resources is a slice of resource packs to use on the server. When joining
	// the server, the player will then first be requested to download these
	// resource packs.
	Resources []*resource.Pack
	// ResourcesRequires specifies if the downloading of resource packs is
	// required to join the server. If set to true, players will not be able to
	// join without first downloading and applying the Resources above.
	ResourcesRequired bool
	// DisableResourceBuilding specifies if automatic resource pack building for
	// custom items should be disabled. Dragonfly, by default, automatically
	// produces a resource pack for custom items. If this is not desired (for
	// example if a resource pack already exists), this can be set to false.
	DisableResourceBuilding bool
	// Allower may be used to specify what players can join the server and what
	// players cannot. By returning false in the Allow method, for example if
	// the player has been banned, will prevent the player from joining.
	Allower Allower
	// AuthDisabled specifies if XBOX Live authentication should be disabled.
	// Note that this should generally only be done for testing purposes or for
	// local games. Allowing players to join without authentication is generally
	// a security hazard.
	AuthDisabled bool
	// MuteEmoteChat specifies if the player emote chat should be muted or not.
	MuteEmoteChat bool
	// MaxPlayers is the maximum amount of players allowed to join the server at
	// once.
	MaxPlayers int
	// MaxChunkRadius is the maximum view distance that each player may have,
	// measured in chunks. A chunk radius generally leads to more memory usage.
	MaxChunkRadius int
	// JoinMessage, QuitMessage and ShutdownMessage are the messages to send for
	// when a player joins or quits the server and when the server shuts down,
	// kicking all online players. If set, JoinMessage and QuitMessage must have
	// exactly 1 argument, which will be replaced with the name of the player
	// joining or quitting.
	// ShutdownMessage is set to chat.MessageServerDisconnect if empty.
	JoinMessage, QuitMessage, ShutdownMessage chat.Translation
	// StatusProvider provides the server status shown to players in the server
	// list. By default, StatusProvider will show the server name from the Name
	// field and the current player count and maximum players.
	StatusProvider minecraft.ServerStatusProvider
	// PlayerProvider is the player.Provider used for storing and loading player
	// data. If left as nil, player data will be newly created every time a
	// player joins the server and no data will be stored.
	PlayerProvider player.Provider
	// WorldProvider is the world.Provider used for storing and loading world
	// data. If left as nil, world data will be newly created every time and
	// chunks will always be newly generated when loaded. The world provider
	// will be used for storing/loading the default overworld, nether and end.
	WorldProvider world.Provider
	// ReadOnlyWorld specifies if the standard worlds should be read only. If
	// set to true, the WorldProvider won't be saved to at all.
	ReadOnlyWorld bool
	// Generator should return a function that specifies the world.Generator to
	// use for every world.Dimension (world.Overworld, world.Nether and
	// world.End). If left empty, Generator will be set to a flat world for each
	// of the dimensions (with netherrack and end stone for nether/end
	// respectively).
	Generator func(dim world.Dimension) world.Generator
	// RandomTickSpeed specifies the rate at which blocks should be ticked in
	// the default worlds. Setting this value to -1 or lower will stop random
	// ticking altogether, while setting it higher results in faster ticking. If
	// left as 0, the RandomTickSpeed will default to a speed of 3 blocks per
	// sub chunk per tick (normal ticking speed).
	RandomTickSpeed int
	// Entities is a world.EntityRegistry with all entity types registered that
	// may be added to the Server's worlds. If no entity types are registered,
	// Entities will be set to entity.DefaultRegistry.
	Entities world.EntityRegistry
}

// New creates a Server using fields of conf. The Server's worlds are created
// and connections from the Server's listeners may be accepted by calling
// Server.Listen() and Server.Accept() afterwards.
func (conf Config) New() *Server {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if len(conf.Listeners) == 0 {
		conf.Log.Warn("config: no listeners set, no connections will be accepted")
	}
	if conf.Name == "" {
		conf.Name = "Dragonfly Server"
	}
	if conf.StatusProvider == nil {
		conf.StatusProvider = statusProvider{name: conf.Name}
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
	if conf.ShutdownMessage.Zero() {
		conf.ShutdownMessage = chat.MessageServerDisconnect
	}
	if len(conf.Entities.Types()) == 0 {
		conf.Entities = entity.DefaultRegistry
	}
	if !conf.DisableResourceBuilding {
		if pack, ok := packbuilder.BuildResourcePack(); ok {
			conf.Resources = append(conf.Resources, pack)
		}
	}
	// Copy resources so that the slice can't be edited afterward.
	conf.Resources = slices.Clone(conf.Resources)

	srv := &Server{
		conf:     conf,
		incoming: make(chan incoming),
		p:        make(map[uuid.UUID]*onlinePlayer),
		world:    &world.World{}, nether: &world.World{}, end: &world.World{},
	}
	for _, lf := range conf.Listeners {
		l, err := lf(conf)
		if err != nil {
			conf.Log.Error("create listener: " + err.Error())
		}
		srv.listeners = append(srv.listeners, l)
	}

	creative_registerCreativeItems()
	world_finaliseBlockRegistry()
	recipe_registerVanilla()

	srv.world = srv.createWorld(world.Overworld, &srv.nether, &srv.end)
	srv.nether = srv.createWorld(world.Nether, &srv.world, &srv.end)
	srv.end = srv.createWorld(world.End, &srv.nether, &srv.world)

	return srv
}

// UserConfig is the user configuration for a Dragonfly server. It holds
// settings that affect different aspects of the server, such as its name and
// maximum players. UserConfig may be serialised and can be converted to a
// Config by calling UserConfig.Config().
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
		// AuthEnabled controls whether players must be connected to Xbox Live
		// in order to join the server.
		AuthEnabled bool
		// DisableJoinQuitMessages specifies if default join and quit messages
		// for players should be disabled.
		DisableJoinQuitMessages bool
		// MuteEmoteChat specifies if the player emote chat should be muted or not.
		MuteEmoteChat bool
	}
	World struct {
		// SaveData controls whether a world's data will be saved and loaded.
		// If true, the server will use the default LevelDB data provider and if
		// false, an empty provider will be used. To use your own provider, turn
		// this value to false, as you will still be able to pass your own
		// provider.
		SaveData bool
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
		// this value to false, as you will still be able to pass your own
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
func (uc UserConfig) Config(log *slog.Logger) (Config, error) {
	var err error
	conf := Config{
		Log:                     log,
		Name:                    uc.Server.Name,
		ResourcesRequired:       uc.Resources.Required,
		AuthDisabled:            !uc.Server.AuthEnabled,
		MuteEmoteChat:           uc.Server.MuteEmoteChat,
		MaxPlayers:              uc.Players.MaxCount,
		MaxChunkRadius:          uc.Players.MaximumChunkRadius,
		DisableResourceBuilding: !uc.Resources.AutoBuildPack,
	}
	if !uc.Server.DisableJoinQuitMessages {
		conf.JoinMessage, conf.QuitMessage = chat.MessageJoin, chat.MessageQuit
	}
	if uc.World.SaveData {
		conf.WorldProvider, err = mcdb.Config{Log: log}.Open(uc.World.Folder)
		if err != nil {
			return conf, fmt.Errorf("create world provider: %w", err)
		}
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
	conf.Listeners = append(conf.Listeners, uc.listenerFunc)
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
		packs[i], err = resource.ReadPath(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("compile resource (%v): %w", entry.Name(), err)
		}
	}
	return packs, nil
}

// loadGenerator loads a standard world.Generator for a world.Dimension. The
// generators returned are flat generators with grass/dirt, netherrack or end
// stone depending on the dimension passed.
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
	c.Server.AuthEnabled = true
	c.World.SaveData = true
	c.World.Folder = "world"
	c.Players.MaximumChunkRadius = 32
	c.Players.SaveData = true
	c.Players.Folder = "players"
	c.Resources.AutoBuildPack = true
	c.Resources.Folder = "resources"
	c.Resources.Required = false
	return c
}

// noinspection ALL
//
//go:linkname creative_registerCreativeItems github.com/df-mc/dragonfly/server/item/creative.registerCreativeItems
func creative_registerCreativeItems()

// noinspection ALL
//
//go:linkname recipe_registerVanilla github.com/df-mc/dragonfly/server/item/recipe.registerVanilla
func recipe_registerVanilla()

// noinspection ALL
//
//go:linkname world_finaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func world_finaliseBlockRegistry()
