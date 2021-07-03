package player

import (
	"errors"
	"github.com/google/uuid"
	"io"
)

// Provider represents a value that may provide data to a Player value. It usually does the reading and
// writing of the player data so that the Player may use it.
type Provider interface {
	// Save is called when the player leaves the server and passes on the current player data.
	Save(UUID uuid.UUID, data Data) error
	// Load is called when the player joins and passes the UUID of the player.
	// It expects to the player data, and a bool that indicates whether or not the player has played before.
	// If this bool is false the player will use default values and you can use an empty Data struct.
	Load(UUID uuid.UUID) (Data, error)
	// Closer is used on server close when the server calls Provider.Close() and
	// is useful to safely close your database.
	io.Closer
}

// NopProvider is a player data provider that won't store any data and instead always return default values
type NopProvider struct{}

// Save ...
func (NopProvider) Save(uuid.UUID, Data) error {
	return nil
}

// Load ...
func (NopProvider) Load(uuid.UUID) (Data, error) {
	return Data{}, errors.New("player provider is not implemented")
}

// Close ...
func (NopProvider) Close() error {
	return nil
}
