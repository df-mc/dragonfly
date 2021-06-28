package playerdb

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"os"
)

// Provider ...
type Provider struct {
	db *leveldb.DB
}

// NewProvider creates a new player data provider that saves and loads data using
// a LevelDB database.
func NewProvider(path string) (*Provider, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0777)
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{db: db}, nil
}

// Save ...
func (p *Provider) Save(UUID uuid.UUID, d player.Data) error {
	data := toJson(d)
	jsondata, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_ = p.db.Put(d.UUID[:], jsondata, nil)
	return nil
}

// Load ...
func (p *Provider) Load(UUID uuid.UUID) (player.Data, error) {
	jsondata, err := p.db.Get(UUID[:], nil)
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
func (p *Provider) Close() {
	_ = p.db.Close()
}
