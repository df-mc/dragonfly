package player

// Provider represents a value that may provide data to a Player value. It usually does the reading and
// writing of the player data so that the Player may use it.
type Provider interface {
	// Save is called when the player leaves the server and passes on the current player data.
	Save(data Data)
	// Load is called when the player joins and passes the XUID of the player.
	// It expects to the player data, and a bool that indicates whether or not the player has played before.
	// If this bool is false the player will use default values and you can use an empty Data struct.
	Load(XUID string) (Data, bool)
}

// NopProvider is a player data provider that won't store any data and instead always return default values
type NopProvider struct {}

// Save ...
func (NopProvider) Save(Data) {}

// Load ...
func (NopProvider) Load(string) (Data, bool) {
	return Data{}, false
}