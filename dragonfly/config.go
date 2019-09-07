package dragonfly

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
		// WorldName is the name of the world that the server holds. A world with this name will be loaded and
		// the name will be displayed at the top of the player list in the in-game pause menu.
		WorldName string
		// MaximumPlayers is the maximum amount of players allowed to join the server at the same time. If set
		// to 0, the amount of maximum players will grow every time a player joins.
		MaximumPlayers int
	}
}

// DefaultConfig returns a configuration with the default values filled out.
func DefaultConfig() Config {
	c := Config{}
	c.Network.Address = ":19132"
	c.Server.Name = "Dragonfly Server"
	c.Server.WorldName = "World"
	return c
}
