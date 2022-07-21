package server

import "github.com/df-mc/dragonfly/server/world"

// Config is the configuration of a Dragonfly server. It holds settings that affect different aspects of the
// server, such as its name and maximum players.
type Config struct {
	// Network holds settings related to network aspects of the server.
	Network struct {
		// Address is the address on which the server should listen. Players may connect to this address in
		// order to join.
		Address string
	}
	Server struct {
		// Name is the name of the server as it shows up in the server list.
		Name string
		// ShutdownMessage is the message shown to players when the server shuts down. If empty, players will
		// be directed to the menu screen right away.
		ShutdownMessage string
		// AuthEnabled controls whether players must be connected to Xbox Live in order to join the server.
		AuthEnabled bool
		// JoinMessage is the message that appears when a player joins the server. Leave this empty to disable it.
		// %v is the placeholder for the username of the player
		JoinMessage string
		// QuitMessage is the message that appears when a player leaves the server. Leave this empty to disable it.
		// %v is the placeholder for the username of the player
		QuitMessage string
	}
	World struct {
		// Name is the name of the world that the server holds. A world with this name will be loaded and
		// the name will be displayed at the top of the player list in the in-game pause menu.
		Name string
		// Folder is the folder that the data of the world resides in.
		Folder string
	}
	Players struct {
		// MaxCount is the maximum amount of players allowed to join the server at the same time. If set
		// to 0, the amount of maximum players will grow every time a player joins.
		MaxCount int
		// MaximumChunkRadius is the maximum chunk radius that players may set in their settings. If they try
		// to set it above this number, it will be capped and set to the max.
		MaximumChunkRadius int
		// SaveData controls whether a player's data will be saved and loaded. If true, the server will use the default
		// LevelDB data provider and if false, an empty provider will be used. To use your own provider, turn this value
		// to false as you will still be able to pass your own provider.
		SaveData bool
		// Folder controls where the player data will be stored by the default LevelDB
		// player provider if it is enabled.
		Folder string
	}
	Resources struct {
		// AutoBuildPack is if the server should automatically generate a resource pack for custom features.
		AutoBuildPack bool
		// Folder controls the location where resource packs will be loaded from.
		Folder string
		// Required is a boolean to force the client to load the resource pack on join. If they do not accept, they'll have to leave the server.
		Required bool
	}
	// WorldConfig, if non-nil, is called for every default world that a Server creates. It may be used to change the
	// world.Config that these worlds are created with. The default world.Config is passed to the function and should
	// be edited, then returned by the WorldConfig function.
	WorldConfig func(def world.Config) world.Config `toml:"-"`
}

// DefaultConfig returns a configuration with the default values filled out.
func DefaultConfig() Config {
	c := Config{}
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
