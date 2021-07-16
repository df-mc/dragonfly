package playerdb

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/player"
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
	db, err := leveldb.OpenFile(path, &opt.Options{
		Compression: opt.SnappyCompression,
	})
	if err != nil {
		return nil, err
	}
	return &Provider{db: db}, nil
}

// Save ...
func (p *Provider) Save(id uuid.UUID, d player.Data) error {
	data := toJson(d)
	jsondata, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_ = p.db.Put(id[:], jsondata, nil)
	return nil
}

// Load ...
func (p *Provider) Load(id uuid.UUID) (player.Data, error) {
	jsondata, err := p.db.Get(id[:], nil)
	if err != nil {
		return player.Data{}, err
	}
	d := jsonData{}
	err = json.Unmarshal(jsondata, &d)
	if err != nil {
		return player.Data{}, err
	}

	return fromJson(d), nil
}

// Close ...
func (p *Provider) Close() error {
	return p.db.Close()
}
