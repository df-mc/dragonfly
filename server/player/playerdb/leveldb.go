package playerdb

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/google/uuid"
	"os"
)

// Provider is a player data provider that uses a LevelDB database to store data. The data passed on
// will first be converted to make sure it can be marshaled into JSON. This JSON (in bytes) will then
// be stored in the database under a key that is the byte representation of the player's UUID.
type Provider struct {
	db *leveldb.DB
}

// NewProvider creates a new player data provider that saves and loads data using
// a LevelDB database.
func NewProvider(path string) (*Provider, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0777)
	}
	db, err := leveldb.OpenFile(path, &opt.Options{Compression: opt.SnappyCompression})
	if err != nil {
		return nil, err
	}
	return &Provider{db: db}, nil
}

// Save ...
func (p *Provider) Save(id uuid.UUID, d player.Config, w *world.World) error {
	b, err := json.Marshal(p.toJson(d, w))
	if err != nil {
		return err
	}
	return p.db.Put(id[:], b, nil)
}

// Load ...
func (p *Provider) Load(id uuid.UUID, world func(world.Dimension) *world.World) (player.Config, *world.World, error) {
	b, err := p.db.Get(id[:], nil)
	if err != nil {
		return player.Config{}, nil, err
	}
	var d jsonData
	err = json.Unmarshal(b, &d)
	if err != nil {
		return player.Config{}, nil, err
	}
	conf, w := p.fromJson(d, world)

	return conf, w, nil
}

// Close ...
func (p *Provider) Close() error {
	return p.db.Close()
}
